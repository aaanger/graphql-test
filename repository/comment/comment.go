package comment

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aaanger/graphql-test/graph/model"
	"strings"
	"time"
)

//go:generate mockery --name=ICommentRepository

type ICommentRepository interface {
	CreateComment(ctx context.Context, userID int, req *model.CreateCommentReq) (*model.Comment, error)
	GetCommentByID(ctx context.Context, id int) (*model.Comment, error)
	GetCommentsByPostID(ctx context.Context, postID int, first, last *int, after, before *string) (*model.CommentConnection, error)
	UpdateComment(ctx context.Context, userID int, req *model.UpdateCommentReq) error
	DeleteComment(ctx context.Context, userID, commentID int) error
	IsCommentsAllowed(ctx context.Context, postID int) (bool, error)
}

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{
		db: db,
	}
}

func (r *CommentRepository) CreateComment(ctx context.Context, userID int, req *model.CreateCommentReq) (*model.Comment, error) {
	comment := model.Comment{
		UserID:          userID,
		PostID:          req.PostID,
		ParentCommentID: req.ParentCommentID,
		Body:            req.Body,
	}

	row := r.db.QueryRowContext(ctx, `INSERT INTO comments (post_id, user_id, parent_comment_id, body) VALUES($1, $2, $3, $4) RETURNING id, created_at;`,
		comment.PostID, comment.UserID, comment.ParentCommentID, comment.Body)

	err := row.Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

func (r *CommentRepository) GetCommentByID(ctx context.Context, id int) (*model.Comment, error) {
	var comment model.Comment

	row := r.db.QueryRowContext(ctx, `SELECT id, post_id, user_id, parent_comment_id, body, created_at FROM comments WHERE id = $1;`, id)

	err := row.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.ParentCommentID, &comment.Body, &comment.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

func (r *CommentRepository) GetCommentsByPostID(ctx context.Context, postID int, first, last *int, after, before *string) (*model.CommentConnection, error) {
	query := `WITH RECURSIVE comment_tree AS (
				SELECT id, post_id, user_id, parent_comment_id, body, created_at
				FROM comments
				WHERE post_id = $1
				UNION ALL
				SELECT c.id, c.post_id, c.user_id, c.parent_comment_id, c.body, c.created_at
				FROM comments c
				INNER JOIN comment_tree ct ON c.parent_comment_id = ct.id
				)
				SELECT id, post_id, user_id, parent_comment_id, body, created_at FROM comment_tree`

	keys := make([]string, 0)
	values := []interface{}{postID}
	arg := 2

	if after != nil {
		keys = append(keys, fmt.Sprintf("created_at > $%d", arg))
		parsedCursor, err := time.Parse(time.RFC3339, *after)
		if err != nil {
			return nil, err
		}
		values = append(values, parsedCursor)
		arg++
	}

	if before != nil {
		keys = append(keys, fmt.Sprintf("created_at < $%d", arg))
		parsedCursor, err := time.Parse(time.RFC3339, *before)
		if err != nil {
			return nil, err
		}
		values = append(values, parsedCursor)
		arg++
	}

	if len(keys) > 0 {
		query += " WHERE " + strings.Join(keys, " AND ")
	}

	if first != nil {
		query += fmt.Sprintf(" ORDER BY created_at ASC LIMIT $%d", arg)
		values = append(values, *first)
		arg++
	} else if last != nil {
		query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", arg)
		values = append(values, *last)
		arg++
	}

	rows, err := r.db.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var comments []*model.CommentEdge
	var startCursor, endCursor *string
	var count int
	var hasNextPage bool
	var hasPrevPage bool

	commentMap := make(map[int]*model.Comment)

	for rows.Next() {
		var comment model.Comment

		err = rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.ParentCommentID, &comment.Body, &comment.CreatedAt)
		if err != nil {
			return nil, err
		}

		cursorStr := comment.CreatedAt.Format(time.RFC3339)
		if count == 0 {
			startCursor = &cursorStr
		}
		if (first != nil && count == *first) || (last != nil && count == *last) {
			if first != nil {
				hasNextPage = true
			} else {
				hasPrevPage = true
			}
			break
		}
		comments = append(comments, &model.CommentEdge{
			Cursor: cursorStr,
			Node:   &comment,
		})

		commentMap[comment.ID] = &comment

		if comment.ParentCommentID != nil {
			parentComment := commentMap[*comment.ParentCommentID]
			if parentComment != nil {
				parentComment.Replies = append(parentComment.Replies, &comment)
			}
		}
		endCursor = &cursorStr
		count++
	}

	if last != nil {
		for i, j := 0, len(comments)-1; i < j; i, j = i+1, j-1 {
			comments[i], comments[j] = comments[j], comments[i]
		}
	}

	return &model.CommentConnection{
		Edges: comments,
		PageInfo: &model.PageInfo{
			StartCursor: startCursor,
			EndCursor:   endCursor,
			HasNextPage: first != nil && hasNextPage,
			HasPrevPage: last != nil && hasPrevPage,
		},
	}, nil
}

func (r *CommentRepository) UpdateComment(ctx context.Context, userID int, req *model.UpdateCommentReq) error {
	_, err := r.db.ExecContext(ctx, `UPDATE comments SET body = $1 WHERE user_id = $2 AND id = $3;`, req.Body, userID, req.ID)
	if err != nil {
		return err
	}

	return err
}

func (r *CommentRepository) DeleteComment(ctx context.Context, userID, commentID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM comments WHERE user_id = $1 AND id = $2;`, userID, commentID)
	if err != nil {
		return err
	}

	return nil
}

func (r *CommentRepository) IsCommentsAllowed(ctx context.Context, postID int) (bool, error) {
	var allowComments bool

	row := r.db.QueryRowContext(ctx, `SELECT allow_comments FROM posts WHERE id = $1;`, postID)

	err := row.Scan(&allowComments)
	if err != nil {
		return false, err
	}

	return allowComments, nil
}
