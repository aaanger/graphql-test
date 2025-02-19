package post

import (
	"context"
	"database/sql"
	"fmt"
	model2 "github.com/aaanger/graphql-test/internal/graph/model"
	"strings"
)

//go:generate mockery --name=IPostRepository

type IPostRepository interface {
	CreatePost(ctx context.Context, userID int, req *model2.CreatePostReq) (*model2.Post, error)
	GetAllPostsByUserID(ctx context.Context, userID int) ([]*model2.Post, error)
	GetPostByID(ctx context.Context, id int) (*model2.Post, error)
	UpdatePost(ctx context.Context, userID, postID int, req *model2.UpdatePostReq) error
	DeletePost(ctx context.Context, userID, postID int) error
}

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{
		db: db,
	}
}

func (r *PostRepository) CreatePost(ctx context.Context, userID int, req *model2.CreatePostReq) (*model2.Post, error) {
	post := model2.Post{
		Title:         req.Title,
		Body:          req.Body,
		AllowComments: req.AllowComments,
	}

	row := r.db.QueryRowContext(ctx, `INSERT INTO posts (user_id, title, body, created_at, allow_comments) VALUES($1, $2, $3, current_timestamp, $4) RETURNING id;`,
		userID, req.Title, req.Body, req.AllowComments)

	err := row.Scan(&post.ID)
	if err != nil {
		return nil, err
	}

	user := model2.User{
		ID: userID,
	}
	userRow := r.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id = $1;`, userID)
	err = userRow.Scan(&user.Username)

	post.User = &user

	return &post, nil
}

func (r *PostRepository) GetAllPostsByUserID(ctx context.Context, userID int) ([]*model2.Post, error) {
	var posts []*model2.Post

	rows, err := r.db.QueryContext(ctx, `SELECT (p.id, p.title, p.body, p.allow_comments, p.created_at, u.id, u.username) 
												FROM posts p INNER JOIN users u ON p.user_id = u.id 
												WHERE p.user_id = $1 ORDER BY created_at DESC;`,
		userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var post model2.Post
		var user model2.User
		err = rows.Scan(&post.ID, &post.Title, &post.Body, &post.AllowComments, &post.CreatedAt, &user.ID, &user.Username)
		if err != nil {
			return nil, err
		}

		post.User = &user
		posts = append(posts, &post)
	}

	return posts, nil
}

func (r *PostRepository) GetPostByID(ctx context.Context, id int) (*model2.Post, error) {
	var post model2.Post
	var user model2.User

	row := r.db.QueryRowContext(ctx, `SELECT (p.id, p.title, p.body, p.created_at, p.allow_comments, u.id, u.username) 
											FROM posts p INNER JOIN users u ON p.user_id = u.id 
											WHERE p.id = $1;`, id)

	err := row.Scan(&post.ID, &post.Title, &post.Body, &post.AllowComments, &post.CreatedAt, &user.ID, &user.Username)
	if err != nil {
		return nil, err
	}

	post.User = &user

	return &post, nil
}

func (r *PostRepository) UpdatePost(ctx context.Context, userID, postID int, req *model2.UpdatePostReq) error {
	keys := make([]string, 0)
	values := make([]interface{}, 0)

	arg := 1

	if req.Title != nil {
		keys = append(keys, fmt.Sprintf(`title=$%d`, arg))
		values = append(values, *req.Title)
		arg++
	}

	if req.Body != nil {
		keys = append(keys, fmt.Sprintf(`description=$%d`, arg))
		values = append(values, *req.Body)
		arg++
	}

	if req.AllowComments != nil {
		keys = append(keys, fmt.Sprintf("allow_comments=$%d", arg))
		values = append(values, *req.AllowComments)
		arg++
	}

	joinQuery := strings.Join(keys, ", ")

	query := fmt.Sprintf(`UPDATE posts SET %s WHERE id=$%d AND user_id=$%d;`, joinQuery, arg, arg+1)
	values = append(values, postID, userID)

	_, err := r.db.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostRepository) DeletePost(ctx context.Context, userID, postID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE user_id = $1 AND id = $2;`, userID, postID)
	if err != nil {
		return err
	}

	return nil
}
