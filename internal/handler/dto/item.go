package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stpnv0/WarehouseControl/internal/domain"
)

// DTO для POST /api/items.
type CreateItemRequest struct {
	Name     string          `json:"name"     binding:"required,max=255"`
	SKU      string          `json:"sku"      binding:"required,max=64"`
	Quantity int             `json:"quantity"  binding:"gte=0"`
	Price    decimal.Decimal `json:"price"    binding:"required"`
	Location *string         `json:"location" binding:"omitempty,max=128"`
}

func (r *CreateItemRequest) ToInput() *domain.CreateItemInput {
	return &domain.CreateItemInput{
		Name:     r.Name,
		SKU:      r.SKU,
		Quantity: r.Quantity,
		Price:    r.Price,
		Location: r.Location,
	}
}

// DTO для PUT /api/items/:id.
type UpdateItemRequest struct {
	Name     *string          `json:"name"     binding:"omitempty,max=255"`
	SKU      *string          `json:"sku"      binding:"omitempty,max=64"`
	Quantity *int             `json:"quantity"  binding:"omitempty,gte=0"`
	Price    *decimal.Decimal `json:"price"`
	Location *string          `json:"location" binding:"omitempty,max=128"`
}

func (r *UpdateItemRequest) ToInput() *domain.UpdateItemInput {
	return &domain.UpdateItemInput{
		Name:     r.Name,
		SKU:      r.SKU,
		Quantity: r.Quantity,
		Price:    r.Price,
		Location: r.Location,
	}
}

// ItemResponse - DTO ответа для одного товара
type ItemResponse struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	SKU       string          `json:"sku"`
	Quantity  int             `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
	Location  *string         `json:"location,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func NewItemResponse(item *domain.Item) *ItemResponse {
	return &ItemResponse{
		ID:        item.ID,
		Name:      item.Name,
		SKU:       item.SKU,
		Quantity:  item.Quantity,
		Price:     item.Price,
		Location:  item.Location,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func NewItemListResponse(items []*domain.Item) []*ItemResponse {
	resp := make([]*ItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, NewItemResponse(item))
	}
	return resp
}

// ItemListResponse - DTO ответа для списка товаров с пагинацией
type ItemListResponse struct {
	Items      []*ItemResponse `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

func NewItemListFromDomain(list *domain.ItemList) *ItemListResponse {
	return &ItemListResponse{
		Items:      NewItemListResponse(list.Items),
		Total:      list.Total,
		Page:       list.Page,
		PageSize:   list.PageSize,
		TotalPages: list.TotalPages,
	}
}
