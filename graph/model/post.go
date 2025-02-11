package model

import "time"

type Post struct {
	ID            int                `json:"id"`
	UserID        int                `json:"-"`
	User          *User              `json:"user"`
	Title         string             `json:"title"`
	Body          string             `json:"body"`
	AllowComments bool               `json:"allowComments"`
	CreatedAt     time.Time          `json:"createdAt"`
	Comments      *CommentConnection `json:"comments,omitempty"`
}
