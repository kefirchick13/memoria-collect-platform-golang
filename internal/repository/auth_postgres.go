package repository

import (
	"database/sql"
	"fmt"

	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"go.uber.org/zap"
)

type AuthRepository struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewAuthPostgres(db *sql.DB, logger *zap.SugaredLogger) *AuthRepository {
	return &AuthRepository{
		db:     db,
		logger: logger,
	}
}

func (r *AuthRepository) CreateUser(user *models.User) (*models.User, error) {
	query := fmt.Sprintf("INSERT INTO %s (name, mail, password, github_id, github_login, avatar_url, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id", usersTable)

	row := r.db.QueryRow(query, user.Name, user.Mail, user.Password, user.GitHubID, user.GitHubLogin, user.AvatarURL, user.CreatedAt, user.UpdatedAt)

	var id int
	if err := row.Scan(&id); err != nil {
		r.logger.Error(err.Error())
		return nil, err
	}

	user.ID = id
	return user, nil
}

func (r *AuthRepository) LinkGitHubToExistingUser(userID int, githubUser *models.GitHubUser) (*models.User, error) {
	updateQuery := `UPDATE users SET github_id = $1, github_login = $2 WHERE id = $3`

	result, err := r.db.Exec(updateQuery, githubUser.ID, githubUser.Login, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Проверяем, что обновление прошло
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("no rows were updated")
	}

	// Получаем обновленные данные
	var user models.User
	selectQuery := `SELECT id, email, github_id, github_login, created_at, updated_at FROM users WHERE id = $1`
	err = r.db.QueryRow(selectQuery, userID).Scan(
		&user.ID,
		&user.Mail,
		&user.GitHubID,
		&user.GitHubLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	return &user, nil
}

// В pkg/repository/auth_postgres.go
func (r *AuthRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, name, mail, password, avatar_url, github_id, github_login, created_at 
	          FROM users WHERE mail = $1`

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Mail, &user.Password, &user.AvatarURL,
		&user.GitHubID, &user.GitHubLogin, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) GetUserByGitHubID(githubID int) (*models.User, error) {
	query := `SELECT id, name, mail, password, avatar_url, github_id, github_login, created_at 
	          FROM users WHERE github_id = $1`

	var user models.User
	err := r.db.QueryRow(query, githubID).Scan(
		&user.ID, &user.Name, &user.Mail, &user.Password, &user.AvatarURL,
		&user.GitHubID, &user.GitHubLogin, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
