package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Item struct {
	ID        uuid.UUID       `json:"id"         db:"id"`
	Name      string          `json:"name"       db:"name"`
	SKU       string          `json:"sku"        db:"sku"`
	Quantity  int             `json:"quantity"    db:"quantity"`
	Price     decimal.Decimal `json:"price"      db:"price"`
	Location  *string         `json:"location"   db:"location"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// CreateItemInput - DTO для создания товара (от клиента)
type CreateItemInput struct {
	Name     string          `json:"name"     validate:"required,max=255"`
	SKU      string          `json:"sku"      validate:"required,max=64"`
	Quantity int             `json:"quantity" validate:"gte=0"`
	Price    decimal.Decimal `json:"price"    validate:"required"`
	Location *string         `json:"location" validate:"omitempty,max=128"`
}

// UpdateItemInput - DTO для обновления (все поля опциональны, partial update)
type UpdateItemInput struct {
	Name     *string          `json:"name"     validate:"omitempty,max=255"`
	SKU      *string          `json:"sku"      validate:"omitempty,max=64"`
	Quantity *int             `json:"quantity" validate:"omitempty,gte=0"`
	Price    *decimal.Decimal `json:"price"    validate:"omitempty"`
	Location *string          `json:"location" validate:"omitempty,max=128"`
}

// HasChanges - проверяет, что хотя бы одно поле задано
func (u *UpdateItemInput) HasChanges() bool {
	return u.Name != nil ||
		u.SKU != nil ||
		u.Quantity != nil ||
		u.Price != nil ||
		u.Location != nil
}

// ItemFilter - фильтрация и пагинация для GET /items
type ItemFilter struct {
	Search *string `json:"search"`
}
type ItemList struct {
	Items      []*Item
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}
