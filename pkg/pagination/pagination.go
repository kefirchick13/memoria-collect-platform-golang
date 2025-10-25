package pagination

import (
	"math"
)

const (
	minPaginationlimit = 10
	maxPaginationLimit = 100
)

type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse[T any] struct {
	Data       []T                `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

type PaginationRequest struct {
	page  int
	limit int
}

// Конструктор с валидацией лимита и страницы
func NewPaginationRequest(page, limit int) PaginationRequest {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = minPaginationlimit
	}
	if limit > 100 {
		limit = maxPaginationLimit
	}

	return PaginationRequest{
		page:  page,
		limit: limit,
	}
}

// Конструктор с безлимитным количеством
func NewUnlimitedPagination() PaginationRequest {
	return PaginationRequest{
		page:  1,
		limit: 0, // 0 может означать "без лимита"
	}
}

// Конструктор с значениями по умолчанию (без ошибок)
func DefaultPaginationRequest() PaginationRequest {
	return PaginationRequest{
		page:  1,
		limit: 10,
	}
}

// Геттеры
func (pr PaginationRequest) Page() int {
	return pr.page
}

func (pr PaginationRequest) Limit() int {
	return pr.limit
}

func (pr PaginationRequest) Offset() int {
	return (pr.page - 1) * pr.limit
}

func (pr PaginationRequest) ToPagination(total int64) PaginationResponse {
	totalPages := int(math.Ceil(float64(total) / float64(pr.limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	return PaginationResponse{
		Page:       pr.page,
		Limit:      pr.limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
