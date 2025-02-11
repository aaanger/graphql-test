package graph

import (
	"github.com/aaanger/graphql-test/repository/comment"
	"github.com/aaanger/graphql-test/repository/post"
	"github.com/aaanger/graphql-test/repository/user"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UserRepo    user.IUserRepository
	PostRepo    post.IPostRepository
	CommentRepo comment.ICommentRepository
}
