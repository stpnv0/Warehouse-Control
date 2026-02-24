package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stpnv0/WarehouseControl/internal/handler/dto"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/logger"
)

type itemService interface {
	CreateItem(ctx context.Context, claims *domain.AuthClaims, input *domain.CreateItemInput) (*domain.Item, error)
	GetByID(ctx context.Context, claims *domain.AuthClaims, id uuid.UUID) (*domain.Item, error)
	ListItems(ctx context.Context, claims *domain.AuthClaims, filter *domain.ItemFilter, page, pageSize int) (*domain.ItemList, error)
	Update(ctx context.Context, claims *domain.AuthClaims, id uuid.UUID, input *domain.UpdateItemInput) (*domain.Item, error)
	Delete(ctx context.Context, claims *domain.AuthClaims, id uuid.UUID) error
}

type ItemHandler struct {
	service itemService
	log     logger.Logger
}

func NewItemHandler(service itemService, log logger.Logger) *ItemHandler {
	return &ItemHandler{
		service: service,
		log:     log.With("handler", "item"),
	}
}

// POST /api/items
func (h *ItemHandler) Create(c *ginext.Context) {
	claims := getClaims(c)
	if claims == nil {
		return //getClaims уже вызвал abort и записал ответ
	}

	var req dto.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid request body"})
		return
	}

	item, err := h.service.CreateItem(c.Request.Context(), claims, req.ToInput())
	if err != nil {
		writeError(c, err)
		return
	}

	writeJSON(c, http.StatusCreated, dto.NewItemResponse(item))
}

// GET /api/items
func (h *ItemHandler) List(c *ginext.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	filter := &domain.ItemFilter{}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	list, err := h.service.ListItems(c.Request.Context(), claims, filter, page, pageSize)
	if err != nil {
		writeError(c, err)
		return
	}

	writeJSON(c, http.StatusOK, dto.NewItemListFromDomain(list))
}

// GET /api/items/:id
func (h *ItemHandler) GetByID(c *ginext.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid item id"})
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), claims, id)
	if err != nil {
		writeError(c, err)
		return
	}

	writeJSON(c, http.StatusOK, dto.NewItemResponse(item))
}

// PUT /api/items/:id
func (h *ItemHandler) Update(c *ginext.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid item id"})
		return
	}

	var req dto.UpdateItemRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid request body"})
		return
	}

	item, err := h.service.Update(c.Request.Context(), claims, id, req.ToInput())
	if err != nil {
		writeError(c, err)
		return
	}

	writeJSON(c, http.StatusOK, dto.NewItemResponse(item))
}

// DELETE /api/items/:id
func (h *ItemHandler) Delete(c *ginext.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid item id"})
		return
	}

	if err = h.service.Delete(c.Request.Context(), claims, id); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
