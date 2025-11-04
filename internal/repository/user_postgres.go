// internal/repository/user_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"go.uber.org/zap"
)

type userRepository struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewUserPostgres(db *sql.DB, logger *zap.SugaredLogger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

func (r *userRepository) CreateUser(user *models.User) (*models.User, error) {
	query := `INSERT INTO users (name, mail, password, github_id, github_login, avatar_url, created_at, updated_at, last_login_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
	          RETURNING id`

	var lastLoginAt interface{}
	if user.LastLoginAt != nil {
		lastLoginAt = *user.LastLoginAt
	} else {
		lastLoginAt = nil
	}

	row := r.db.QueryRow(query,
		user.Name, user.Mail, user.Password, user.GitHubID,
		user.GitHubLogin, user.AvatarURL, user.CreatedAt, user.UpdatedAt, lastLoginAt,
	)

	var id int
	if err := row.Scan(&id); err != nil {
		r.logger.Errorf("Failed to create user: %v", err)
		return nil, err
	}

	user.ID = id
	return user, nil
}

func (r *userRepository) GetUserByID(id int) (*models.User, error) {
	query := `SELECT id, name, mail, password, avatar_url, github_id, github_login, 
	                 created_at, updated_at, last_login_at
	          FROM users WHERE id = $1 AND deleted_at IS NULL`

	var user models.User
	var password, avatarURL, githubLogin sql.NullString
	var githubID sql.NullInt64
	var lastLoginAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Mail, &password, &avatarURL,
		&githubID, &githubLogin, &user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Errorf("Failed to get user by ID %d: %v", id, err)
		return nil, err
	}

	// Обрабатываем nullable поля
	if password.Valid {
		user.Password = &password.String
	}
	if avatarURL.Valid {
		user.AvatarURL = &avatarURL.String
	}
	if githubID.Valid {
		githubIDInt := int(githubID.Int64)
		user.GitHubID = &githubIDInt
	}
	if githubLogin.Valid {
		user.GitHubLogin = &githubLogin.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, name, mail, password, avatar_url, github_id, github_login, 
	                 created_at, updated_at, last_login_at
	          FROM users WHERE mail = $1 AND deleted_at IS NULL`

	var user models.User
	var password, avatarURL, githubLogin sql.NullString
	var githubID sql.NullInt64
	var lastLoginAt sql.NullTime

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Mail, &password, &avatarURL,
		&githubID, &githubLogin, &user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Errorf("Failed to get user by email %s: %v", email, err)
		return nil, err
	}

	if password.Valid {
		user.Password = &password.String
	}
	if avatarURL.Valid {
		user.AvatarURL = &avatarURL.String
	}
	if githubID.Valid {
		githubIDInt := int(githubID.Int64)
		user.GitHubID = &githubIDInt
	}
	if githubLogin.Valid {
		user.GitHubLogin = &githubLogin.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (r *userRepository) GetUserByGitHubID(githubID int) (*models.User, error) {
	query := `SELECT id, name, mail, password, avatar_url, github_id, github_login, 
	                 created_at, updated_at, last_login_at
	          FROM users WHERE github_id = $1 AND deleted_at IS NULL`

	var user models.User
	var password, avatarURL, githubLogin sql.NullString
	var dbGitHubID sql.NullInt64
	var lastLoginAt sql.NullTime

	err := r.db.QueryRow(query, githubID).Scan(
		&user.ID, &user.Name, &user.Mail, &password, &avatarURL,
		&dbGitHubID, &githubLogin, &user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Errorf("Failed to get user by GitHub ID %d: %v", githubID, err)
		return nil, err
	}

	if password.Valid {
		user.Password = &password.String
	}
	if avatarURL.Valid {
		user.AvatarURL = &avatarURL.String
	}
	if dbGitHubID.Valid {
		githubIDInt := int(dbGitHubID.Int64)
		user.GitHubID = &githubIDInt
	}
	if githubLogin.Valid {
		user.GitHubLogin = &githubLogin.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (r *userRepository) UpdateUser(user *models.User) error {
	query := `UPDATE users 
	          SET name = $1, mail = $2, password = $3, avatar_url = $4, 
	              github_id = $5, github_login = $6, updated_at = $7, last_login_at = $8
	          WHERE id = $9 AND deleted_at IS NULL`

	result, err := r.db.Exec(query,
		user.Name, user.Mail, user.Password, user.AvatarURL,
		user.GitHubID, user.GitHubLogin, time.Now(), user.LastLoginAt, user.ID,
	)

	if err != nil {
		r.logger.Errorf("Failed to update user %d: %v", user.ID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

func (r *userRepository) DeleteUser(id int) error {
	// Soft delete
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		r.logger.Errorf("Failed to delete user %d: %v", id, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

func (r *userRepository) LinkGitHubToExistingUser(userID int, githubUser *models.GitHubUser) (*models.User, error) {
	query := `UPDATE users 
	          SET github_id = $1, github_login = $2, avatar_url = COALESCE($3, avatar_url), 
	              updated_at = $4, last_login_at = $5
	          WHERE id = $6 AND deleted_at IS NULL
	          RETURNING id, name, mail, avatar_url, github_id, github_login, created_at, updated_at, last_login_at`

	var user models.User
	var avatarURL, githubLogin sql.NullString
	var githubID sql.NullInt64
	var lastLoginAt sql.NullTime

	err := r.db.QueryRow(query,
		githubUser.ID, githubUser.Login, githubUser.AvatarURL, time.Now(), time.Now(), userID,
	).Scan(
		&user.ID, &user.Name, &user.Mail, &avatarURL, &githubID, &githubLogin,
		&user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
	)

	if err != nil {
		r.logger.Errorf("Failed to link GitHub account for user %d: %v", userID, err)
		return nil, err
	}

	if avatarURL.Valid {
		user.AvatarURL = &avatarURL.String
	}
	if githubID.Valid {
		githubIDInt := int(githubID.Int64)
		user.GitHubID = &githubIDInt
	}
	if githubLogin.Valid {
		user.GitHubLogin = &githubLogin.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (r *userRepository) UpdateUserPassword(userID int, hashedPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = $2 
	          WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, hashedPassword, time.Now(), userID)
	if err != nil {
		r.logger.Errorf("Failed to update password for user %d: %v", userID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

func (r *userRepository) UpdateLastLogin(userID int) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, time.Now(), userID)
	if err != nil {
		r.logger.Errorf("Failed to update last login for user %d: %v", userID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

func (r *userRepository) GetActiveUsersCount(since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM users 
	          WHERE last_login_at >= $1 AND deleted_at IS NULL`

	var count int
	err := r.db.QueryRow(query, since).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
