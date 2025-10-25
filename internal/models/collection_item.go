package models

import "time"

type CollectionItem struct {
	ID          string    `json:"id" db:"id"`
	Type        string    `json:"type" db:"type"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	CoverImage  *string   `json:"cover_image" db:"cover_image"`
	IsCustom    bool      `json:"is_custom" db:"is_custom"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	CreatorID   *int      `json:"creator_id" db:"creator_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
