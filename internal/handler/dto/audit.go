package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
)

// DTO для записи аудита
type AuditEntryResponse struct {
	ID        int64            `json:"id"`
	ItemID    uuid.UUID        `json:"item_id"`
	Action    string           `json:"action"`
	ChangedBy uuid.UUID        `json:"changed_by"`
	Username  string           `json:"username"`
	OldData   json.RawMessage  `json:"old_data,omitempty"`
	NewData   json.RawMessage  `json:"new_data,omitempty"`
	Diff      json.RawMessage  `json:"diff,omitempty"`
	Changes   []FieldChangeDTO `json:"changes,omitempty"`
	ChangedAt time.Time        `json:"changed_at"`
}

// FieldChangeDTO - одно изменённое поле.
type FieldChangeDTO struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

func NewAuditEntryResponse(e *domain.AuditEntryWithUser) *AuditEntryResponse {
	resp := &AuditEntryResponse{
		ID:        e.ID,
		ItemID:    e.ItemID,
		Action:    string(e.Action),
		ChangedBy: e.ChangedBy,
		Username:  e.Username,
		OldData:   e.OldData,
		NewData:   e.NewData,
		Diff:      e.Diff,
		ChangedAt: e.ChangedAt,
	}

	// Парсинг diff в формат для фронтенда
	if changes, err := e.ParseDiff(); err == nil && len(changes) > 0 {
		resp.Changes = make([]FieldChangeDTO, 0, len(changes))
		for _, c := range changes {
			resp.Changes = append(resp.Changes, FieldChangeDTO{
				Field:    c.Field,
				OldValue: c.OldValue,
				NewValue: c.NewValue,
			})
		}
	}

	return resp
}

func NewAuditListResponse(entries []*domain.AuditEntryWithUser) []*AuditEntryResponse {
	resp := make([]*AuditEntryResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, NewAuditEntryResponse(e))
	}
	return resp
}

// AuditListResponse - DTO для списка аудита
type AuditListResponse struct {
	Entries    []*AuditEntryResponse `json:"entries"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

func NewAuditListFromDomain(list *domain.AuditList) *AuditListResponse {
	return &AuditListResponse{
		Entries:    NewAuditListResponse(list.Entries),
		Total:      list.Total,
		Page:       list.Page,
		PageSize:   list.PageSize,
		TotalPages: list.TotalPages,
	}
}
