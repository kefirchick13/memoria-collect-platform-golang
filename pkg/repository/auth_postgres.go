package repository

import (
	"database/sql"
	"fmt"

	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/models"
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
	query := fmt.Sprintf("INSERT INTO %s (name, mail, password, auth_provider, github_id, github_login, avatar_url, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id", usersTable)

	row := r.db.QueryRow(query, user.Name, user.Mail, user.Password, user.AuthProvider, user.GitHubID, user.GitHubLogin, user.AvatarURL, user.CreatedAt, user.UpdatedAt)

	var id int
	if err := row.Scan(&id); err != nil {
		r.logger.Error(err.Error())
		return nil, err
	}

	user.ID = id
	return user, nil
}

func (r *AuthRepository) LinkGitHubToExistingUser(userID int, githubUser *models.GitHubUser) (*models.User, error) {
	// TODO: Implement this method to link GitHub account to existing user
	return nil, fmt.Errorf("not implemented")
}

// В pkg/repository/auth_postgres.go
func (r *AuthRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, name, mail, password, avatar_url, github_id, github_login, auth_provider, created_at 
	          FROM users WHERE mail = $1`

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Mail, &user.Password, &user.AvatarURL,
		&user.GitHubID, &user.GitHubLogin, &user.AuthProvider, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) GetUserByGitHubID(githubID int) (*models.User, error) {
	query := `SELECT id, name, mail, password, avatar_url, github_id, github_login, auth_provider, created_at 
	          FROM users WHERE github_id = $1`

	var user models.User
	err := r.db.QueryRow(query, githubID).Scan(
		&user.ID, &user.Name, &user.Mail, &user.Password, &user.AvatarURL,
		&user.GitHubID, &user.GitHubLogin, &user.AuthProvider, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
