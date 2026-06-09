package blog

import (
	"cellar-backend/internal/modules/user"
	"time"
)

type CreateBlogRequest struct {
	Title    string `json:"title" validate:"required,min=3,max=200"`
	Desc     string `json:"desc" validate:"required,min=10"`
	Category string `json:"category" validate:"required,oneof=Education Guides Stories Tips"`
}

type UpdateBlogRequest struct {
	Title    string `json:"title" validate:"omitempty,min=3,max=200"`
	Desc     string `json:"desc" validate:"omitempty,min=10"`
	Category string `json:"category" validate:"omitempty,oneof=Education Guides Stories Tips"`
}

type BlogResponse struct {
	ID        string             `json:"id"`
	UserID    *string            `json:"user_id,omitempty"`
	Author    *user.UserResponse `json:"author,omitempty"`
	Title     string             `json:"title"`
	Slug      string             `json:"slug"`
	Desc      string             `json:"desc"`
	Category  string             `json:"category"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func ToBlogResponse(b *Blog) BlogResponse {
	var userIDStr *string
	if b.UserID != nil {
		str := b.UserID.String()
		userIDStr = &str
	}

	var authorRes *user.UserResponse
	if b.Author != nil {
		res := user.ToUserResponse(b.Author)
		authorRes = &res
	}

	return BlogResponse{
		ID:        b.ID.String(),
		UserID:    userIDStr,
		Author:    authorRes,
		Title:     b.Title,
		Slug:      b.Slug,
		Desc:      b.Desc,
		Category:  b.Category,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

func ToBlogResponses(blogs []Blog) []BlogResponse {
	res := make([]BlogResponse, len(blogs))
	for i, b := range blogs {
		res[i] = ToBlogResponse(&b)
	}
	return res
}
