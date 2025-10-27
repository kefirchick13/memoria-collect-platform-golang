// repository/collection_repository.go
package repository

import (
	"database/sql"
	"fmt"

	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/pagination"
	"go.uber.org/zap"
)

type CollectionRepository struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewCollectionPostgres(db *sql.DB, logger *zap.SugaredLogger) *CollectionRepository {
	return &CollectionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *CollectionRepository) CreateCollection(collection *models.Collection) (*models.Collection, error) {
	query := fmt.Sprintf(`
        INSERT INTO %s (name, description, is_public, cover_image, type, user_id) 
        VALUES ($1, $2, $3, $4, $5, $6) 
        RETURNING id, name, description, is_public, cover_image, type, created_at, user_id
    `, collectionsTable)

	var createdCollection models.Collection
	err := r.db.QueryRow(
		query,
		collection.Name,
		collection.Description,
		collection.IsPublic,
		collection.CoverImage,
		collection.Type,
		collection.UserID,
	).Scan(
		&createdCollection.ID,
		&createdCollection.Name,
		&createdCollection.Description,
		&createdCollection.IsPublic,
		&createdCollection.CoverImage,
		&createdCollection.Type,
		&createdCollection.CreatedAt,
		&createdCollection.UserID,
	)

	if err != nil {
		r.logger.Error("Failed to create collection: " + err.Error())
		return nil, err
	}

	return &createdCollection, nil
}

// GetCollectionsWithPagination - метод с пагинацией
func (r *CollectionRepository) GetCollectionsWithPagination(userID int, req pagination.PaginationRequest) (*pagination.PaginatedResponse[models.Collection], error) {
	// Сначала получаем общее количество коллекций пользователя
	countQuery := fmt.Sprintf(`
        SELECT COUNT(*) 
        FROM %s 
        WHERE user_id = $1
    `, collectionsTable)

	var total int64
	err := r.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.Errorf("Failed to count collections for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to count collections: %w", err)
	}

	// Получаем данные с пагинацией
	query := fmt.Sprintf(`
        SELECT id, name, description, is_public, cover_image, created_at, user_id, type
        FROM %s
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `, collectionsTable)

	r.logger.Infof("Getting collections for user_id: %d, page: %d, limit: %d", userID, req.Page(), req.Limit())

	rows, err := r.db.Query(query, userID, req.Limit(), req.Offset())
	if err != nil {
		r.logger.Errorf("Query execution failed: %v", err)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var collections []models.Collection
	for rows.Next() {
		var collection models.Collection
		err = rows.Scan(
			&collection.ID,
			&collection.Name,
			&collection.Description,
			&collection.IsPublic,
			&collection.CoverImage,
			&collection.CreatedAt,
			&collection.UserID,
			&collection.Type,
		)

		if err != nil {
			r.logger.Errorf("Scan failed: %v", err)
			return nil, fmt.Errorf("failed to scan collection: %w", err)
		}

		collections = append(collections, collection)
	}

	if err := rows.Err(); err != nil {
		r.logger.Errorf("Rows iteration error: %v", err)
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return &pagination.PaginatedResponse[models.Collection]{
		Data:       collections,
		Pagination: req.ToPagination(total),
	}, nil
}

func (r *CollectionRepository) GetCollections(userID int) ([]models.Collection, error) {
	// Используем пагинацию с большим лимитом для эмуляции получения всех данных
	req := pagination.NewUnlimitedPagination()

	result, err := r.GetCollectionsWithPagination(userID, req)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (r *CollectionRepository) GetCollectionByID(collectionID string) (*models.Collection, error) {
	query := fmt.Sprintf(`
        SELECT id, name, description, is_public, cover_image, created_at, user_id, type
        FROM %s
        WHERE id = $1
    `, collectionsTable)

	var collection models.Collection
	err := r.db.QueryRow(query, collectionID).Scan(
		&collection.ID,
		&collection.Name,
		&collection.Description,
		&collection.IsPublic,
		&collection.CoverImage,
		&collection.CreatedAt,
		&collection.UserID,
		&collection.Type,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("collection not found")
		}
		r.logger.Errorf("Failed to get collection %d: %v", collectionID, err)
		return nil, err
	}

	return &collection, nil
}

func (r *CollectionRepository) DeleteCollection(collectionID string) error {
	query := fmt.Sprintf(`
        DELETE FROM %s 
        WHERE id = $1
    `, collectionsTable)

	result, err := r.db.Exec(query, collectionID)
	if err != nil {
		r.logger.Errorf("Failed to delete collection %d: %v", collectionID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("collection not found or access denied")
	}

	return nil
}
