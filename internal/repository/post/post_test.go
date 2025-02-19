package post

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	model2 "github.com/aaanger/graphql-test/internal/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type PostRepositorySuite struct {
	suite.Suite
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo *PostRepository
}

func (suite *PostRepositorySuite) SetupTest() {
	var err error
	suite.db, suite.mock, err = sqlmock.New()
	assert.NoError(suite.T(), err)
	suite.repo = NewPostRepository(suite.db)
}

func TestPostRepositorySuite(t *testing.T) {
	suite.Run(t, new(PostRepositorySuite))
}

// CreatePost
// ==============================================

func (suite *PostRepositorySuite) TestRepository_CreatePostSuccess() {
	req := &model2.CreatePostReq{
		Title:         "test",
		Body:          "test",
		AllowComments: true,
	}
	userID := 1

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	suite.mock.ExpectQuery("INSERT INTO posts").WithArgs(userID, req.Title, req.Body, req.AllowComments).
		WillReturnRows(rows)

	userRows := sqlmock.NewRows([]string{"username"}).AddRow("user")
	suite.mock.ExpectQuery("SELECT username FROM users WHERE (.+)").
		WithArgs(userID).
		WillReturnRows(userRows)

	post, err := suite.repo.CreatePost(context.Background(), userID, req)

	suite.NotNil(post)
	suite.Nil(err)
}

func (suite *PostRepositorySuite) TestRepository_CreatePostEmptyFields() {
	req := &model2.CreatePostReq{
		AllowComments: true,
	}
	userID := 1

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	suite.mock.ExpectQuery("INSERT INTO posts").WithArgs(userID, req.AllowComments).
		WillReturnRows(rows)

	post, err := suite.repo.CreatePost(context.Background(), userID, req)

	suite.Nil(post)
	suite.NotNil(err)
}

// GetAllPosts
// =======================================================================

func (suite *PostRepositorySuite) TestRepository_GetAllPostsSuccess() {
	userID := 1

	rows := sqlmock.NewRows([]string{"id", "title", "body", "allow_comments", "created_at", "id", "username"}).
		AddRow(1, "1", "1", true, time.Now(), 1, "user").AddRow(2, "2", "2", false, time.Now(), 2, "user2")
	suite.mock.ExpectQuery(`SELECT (.+) FROM posts p INNER JOIN users u ON (.+) WHERE (.+);`).
		WithArgs(userID).WillReturnRows(rows)

	posts, err := suite.repo.GetAllPostsByUserID(context.Background(), userID)

	expected := []*model2.Post{
		{
			ID: 1,
			User: &model2.User{
				ID:       1,
				Username: "user",
			},
			Title:         "1",
			Body:          "1",
			CreatedAt:     time.Now(),
			AllowComments: true,
		},
		{
			ID: 2,
			User: &model2.User{
				ID:       2,
				Username: "user2",
			},
			Title:         "2",
			Body:          "2",
			CreatedAt:     time.Now(),
			AllowComments: false,
		},
	}

	suite.NotNil(posts)
	suite.Nil(err)
	suite.Equal(expected, posts)
}

func (suite *PostRepositorySuite) TestRepository_GetAllPostsFailure() {
	userID := 1

	suite.mock.ExpectQuery(`SELECT (.+) FROM posts p INNER JOIN users u ON (.+) WHERE (.+);`).
		WithArgs(userID).WillReturnError(sql.ErrNoRows)

	posts, err := suite.repo.GetAllPostsByUserID(context.Background(), userID)

	suite.Nil(posts)
	suite.NotNil(err)
}

// GetPostByID
// =======================================================================

func (suite *PostRepositorySuite) TestRepository_GetPostByIDSuccess() {
	rows := sqlmock.NewRows([]string{"id", "title", "body", "allow_comments", "created_at", "id", "username"}).
		AddRow(1, "1", "1", true, time.Now(), 1, "user")
	suite.mock.ExpectQuery("SELECT (.+) FROM posts p INNER JOIN users u ON (.+) WHERE (.+);").WithArgs(1).WillReturnRows(rows)

	post, err := suite.repo.GetPostByID(context.Background(), 1)

	suite.NotNil(post)
	suite.Nil(err)
}

func (suite *PostRepositorySuite) TestRepository_GetPostByIDFailure() {
	suite.mock.ExpectQuery("SELECT (.+) FROM posts p INNER JOIN users u ON (.+) WHERE (.+);").
		WithArgs(1).WillReturnError(sql.ErrNoRows)

	post, err := suite.repo.GetPostByID(context.Background(), 1)

	suite.Nil(post)
	suite.NotNil(err)
}

// UpdatePost
// ======================================================================

func (suite *PostRepositorySuite) TestRepository_UpdatePostSuccess() {
	req := &model2.UpdatePostReq{
		Title:         strPointer("test"),
		Body:          strPointer("test"),
		AllowComments: boolPointer(true),
	}

	suite.mock.ExpectExec("UPDATE posts SET (.+) WHERE (.+)").
		WithArgs("test", "test", true, 1, 1).WillReturnResult(sqlmock.NewResult(0, 1))

	err := suite.repo.UpdatePost(context.Background(), 1, 1, req)

	suite.Nil(err)
}

func (suite *PostRepositorySuite) TestRepository_UpdatePostWithoutSomeFields() {
	req := &model2.UpdatePostReq{
		AllowComments: boolPointer(true),
	}

	suite.mock.ExpectExec("UPDATE posts SET (.+) WHERE (.+)").
		WithArgs(true, 1, 1).WillReturnResult(sqlmock.NewResult(0, 1))

	err := suite.repo.UpdatePost(context.Background(), 1, 1, req)

	suite.Nil(err)
}

func (suite *PostRepositorySuite) TestRepository_UpdatePostWithoutAllFields() {
	req := &model2.UpdatePostReq{}

	suite.mock.ExpectExec("UPDATE posts SET WHERE (.+)").
		WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))

	err := suite.repo.UpdatePost(context.Background(), 1, 1, req)

	suite.Nil(err)
}

// DeletePost
// ====================================================================================

func (suite *PostRepositorySuite) TestRepository_DeletePostSuccess() {
	suite.mock.ExpectExec("DELETE FROM posts WHERE (.+)").
		WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))

	err := suite.repo.DeletePost(context.Background(), 1, 1)

	suite.Nil(err)
}

func (suite *PostRepositorySuite) TestRepository_DeletePostFailure() {
	suite.mock.ExpectExec("DELETE FROM posts WHERE (.+)").
		WithArgs(1, 1).WillReturnError(sql.ErrNoRows)

	err := suite.repo.DeletePost(context.Background(), 1, 1)

	suite.NotNil(err)
}

func strPointer(s string) *string {
	return &s
}

func boolPointer(b bool) *bool {
	return &b
}
