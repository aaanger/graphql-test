package comment

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aaanger/graphql-test/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type CommentRepositorySuite struct {
	suite.Suite
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo *CommentRepository
}

func (suite *CommentRepositorySuite) SetupTest() {
	var err error
	suite.db, suite.mock, err = sqlmock.New()
	assert.NoError(suite.T(), err)
	suite.repo = NewCommentRepository(suite.db)
}

func TestCommentRepositorySuite(t *testing.T) {
	suite.Run(t, new(CommentRepositorySuite))
}

// CreateComment
// =====================================================================

func (suite *CommentRepositorySuite) TestRepository_CreateCommentSuccess() {
	req := &model.CreateCommentReq{
		PostID: 1,
		Body:   "test",
	}

	rows := sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, time.Now())
	suite.mock.ExpectQuery("INSERT INTO comments").
		WithArgs(1, 1, nil, "test").WillReturnRows(rows)

	comment, err := suite.repo.CreateComment(context.Background(), 1, req)

	suite.NotNil(comment)
	suite.Nil(err)
}

func (suite *CommentRepositorySuite) TestRepository_CreateCommentEmptyFields() {
	req := &model.CreateCommentReq{
		PostID: 1,
	}

	rows := sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, time.Now()).
		RowError(0, errors.New("error"))
	suite.mock.ExpectQuery("INSERT INTO comments").
		WithArgs(1, 1, nil, "test").WillReturnRows(rows)

	comment, err := suite.repo.CreateComment(context.Background(), 1, req)

	suite.Nil(comment)
	suite.NotNil(err)
}

// GetComments
// ================================================================

func (suite *CommentRepositorySuite) TestRepository_GetCommentsByPostIDSuccess() {
	rows := sqlmock.NewRows([]string{"id", "post_id", "user_id", "parent_comment_id", "body", "created_at"}).
		AddRow(1, 1, 1, nil, "test1", time.Now()).
		AddRow(2, 1, 2, 1, "reply", time.Now().Add(time.Minute))

	first := 2
	suite.mock.ExpectQuery("WITH RECURSIVE comment_tree").
		WithArgs(1, &first).WillReturnRows(rows)

	comments, err := suite.repo.GetCommentsByPostID(context.Background(), 1, &first, nil, nil, nil)

	suite.NotNil(comments)
	suite.Nil(err)
}

func (suite *CommentRepositorySuite) TestRepository_GetCommentsByPostIDWithCursorsSuccess() {
	rows := sqlmock.NewRows([]string{"id", "post_id", "user_id", "parent_comment_id", "body", "created_at"}).
		AddRow(1, 1, 1, nil, "test1", time.Now()).
		AddRow(2, 1, 2, 1, "reply", time.Now().Add(time.Minute)).
		AddRow(3, 1, 3, nil, "third comment", time.Now().Add(time.Hour)).
		AddRow(4, 1, 3, 1, "second reply", time.Now().Add(3*time.Hour))

	first := 2
	after := time.Now().Format(time.RFC3339)
	before := time.Now().Add(2 * time.Hour).Format(time.RFC3339)

	afterTime, _ := time.Parse(time.RFC3339, after)
	beforeTime, _ := time.Parse(time.RFC3339, before)

	suite.mock.ExpectQuery("WITH RECURSIVE comment_tree").
		WithArgs(1, afterTime, beforeTime, first).WillReturnRows(rows)

	comments, err := suite.repo.GetCommentsByPostID(context.Background(), 1, &first, nil, &after, &before)

	suite.NotNil(comments)
	suite.Nil(err)
}

func (suite *CommentRepositorySuite) TestRepository_GetCommentsByPostIDFailure() {
	first := 2
	after := time.Now().Format(time.RFC3339)
	before := time.Now().Add(2 + time.Hour).Format(time.RFC3339)

	afterTime, _ := time.Parse(time.RFC3339, after)
	beforeTime, _ := time.Parse(time.RFC3339, before)

	suite.mock.ExpectQuery("WITH RECURSIVE comment_tree").
		WithArgs(1, afterTime, beforeTime, first).WillReturnError(sql.ErrNoRows)

	comments, err := suite.repo.GetCommentsByPostID(context.Background(), 1, &first, nil, &after, &before)

	suite.Nil(comments)
	suite.NotNil(err)
}

// GetCommentByID
// ================================================================

func (suite *CommentRepositorySuite) TestRepository_GetCommentByIDSuccess() {
	rows := sqlmock.NewRows([]string{"id", "post_id", "user_id", "body", "created_at", "parent_comment_id"}).
		AddRow(1, 1, 1, nil, "test", time.Now())
	suite.mock.ExpectQuery("SELECT (.+) FROM comments WHERE (.+)").
		WithArgs(1).WillReturnRows(rows)

	comment, err := suite.repo.GetCommentByID(context.Background(), 1)

	suite.NotNil(comment)
	suite.Nil(err)
}

func (suite *CommentRepositorySuite) TestRepository_GetCommentByIDFailure() {
	suite.mock.ExpectQuery("SELECT (.+) FROM comments WHERE (.+)").
		WithArgs(1).WillReturnError(sql.ErrNoRows)

	comment, err := suite.repo.GetCommentByID(context.Background(), 1)

	suite.Nil(comment)
	suite.NotNil(err)
}

// UpdateComment
// ================================================================

func (suite *CommentRepositorySuite) TestRepository_UpdateCommentSuccess() {
	req := &model.UpdateCommentReq{
		ID:   1,
		Body: "test",
	}

	suite.mock.ExpectExec("UPDATE comments SET (.+) WHERE (.+)").
		WithArgs("test", 1, 1).WillReturnResult(sqlmock.NewResult(0, 1))

	err := suite.repo.UpdateComment(context.Background(), 1, req)

	suite.Nil(err)
}

func (suite *CommentRepositorySuite) TestRepository_UpdateCommentEmptyBody() {
	req := &model.UpdateCommentReq{
		ID: 1,
	}

	suite.mock.ExpectExec("UPDATE comments SET (.+) WHERE (.+)").
		WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))

	err := suite.repo.UpdateComment(context.Background(), 1, req)

	suite.NotNil(err)
}

// DeleteComment
// ========================================================================================

func (suite *CommentRepositorySuite) TestRepository_DeleteCommentSuccess() {
	suite.mock.ExpectExec("DELETE FROM comments WHERE (.+)").
		WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))

	err := suite.repo.DeleteComment(context.Background(), 1, 1)

	suite.Nil(err)
}

func (suite *CommentRepositorySuite) TestRepository_DeleteCommentFailure() {
	suite.mock.ExpectExec("DELETE FROM comments WHERE (.+)").
		WithArgs(1, 1).WillReturnError(sql.ErrNoRows)

	err := suite.repo.DeleteComment(context.Background(), 1, 1)

	suite.NotNil(err)
}
