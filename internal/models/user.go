// internal/models/user.go
package models

import "time"

type User struct {
	ID       int     `json:"id" db:"id"`
	Name     string  `json:"name" binding:"required"`
	Mail     string  `json:"mail" binding:"required"`
	Password *string `json:"password,omitempty" db:"password"` // Может быть NULL для OAuth

	AvatarURL   *string `json:"avatar_url,omitempty" db:"avatar_url"`
	GitHubID    *int    `json:"github_id,omitempty" db:"github_id"`
	GitHubLogin *string `json:"github_login,omitempty" db:"github_login"`

	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"` // скрываем от JSON
}

// UserResponse - DTO для ответа API
type UserResponse struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Mail      string  `json:"mail"`
	AvatarURL *string `json:"avatar_url"`
}

// Метод преобразования
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Mail:      u.Mail,
		AvatarURL: u.AvatarURL,
	}
}

type GitHubUser struct {
	ID        int     `json:"id"`
	Login     string  `json:"login"`
	Email     *string `json:"email"` // Указатель для nullable поля
	Name      *string `json:"name"`  // Name тоже может быть null
	AvatarURL string  `json:"avatar_url"`
}
