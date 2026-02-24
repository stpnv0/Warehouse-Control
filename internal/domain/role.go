package domain

import "fmt"

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleViewer  Role = "viewer"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleManager, RoleViewer:
		return true
	}
	return false
}

func ParseRole(s string) (Role, error) {
	r := Role(s)
	if !r.IsValid() {
		return "", fmt.Errorf("unknown role: %q", s)
	}
	return r, nil
}

func (r Role) CanCreate() bool { return r == RoleAdmin || r == RoleManager }
func (r Role) CanUpdate() bool { return r == RoleAdmin || r == RoleManager }
func (r Role) CanDelete() bool { return r == RoleAdmin }
func (r Role) CanView() bool   { return true }

// CanViewAudit - просмотр истории изменений
func (r Role) CanViewAudit() bool { return r == RoleAdmin || r == RoleManager }

// CanExport - экспорт в CSV
func (r Role) CanExport() bool { return r == RoleAdmin || r == RoleManager }
