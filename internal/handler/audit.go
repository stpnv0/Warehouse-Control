package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stpnv0/WarehouseControl/internal/handler/dto"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/logger"
)

type auditService interface {
	GetByItemID(ctx context.Context, claims *domain.AuthClaims, itemID uuid.UUID) ([]*domain.AuditEntryWithUser, error)
	List(ctx context.Context, claims *domain.AuthClaims, filter *domain.AuditFilter, page, pageSize int) (*domain.AuditList, error)
	ExportCSV(ctx context.Context, claims *domain.AuthClaims, filter *domain.AuditFilter, w io.Writer) error
}

type AuditHandler struct {
	service auditService
	log     logger.Logger
}

func NewAuditHandler(service auditService, log logger.Logger) *AuditHandler {
	return &AuditHandler{
		service: service,
		log:     log.With("handler", "audit"),
	}
}

// GET /api/items/:id/audit
func (h *AuditHandler) GetByItemID(c *ginext.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid item id"})
		return
	}

	entries, err := h.service.GetByItemID(c.Request.Context(), claims, itemID)
	if err != nil {
		writeError(c, err)
		return
	}

	writeJSON(c, http.StatusOK, dto.NewAuditListResponse(entries))
}

// GET /api/audit
func (h *AuditHandler) List(c *ginext.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	filter, err := h.parseAuditFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	list, err := h.service.List(c.Request.Context(), claims, filter, page, pageSize)
	if err != nil {
		writeError(c, err)
		return
	}

	writeJSON(c, http.StatusOK, dto.NewAuditListFromDomain(list))
}

// GET /api/audit/export
func (h *AuditHandler) ExportCSV(c *ginext.Context) {
	claims := getClaims(c)
	if claims == nil {
		return //getClaims уже вызвал abort и записал ответ
	}

	filter, err := h.parseAuditFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	var buf bytes.Buffer
	if err = h.service.ExportCSV(c.Request.Context(), claims, filter, &buf); err != nil {
		writeError(c, err)
		return
	}

	filename := fmt.Sprintf("audit_%s.csv", time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

// parseAuditFilter - парсинг query-параметров
func (h *AuditHandler) parseAuditFilter(c *ginext.Context) (*domain.AuditFilter, error) {
	filter := &domain.AuditFilter{}

	if v := c.Query("item_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("invalid item_id: %s", v)
		}
		filter.ItemID = &id
	}

	if v := c.Query("user_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id: %s", v)
		}
		filter.UserID = &id
	}

	if v := c.Query("action"); v != "" {
		action := domain.AuditAction(v)
		if !action.IsValid() {
			return nil, fmt.Errorf("invalid action: %s (allowed: INSERT, UPDATE, DELETE)", v)
		}
		filter.Action = &action
	}

	if v := c.Query("date_from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from: use RFC3339 format")
		}
		filter.DateFrom = &t
	}

	if v := c.Query("date_to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to: use RFC3339 format")
		}
		filter.DateTo = &t
	}

	return filter, nil
}
