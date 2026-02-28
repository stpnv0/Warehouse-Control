package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRole_IsValid(t *testing.T) {
	assert.True(t, RoleAdmin.IsValid())
	assert.True(t, RoleManager.IsValid())
	assert.True(t, RoleViewer.IsValid())
	assert.False(t, Role("unknown").IsValid())
	assert.False(t, Role("").IsValid())
}

func TestParseRole(t *testing.T) {
	tests := []struct {
		input   string
		want    Role
		wantErr bool
	}{
		{"admin", RoleAdmin, false},
		{"manager", RoleManager, false},
		{"viewer", RoleViewer, false},
		{"unknown", "", true},
		{"", "", true},
		{"Admin", "", true}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseRole(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRole_Permissions(t *testing.T) {
	tests := []struct {
		role         Role
		canCreate    bool
		canUpdate    bool
		canDelete    bool
		canView      bool
		canViewAudit bool
		canExport    bool
	}{
		{RoleAdmin, true, true, true, true, true, true},
		{RoleManager, true, true, false, true, true, true},
		{RoleViewer, false, false, false, true, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			assert.Equal(t, tt.canCreate, tt.role.CanCreate())
			assert.Equal(t, tt.canUpdate, tt.role.CanUpdate())
			assert.Equal(t, tt.canDelete, tt.role.CanDelete())
			assert.Equal(t, tt.canView, tt.role.CanView())
			assert.Equal(t, tt.canViewAudit, tt.role.CanViewAudit())
			assert.Equal(t, tt.canExport, tt.role.CanExport())
		})
	}
}
