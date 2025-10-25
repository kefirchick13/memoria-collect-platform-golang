package repository

import (
	"database/sql"
	"fmt"

	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/pagination"
	"go.uber.org/zap"
)

type CollectionItemRepository struct {
	logger *zap.SugaredLogger
	db     *sql.DB
}

func NewCollectionItemPostgres(db *sql.DB, logger *zap.SugaredLogger) *CollectionItemRepository {
	return &CollectionItemRepository{
		logger: logger,
		db:     db,
	}
}

// FOR items ribbon in public view
func (r *CollectionItemRepository) GetAllItemsWithCurrentTypePaginated(currentType string, req pagination.PaginationRequest) (*pagination.PaginatedResponse[models.CollectionItem], error) {

	// Сначала получаем общее количество коллекций пользователя
	countQuery := fmt.Sprintf(`
        SELECT COUNT(*) 
        FROM %s 
        WHERE type = $1
    `, collectionItemsTable)

	var total int64
	err := r.db.QueryRow(countQuery, currentType).Scan(&total)
	if err != nil {
		r.logger.Errorf("Failed to count collection_items with type %d: %v", currentType, err)
		return nil, fmt.Errorf("failed to count collection_items: %w", err)
	}

	// Order by ASC - по возрастанию(сначала is_custom = false) для ленты
	query := fmt.Sprintf(`
		SELECT id, type, title, description, cover_image, is_public, is_custom, created_at, updated_at 
		FROM %s 
		WHERE type = $1 AND is_public = TRUE ORDER BY is_custom ASC
		LIMIT $2 OFFSET $3`, collectionItemsTable)

	rows, err := r.db.Query(query, currentType, req.Limit(), req.Offset())
	if err != nil {
		r.logger.Errorf("Error executing query: %v", err)
		return nil, fmt.Errorf("failed to get collection items: %w", err)
	}
	defer rows.Close()

	var items []models.CollectionItem

	for rows.Next() {
		var item models.CollectionItem
		err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Title,
			&item.Description,
			&item.CoverImage,
			&item.IsPublic,
			&item.IsCustom,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			r.logger.Errorf("Error scanning row: %v", err)
			return nil, fmt.Errorf("failed to scan collection item: %w", err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		r.logger.Errorf("Error iterating rows: %v", err)
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return &pagination.PaginatedResponse[models.CollectionItem]{
		Data:       items,
		Pagination: req.ToPagination(total),
	}, nil
}

// For private Collection
func (r *CollectionItemRepository) GetItemsByCollection(collection_id string) ([]models.CollectionItem, error) {
	query := fmt.Sprintf(`
	SELECT ci.id, ci.type, ci.title, ci.description, ci.cover_image, ci.is_public, is_custom, ci.created_at, ci.updated_at 
	FROM %s ci
	JOIN %s cia ON ci.id = cia.item_id
	WHERE cia.collection_id = $1
	ORDER BY cia.added_at DESC
	`, collectionItemsTable, collectionItemsAssignmentTable)

	var items []models.CollectionItem
	rows, err := r.db.Query(query, collection_id)
	if err != nil {
		r.logger.Errorf("Error executing query: %v", err)
		return nil, fmt.Errorf("failed to get collection items: %w", err)
	}

	for rows.Next() {
		var item models.CollectionItem
		rows.Scan(
			&item.ID,
			&item.Type,
			&item.Title,
			&item.Description,
			&item.CoverImage,
			&item.IsPublic,
			&item.IsCustom,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			r.logger.Errorf("Error scanning row: %v", err)
			return nil, fmt.Errorf("failed to scan collection item: %w", err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		r.logger.Errorf("Error iterating rows: %v", err)
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return items, nil
}

func (r *CollectionItemRepository) GetItemByID(id string) (*models.CollectionItem, error) {
	query := fmt.Sprintf("SELECT id, type, title, description, cover_image, is_custom, created_at, updated_at FROM %s WHERE id = $1", collectionItemsTable)

	var collectionItem models.CollectionItem
	err := r.db.QueryRow(
		query,
		id,
	).Scan(
		&collectionItem.ID,
		&collectionItem.Type,
		&collectionItem.Title,
		&collectionItem.Description,
		&collectionItem.CoverImage,
		&collectionItem.IsCustom,
		&collectionItem.CreatedAt,
		&collectionItem.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to execute collectionItem: " + err.Error())
	}
	return &collectionItem, nil
}

func (r *CollectionItemRepository) CreateItem(collectionItem *models.CollectionItem) (string, error) {
	query := fmt.Sprintf(
		`INSERT INTO %s (type, title, description, cover_image, is_custom, is_public, creator_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
		`, collectionItemsTable)

	var id string
	err := r.db.QueryRow(
		query,
		&collectionItem.Type,
		&collectionItem.Title,
		&collectionItem.Description,
		&collectionItem.CoverImage,
		&collectionItem.IsCustom,
		&collectionItem.IsPublic,
		&collectionItem.CreatorID,
	).Scan(
		&id,
	)

	if err != nil {
		r.logger.Errorf("Error creating item collection: ", err.Error())
		return "", err
	}

	return id, nil

}

func (r *CollectionItemRepository) DeleteCollectionItem(id string) error {
	query := fmt.Sprintf(`
	DELETE FROM %s WHERE id = $1
	`, collectionItemsTable)
	_, err := r.db.Query(query, id)
	return err
}

func (r *CollectionItemRepository) UpdateCollectionItem(item *models.CollectionItem) error {
	query := fmt.Sprintf(`
        UPDATE %s 
        SET type = $1, 
            title = $2, 
            description = $3, 
            cover_image = $4, 
            is_custom = $5, 
            updated_at = NOW()
        WHERE id = $6
    `, collectionItemsTable)

	result, err := r.db.Exec(
		query,
		item.Type,
		item.Title,
		item.Description,
		item.CoverImage,
		item.IsCustom,
		item.ID,
	)

	if err != nil {
		r.logger.Errorf("Error updating collection item: %v", err)
		return fmt.Errorf("failed to update collection item: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("collection item not found")
	}

	return nil
}

func (r *CollectionItemRepository) AddItemToCollection(collection_id string, item_id string, user_review string) (int, error) {
	query := fmt.Sprintf(`INSERT INTO %s (collection_id, item_id, user_review) VALUES ($1, $2, $3) RETURNING id`, collectionItemsAssignmentTable)

	var id int
	err := r.db.QueryRow(query, collection_id, item_id, user_review).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
