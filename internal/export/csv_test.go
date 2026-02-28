package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAuditCSV_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := WriteAuditCSV(&buf, nil)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "ID,Item ID,Action")
	// only header, no data rows
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 1)
}

func TestWriteAuditCSV_WithEntries(t *testing.T) {
	itemID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	entries := []*domain.AuditEntryWithUser{
		{
			AuditEntry: domain.AuditEntry{
				ID:        1,
				ItemID:    itemID,
				Action:    domain.AuditInsert,
				ChangedBy: userID,
				ChangedAt: now,
			},
			Username: "admin",
		},
		{
			AuditEntry: domain.AuditEntry{
				ID:        2,
				ItemID:    itemID,
				Action:    domain.AuditUpdate,
				ChangedBy: userID,
				Diff:      json.RawMessage(`{"name":{"old":"Laptop","new":"Gaming Laptop"}}`),
				ChangedAt: now,
			},
			Username: "admin",
		},
	}

	var buf bytes.Buffer
	err := WriteAuditCSV(&buf, entries)

	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 3) // header + 2 rows

	assert.Contains(t, buf.String(), "INSERT")
	assert.Contains(t, buf.String(), "UPDATE")
	assert.Contains(t, buf.String(), "admin")
	assert.Contains(t, buf.String(), "Laptop")
}

func TestFormatDiff_Empty(t *testing.T) {
	assert.Equal(t, "", formatDiff(nil))
	assert.Equal(t, "", formatDiff(json.RawMessage(`null`)))
	assert.Equal(t, "", formatDiff(json.RawMessage(``)))
}

func TestFormatDiff_Valid(t *testing.T) {
	diff := json.RawMessage(`{"name":{"old":"A","new":"B"}}`)
	result := formatDiff(diff)
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "A")
	assert.Contains(t, result, "B")
}

func TestFormatDiff_InvalidJSON(t *testing.T) {
	diff := json.RawMessage(`not json`)
	result := formatDiff(diff)
	assert.Equal(t, "not json", result) // falls back to raw string
}
