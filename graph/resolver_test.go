package graph

import (
	"context"
	"errors"
	"fmt"
	"github.com/aaanger/graphql-test/graph/model"
	commentMocks "github.com/aaanger/graphql-test/repository/comment/mocks"
	postMocks "github.com/aaanger/graphql-test/repository/post/mocks"
	userMocks "github.com/aaanger/graphql-test/repository/user/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"strings"
	"testing"
	"time"
)

type SchemaResolverSuite struct {
	suite.Suite
	userMock         *userMocks.IUserRepository
	postMock         *postMocks.IPostRepository
	commentMock      *commentMocks.ICommentRepository
	mutationResolver MutationResolver
	queryResolver    QueryResolver
}

func (suite *SchemaResolverSuite) SetupTest() {
	suite.userMock = userMocks.NewIUserRepository(suite.T())
	suite.postMock = postMocks.NewIPostRepository(suite.T())
	suite.commentMock = commentMocks.NewICommentRepository(suite.T())

	suite.mutationResolver = &mutationResolver{
		Resolver: &Resolver{
			UserRepo:    suite.userMock,
			PostRepo:    suite.postMock,
			CommentRepo: suite.commentMock,
		},
	}

	suite.queryResolver = &queryResolver{
		Resolver: &Resolver{
			UserRepo:    suite.userMock,
			PostRepo:    suite.postMock,
			CommentRepo: suite.commentMock,
		},
	}
}

func TestSchemaResolverSuite(t *testing.T) {
	suite.Run(t, new(SchemaResolverSuite))
}

// ==================================================================

func (suite *SchemaResolverSuite) TestResolver_RegisterSuccess() {
	req := model.RegisterReq{
		Email:    "test",
		Username: "test",
		Password: "test",
	}

	suite.userMock.On("Register", mock.Anything, &req).Return(&model.User{
		ID:       1,
		Email:    "test",
		Username: "test",
		Password: "test",
	}, "token", nil)

	res, err := suite.mutationResolver.Register(context.Background(), req)

	suite.NotNil(res.User)
	suite.NotEmpty(res.Token)
	suite.Nil(err)
}

func (suite *SchemaResolverSuite) TestResolver_RegisterFailure() {
	req := model.RegisterReq{
		Email:    "test",
		Username: "test",
		Password: "test",
	}

	suite.userMock.On("Register", mock.Anything, &req).Return(nil, "", errors.New("error"))

	res, err := suite.mutationResolver.Register(context.Background(), req)

	suite.Nil(res)
	suite.NotNil(err)
}

func (suite *SchemaResolverSuite) TestResolver_LoginSuccess() {
	req := model.LoginReq{
		Email:    "test",
		Password: "test",
	}

	suite.userMock.On("Login", mock.Anything, &req).Return(&model.User{
		ID:       1,
		Email:    "test",
		Username: "test",
		Password: "test",
	}, "token", nil)

	res, err := suite.mutationResolver.Login(context.Background(), req)

	suite.NotNil(res.User)
	suite.NotEmpty(res.Token)
	suite.Nil(err)
}

func (suite *SchemaResolverSuite) TestResolver_LoginFailure() {
	req := model.LoginReq{
		Email:    "test",
		Password: "test",
	}

	suite.userMock.On("Login", mock.Anything, &req).Return(nil, "", errors.New("error"))

	res, err := suite.mutationResolver.Login(context.Background(), req)

	suite.Nil(res)
	suite.NotNil(err)
}

// ===============================================================

func (suite *SchemaResolverSuite) TestResolver_CreatePostSuccess() {
	req := model.CreatePostReq{
		Title:         "test",
		Body:          "test",
		AllowComments: true,
	}

	ctx := context.WithValue(context.Background(), "userID", 1)

	suite.postMock.On("CreatePost", ctx, 1, &req).
		Return(&model.Post{
			ID: 1,
			User: &model.User{
				ID: 1,
			},
			Title:         req.Title,
			Body:          req.Body,
			AllowComments: req.AllowComments,
			CreatedAt:     time.Now(),
		}, nil)

	res, err := suite.mutationResolver.CreatePost(ctx, req)

	suite.NotNil(res)
	suite.Nil(err)
}

func (suite *SchemaResolverSuite) TestResolver_CreatePostUnauthorized() {
	req := model.CreatePostReq{
		Title:         "test",
		Body:          "test",
		AllowComments: true,
	}

	res, err := suite.mutationResolver.CreatePost(context.Background(), req)

	suite.Nil(res)
	suite.NotNil(err)
}

func (suite *SchemaResolverSuite) TestResolver_CreatePostFailure() {
	req := model.CreatePostReq{
		Title:         "test",
		Body:          "test",
		AllowComments: true,
	}

	ctx := context.WithValue(context.Background(), "userID", 1)

	suite.postMock.On("CreatePost", ctx, 1, &req).
		Return(nil, errors.New("error"))

	res, err := suite.mutationResolver.CreatePost(ctx, req)

	suite.Nil(res)
	suite.NotNil(err)
}

// ==================================================================

func (suite *SchemaResolverSuite) TestResolver_UpdatePostSuccess() {
	req := model.UpdatePostReq{
		Title:         strPointer("test"),
		Body:          strPointer("test"),
		AllowComments: boolPointer(true),
	}

	ctx := context.WithValue(context.Background(), "userID", 1)

	suite.postMock.On("UpdatePost", ctx, 1, 1, &req).
		Return(nil)

	suite.postMock.On("GetPostByID", ctx, 1).
		Return(&model.Post{
			ID: 1,
			User: &model.User{
				ID: 1,
			},
			Title:         "test",
			Body:          "test",
			AllowComments: true,
			CreatedAt:     time.Now(),
		}, nil)

	res, err := suite.mutationResolver.UpdatePost(ctx, 1, req)

	suite.NotNil(res)
	suite.Nil(err)
}

func (suite *SchemaResolverSuite) TestResolver_UpdatePostUnauthorized() {
	req := model.UpdatePostReq{
		Title:         strPointer("test"),
		Body:          strPointer("test"),
		AllowComments: boolPointer(true),
	}

	res, err := suite.mutationResolver.UpdatePost(context.Background(), 1, req)

	suite.Nil(res)
	suite.NotNil(err)
}

func (suite *SchemaResolverSuite) TestResolver_UpdatePostFailure() {
	req := model.UpdatePostReq{
		Title:         strPointer("test"),
		Body:          strPointer("test"),
		AllowComments: boolPointer(true),
	}

	ctx := context.WithValue(context.Background(), "userID", 1)

	suite.postMock.On("UpdatePost", ctx, 1, 1, &req).
		Return(errors.New("error"))

	res, err := suite.mutationResolver.UpdatePost(ctx, 1, req)

	suite.Nil(res)
	suite.NotNil(err)
}

// ========================================================

func (suite *SchemaResolverSuite) TestResolver_DeletePostSuccess() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	suite.postMock.On("DeletePost", ctx, 1, 1).Return(nil)

	status, err := suite.mutationResolver.DeletePost(ctx, 1)

	suite.Equal("Post deleted successfully", status)
	suite.Nil(err)
}

func (suite *SchemaResolverSuite) TestResolver_DeletePostUnauthorized() {
	status, err := suite.mutationResolver.DeletePost(context.Background(), 1)

	suite.Equal("Unauthorized", status)
	suite.NotNil(err)
}

func (suite *SchemaResolverSuite) TestResolver_DeletePostFailure() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	suite.postMock.On("DeletePost", ctx, 1, 1).Return(errors.New("error"))

	status, err := suite.mutationResolver.DeletePost(ctx, 1)

	suite.Equal("Failed to delete post", status)
	suite.NotNil(err)
}

// ==============================================================

func (suite *SchemaResolverSuite) TestResolver_GetPostsByUserIDSuccess() {
	post1 := &model.Post{
		ID:            1,
		Title:         "test1",
		Body:          "test1",
		AllowComments: true,
		User: &model.User{
			ID:       5,
			Username: "test",
		},
	}

	post2 := &model.Post{
		ID:            1,
		Title:         "test2",
		Body:          "test2",
		AllowComments: true,
		User: &model.User{
			ID:       5,
			Username: "test",
		},
	}

	suite.postMock.On("GetAllPostsByUserID", mock.Anything, 5).
		Return([]*model.Post{post1, post2}, nil)

	posts, err := suite.queryResolver.GetPosts(context.Background(), 5)

	suite.NotNil(posts)

	suite.Equal("test1", posts[0].Title)
	suite.Equal("test1", posts[0].Body)
	suite.Equal(5, posts[0].User.ID)
	suite.Equal("test", posts[0].User.Username)

	suite.Equal("test2", posts[1].Title)
	suite.Equal("test2", posts[1].Body)
	suite.Equal(5, posts[1].User.ID)
	suite.Equal("test", posts[1].User.Username)

	suite.Nil(err)
}

func (suite *SchemaResolverSuite) TestResolver_GetPostsByUserIDFailure() {
	suite.postMock.On("GetAllPostsByUserID", mock.Anything, 5).
		Return(nil, errors.New("error"))

	posts, err := suite.queryResolver.GetPosts(context.Background(), 5)

	suite.Nil(posts)
	suite.NotNil(err)
}

// ==============================================================

func (suite *SchemaResolverSuite) TestResolver_GetPostByIDSuccess() {
	suite.postMock.On("GetPostByID", mock.Anything, 1).
		Return(&model.Post{
			ID:            1,
			Title:         "test1",
			Body:          "test1",
			AllowComments: true,
			User: &model.User{
				ID:       5,
				Username: "test",
			},
		}, nil)

	post, err := suite.queryResolver.GetPostByID(context.Background(), 1)

	suite.NotNil(post)

	suite.Equal(1, post.ID)
	suite.Equal("test1", post.Title)
	suite.Equal("test1", post.Body)
	suite.Equal(true, post.AllowComments)

	suite.Nil(err)
}

func (suite *SchemaResolverSuite) TestResolver_GetPostByIDFailure() {
	suite.postMock.On("GetPostByID", mock.Anything, 1).
		Return(nil, errors.New("error"))

	post, err := suite.queryResolver.GetPostByID(context.Background(), 1)

	suite.Nil(post)
	suite.NotNil(err)
}

// Comments
// ==============================================================

func (suite *SchemaResolverSuite) TestResolver_CreateCommentSuccess() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	req := model.CreateCommentReq{
		PostID:          1,
		ParentCommentID: nil,
		Body:            "test",
	}

	suite.commentMock.On("IsCommentsAllowed", ctx, 1).Return(true, nil)
	suite.commentMock.On("CreateComment", ctx, 1, &req).
		Return(&model.Comment{
			ID:        1,
			PostID:    1,
			UserID:    1,
			Body:      "test",
			CreatedAt: time.Now(),
		}, nil)

	comment, err := suite.mutationResolver.CreateComment(ctx, req)

	suite.NotNil(comment)
	suite.Nil(err)
}

func (suite *SchemaResolverSuite) TestResolver_CreateCommentUnauthorized() {
	req := model.CreateCommentReq{
		PostID:          1,
		ParentCommentID: nil,
		Body:            "test",
	}

	comment, err := suite.mutationResolver.CreateComment(context.Background(), req)

	suite.Nil(comment)
	suite.NotNil(err)
}

func (suite *SchemaResolverSuite) TestResolver_CreateCommentNotAllowed() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	req := model.CreateCommentReq{
		PostID:          1,
		ParentCommentID: nil,
		Body:            "test",
	}

	suite.commentMock.On("IsCommentsAllowed", ctx, 1).Return(false, nil)

	comment, err := suite.mutationResolver.CreateComment(ctx, req)

	suite.Nil(comment)
	suite.Equal("comments are not allowed for this post", err.Error())
}

func (suite *SchemaResolverSuite) TestResolver_CreateCommentMoreThan2000Chars() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	req := model.CreateCommentReq{
		PostID:          1,
		ParentCommentID: nil,
		Body:            generateStringWith2000Chars(),
	}

	suite.commentMock.On("IsCommentsAllowed", ctx, 1).Return(true, nil)

	comment, err := suite.mutationResolver.CreateComment(ctx, req)

	suite.Nil(comment)
	suite.Equal("comment must be less than 2000 chars", err.Error())
}

func (suite *SchemaResolverSuite) TestResolver_CreateCommentFailure() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	req := model.CreateCommentReq{
		PostID:          1,
		ParentCommentID: nil,
		Body:            "test",
	}

	suite.commentMock.On("IsCommentsAllowed", ctx, 1).Return(true, nil)
	suite.commentMock.On("CreateComment", ctx, 1, &req).
		Return(nil, errors.New("error"))

	comment, err := suite.mutationResolver.CreateComment(ctx, req)

	suite.Nil(comment)
	suite.NotNil(err)
}

// ====================================================================

func (suite *SchemaResolverSuite) TestResolver_UpdateCommentSuccess() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	req := model.UpdateCommentReq{
		ID:   1,
		Body: "test",
	}

	suite.commentMock.On("UpdateComment", ctx, 1, &req).Return(nil)

	suite.commentMock.On("GetCommentByID", ctx, 1).
		Return(&model.Comment{
			ID:        1,
			PostID:    1,
			UserID:    1,
			Body:      "test",
			CreatedAt: time.Now(),
		}, nil)

	comment, err := suite.mutationResolver.UpdateComment(ctx, req)

	suite.NotNil(comment)
	suite.Nil(err)
}

func (suite *SchemaResolverSuite) TestResolver_UpdateCommentUnauthorized() {
	req := model.UpdateCommentReq{
		ID:   1,
		Body: "test",
	}

	comment, err := suite.mutationResolver.UpdateComment(context.Background(), req)

	suite.Nil(comment)
	suite.NotNil(err)
}

func (suite *SchemaResolverSuite) TestResolver_UpdateCommentMoreThan2000Chars() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	req := model.UpdateCommentReq{
		ID:   1,
		Body: generateStringWith2000Chars(),
	}

	comment, err := suite.mutationResolver.UpdateComment(ctx, req)

	suite.Nil(comment)
	suite.Equal("comment must be less than 2000 chars", err.Error())
}

func (suite *SchemaResolverSuite) TestResolver_UpdateCommentFailure() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	req := model.UpdateCommentReq{
		ID:   1,
		Body: "test",
	}

	suite.commentMock.On("UpdateComment", ctx, 1, &req).Return(errors.New("error"))

	comment, err := suite.mutationResolver.UpdateComment(ctx, req)

	suite.Nil(comment)
	suite.NotNil(err)
}

func (suite *SchemaResolverSuite) TestResolver_UpdateCommentGetCommentFailure() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	req := model.UpdateCommentReq{
		ID:   1,
		Body: "test",
	}

	suite.commentMock.On("UpdateComment", ctx, 1, &req).Return(nil)

	suite.commentMock.On("GetCommentByID", ctx, 1).
		Return(nil, errors.New("error"))

	comment, err := suite.mutationResolver.UpdateComment(ctx, req)

	suite.Nil(comment)
	suite.NotNil(err)
}

// =================================================================

func (suite *SchemaResolverSuite) TestResolver_DeleteCommentSuccess() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	suite.commentMock.On("DeleteComment", ctx, 1, 1).Return(nil)

	status, err := suite.mutationResolver.DeleteComment(ctx, 1)

	suite.Equal("Deleted comment", status)
	suite.Nil(err)
}

func (suite *SchemaResolverSuite) TestResolver_DeleteCommentUnauthorized() {
	status, err := suite.mutationResolver.DeleteComment(context.Background(), 1)

	suite.Equal("Unauthorized", status)
	suite.NotNil(err)
}

func (suite *SchemaResolverSuite) TestResolver_DeleteCommentFailure() {
	ctx := context.WithValue(context.Background(), "userID", 1)

	suite.commentMock.On("DeleteComment", ctx, 1, 1).Return(errors.New("error"))

	status, err := suite.mutationResolver.DeleteComment(ctx, 1)

	suite.Equal("Failed to delete comment", status)
	suite.NotNil(err)
}

// ====================================================

func (suite *SchemaResolverSuite) TestResolver_GetCommentsSuccess() {
	postID := 1
	first := 2
	last := 0
	var after *string
	var before *string

	comments := &model.CommentConnection{
		Edges: []*model.CommentEdge{
			{
				Node: &model.Comment{
					ID:     1,
					Body:   "test1",
					UserID: 1,
					PostID: postID,
				},
			},
			{
				Node: &model.Comment{
					ID:     2,
					Body:   "test2",
					UserID: 1,
					PostID: postID,
				},
			},
		},
		PageInfo: &model.PageInfo{
			StartCursor: nil,
			EndCursor:   nil,
			HasNextPage: false,
			HasPrevPage: false,
		},
	}

	suite.commentMock.On("GetCommentsByPostID", mock.Anything, postID, &first, &last, after, before).
		Return(comments, nil)

	result, err := suite.queryResolver.GetComments(context.Background(), postID, &first, &last, after, before)

	suite.NotNil(result)

	suite.Equal(2, len(result.Edges))
	suite.Equal("test1", result.Edges[0].Node.Body)
	suite.Equal("test2", result.Edges[1].Node.Body)

	suite.Nil(err)
}

func (suite *SchemaResolverSuite) TestResolver_GetCommentsWithPagination() {
	postID := 1
	first := 10
	var last *int
	after := time.Now().Add(-time.Hour)
	before := time.Now().Add(time.Hour)
	afterStr := after.Format(time.RFC3339)
	beforeStr := before.Format(time.RFC3339)

	var commentsList []*model.CommentEdge
	for i := 1; i <= 15; i++ {
		commentsList = append(commentsList, &model.CommentEdge{
			Node: &model.Comment{
				ID:        i,
				Body:      fmt.Sprintf("Comment %d", i),
				UserID:    1,
				PostID:    postID,
				CreatedAt: after.Add(time.Duration(i) * time.Minute),
			},
		})
	}

	expectedComments := commentsList[:10]

	startCursor := expectedComments[0].Node.CreatedAt.Format(time.RFC3339)
	endCursor := expectedComments[len(expectedComments)-1].Node.CreatedAt.Format(time.RFC3339)

	comments := &model.CommentConnection{
		Edges: expectedComments,
		PageInfo: &model.PageInfo{
			StartCursor: &startCursor,
			EndCursor:   &endCursor,
			HasNextPage: true,
			HasPrevPage: false,
		},
	}

	suite.commentMock.On("GetCommentsByPostID", mock.Anything, postID, &first, last, mock.Anything, mock.Anything).
		Return(comments, nil)

	result, err := suite.queryResolver.GetComments(context.Background(), postID, &first, last, &afterStr, &beforeStr)

	suite.NotNil(result)
	suite.Nil(err)
	suite.Equal(10, len(result.Edges))
	suite.Equal("Comment 1", result.Edges[0].Node.Body)
	suite.Equal("Comment 10", result.Edges[9].Node.Body)

	suite.NotNil(result.PageInfo)
	suite.Equal(startCursor, *result.PageInfo.StartCursor)
	suite.Equal(endCursor, *result.PageInfo.EndCursor)
	suite.True(result.PageInfo.HasNextPage)
	suite.False(result.PageInfo.HasPrevPage)
}

func (suite *SchemaResolverSuite) TestResolver_GetCommentsFailure() {
	postID := 1
	first := 2
	last := 0
	var after *string
	var before *string

	suite.commentMock.On("GetCommentsByPostID", mock.Anything, postID, &first, &last, after, before).
		Return(nil, errors.New("error"))

	result, err := suite.queryResolver.GetComments(context.Background(), postID, &first, &last, after, before)

	suite.Nil(result)
	suite.NotNil(err)
}

func (suite *SchemaResolverSuite) TestResolver_GetCommentsEmpty() {
	postID := 1
	first := 2
	last := 0
	var after *string
	var before *string

	comments := &model.CommentConnection{
		Edges: []*model.CommentEdge{},
		PageInfo: &model.PageInfo{
			StartCursor: nil,
			EndCursor:   nil,
			HasNextPage: false,
			HasPrevPage: false,
		},
	}

	suite.commentMock.On("GetCommentsByPostID", mock.Anything, postID, &first, &last, after, before).
		Return(comments, nil)

	result, err := suite.queryResolver.GetComments(context.Background(), postID, &first, &last, after, before)

	suite.NotNil(result)
	suite.Equal(0, len(result.Edges))
	suite.Nil(err)
}

func generateStringWith2000Chars() string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	sb.Grow(2002)

	for i := 0; i < 2002; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}

	return sb.String()
}

func strPointer(s string) *string {
	return &s
}

func boolPointer(b bool) *bool {
	return &b
}
