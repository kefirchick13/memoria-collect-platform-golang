package models

import "time"

type Collection struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" binding:"required"`

	Description string    `json:"description" db:"description"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	CoverImage  *string   `json:"cover_image" db:"cover_image"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`

	UserID int    `json:"user_id" db:"user_id"`
	Type   string `json:"type" db:"type"`
}
