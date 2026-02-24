package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	AuditInsert AuditAction = "INSERT"
	AuditUpdate AuditAction = "UPDATE"
	AuditDelete AuditAction = "DELETE"
)

func (a AuditAction) IsValid() bool {
	switch a {
	case AuditInsert, AuditUpdate, AuditDelete:
		return true
	}
	return false
}

// AuditEntry - одна запись из item_audit_log
type AuditEntry struct {
	ID        int64           `json:"id"         db:"id"`
	ItemID    uuid.UUID       `json:"item_id"    db:"item_id"`
	Action    AuditAction     `json:"action"     db:"action"`
	ChangedBy uuid.UUID       `json:"changed_by" db:"changed_by"`
	OldData   json.RawMessage `json:"old_data"   db:"old_data"`
	NewData   json.RawMessage `json:"new_data"   db:"new_data"`
	Diff      json.RawMessage `json:"diff"       db:"diff"`
	ChangedAt time.Time       `json:"changed_at" db:"changed_at"`
}

// AuditEntryWithUser - запись аудита с именем пользователя (для отображения)
type AuditEntryWithUser struct {
	AuditEntry
	Username string `json:"username" db:"username"`
}

// FieldChange - одно изменение поля (для фронтенда)
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// ParseDiff разбирает JSONB diff в слайс изменений
func (a *AuditEntry) ParseDiff() ([]FieldChange, error) {
	if len(a.Diff) == 0 || string(a.Diff) == "null" {
		return nil, nil
	}

	var raw map[string]struct {
		Old interface{} `json:"old"`
		New interface{} `json:"new"`
	}

	if err := json.Unmarshal(a.Diff, &raw); err != nil {
		return nil, err
	}

	changes := make([]FieldChange, 0, len(raw))
	for field, v := range raw {
		changes = append(changes, FieldChange{
			Field:    field,
			OldValue: v.Old,
			NewValue: v.New,
		})
	}
	return changes, nil
}

// AuditFilter - фильтрация истории изменений
type AuditFilter struct {
	ItemID   *uuid.UUID   `json:"item_id"`
	UserID   *uuid.UUID   `json:"user_id"`
	Action   *AuditAction `json:"action"`
	DateFrom *time.Time   `json:"date_from"`
	DateTo   *time.Time   `json:"date_to"`
}

// AuditList - результат постраничного запроса аудита.
type AuditList struct {
	Entries    []*AuditEntryWithUser
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}
