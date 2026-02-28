package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditAction_IsValid(t *testing.T) {
	assert.True(t, AuditInsert.IsValid())
	assert.True(t, AuditUpdate.IsValid())
	assert.True(t, AuditDelete.IsValid())
	assert.False(t, AuditAction("UPSERT").IsValid())
	assert.False(t, AuditAction("").IsValid())
}

func TestAuditEntry_ParseDiff_Success(t *testing.T) {
	diff := json.RawMessage(`{
		"name": {"old": "Laptop", "new": "Gaming Laptop"},
		"price": {"old": 999, "new": 1299}
	}`)

	entry := &AuditEntry{Diff: diff}
	changes, err := entry.ParseDiff()

	require.NoError(t, err)
	assert.Len(t, changes, 2)

	changeMap := make(map[string]FieldChange)
	for _, c := range changes {
		changeMap[c.Field] = c
	}

	assert.Equal(t, "Laptop", changeMap["name"].OldValue)
	assert.Equal(t, "Gaming Laptop", changeMap["name"].NewValue)
}

func TestAuditEntry_ParseDiff_Null(t *testing.T) {
	entry := &AuditEntry{Diff: json.RawMessage(`null`)}
	changes, err := entry.ParseDiff()

	assert.NoError(t, err)
	assert.Nil(t, changes)
}

func TestAuditEntry_ParseDiff_Empty(t *testing.T) {
	entry := &AuditEntry{Diff: nil}
	changes, err := entry.ParseDiff()

	assert.NoError(t, err)
	assert.Nil(t, changes)
}

func TestAuditEntry_ParseDiff_InvalidJSON(t *testing.T) {
	entry := &AuditEntry{Diff: json.RawMessage(`not json`)}
	_, err := entry.ParseDiff()

	assert.Error(t, err)
}
