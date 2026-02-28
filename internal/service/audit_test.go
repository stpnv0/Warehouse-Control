package service

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newAuditService(t *testing.T) (*AuditService, *mockauditRepository) {
	repo := newMockauditRepository(t)
	svc := NewAuditService(repo, newTestLogger())
	return svc, repo
}

func TestAuditService_GetByItemID_Success(t *testing.T) {
	svc, repo := newAuditService(t)

	itemID := uuid.New()
	entries := []*domain.AuditEntryWithUser{
		{
			AuditEntry: domain.AuditEntry{
				ID:        1,
				ItemID:    itemID,
				Action:    domain.AuditInsert,
				ChangedBy: adminClaims.UserID,
				ChangedAt: time.Now(),
			},
			Username: "admin",
		},
	}

	repo.EXPECT().GetByItemID(mock.Anything, itemID).Return(entries, nil)

	result, err := svc.GetByItemID(context.Background(), adminClaims, itemID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, domain.AuditInsert, result[0].Action)
}

func TestAuditService_GetByItemID_ViewerForbidden(t *testing.T) {
	svc, _ := newAuditService(t)

	_, err := svc.GetByItemID(context.Background(), viewerClaims, uuid.New())

	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAuditService_GetByItemID_NilResultReturnsEmptySlice(t *testing.T) {
	svc, repo := newAuditService(t)

	itemID := uuid.New()
	repo.EXPECT().GetByItemID(mock.Anything, itemID).Return(nil, nil)

	result, err := svc.GetByItemID(context.Background(), adminClaims, itemID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestAuditService_GetByItemID_RepoError(t *testing.T) {
	svc, repo := newAuditService(t)

	itemID := uuid.New()
	repo.EXPECT().GetByItemID(mock.Anything, itemID).Return(nil, errors.New("db error"))

	_, err := svc.GetByItemID(context.Background(), adminClaims, itemID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AuditService.GetByItemID")
}

func TestAuditService_List_Success(t *testing.T) {
	svc, repo := newAuditService(t)

	entries := []*domain.AuditEntryWithUser{
		{
			AuditEntry: domain.AuditEntry{
				ID:     1,
				Action: domain.AuditInsert,
			},
			Username: "admin",
		},
	}

	filter := &domain.AuditFilter{}
	repo.EXPECT().List(mock.Anything, filter, 20, 0).Return(entries, int64(1), nil)

	result, err := svc.List(context.Background(), adminClaims, filter, 1, 20)

	assert.NoError(t, err)
	assert.Len(t, result.Entries, 1)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, 1, result.TotalPages)
}

func TestAuditService_List_ViewerForbidden(t *testing.T) {
	svc, _ := newAuditService(t)

	_, err := svc.List(context.Background(), viewerClaims, &domain.AuditFilter{}, 1, 20)

	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAuditService_List_InvalidAction(t *testing.T) {
	svc, _ := newAuditService(t)

	badAction := domain.AuditAction("INVALID")
	filter := &domain.AuditFilter{Action: &badAction}

	_, err := svc.List(context.Background(), adminClaims, filter, 1, 20)

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestAuditService_List_RepoError(t *testing.T) {
	svc, repo := newAuditService(t)

	filter := &domain.AuditFilter{}
	repo.EXPECT().List(mock.Anything, filter, 20, 0).Return(nil, int64(0), errors.New("db error"))

	_, err := svc.List(context.Background(), adminClaims, filter, 1, 20)

	assert.Error(t, err)
}

func TestAuditService_List_NilResultReturnsEmptySlice(t *testing.T) {
	svc, repo := newAuditService(t)

	filter := &domain.AuditFilter{}
	repo.EXPECT().List(mock.Anything, filter, 20, 0).Return(nil, int64(0), nil)

	result, err := svc.List(context.Background(), adminClaims, filter, 1, 20)

	assert.NoError(t, err)
	assert.NotNil(t, result.Entries)
	assert.Empty(t, result.Entries)
}

func TestAuditService_ExportCSV_Success(t *testing.T) {
	svc, repo := newAuditService(t)

	entries := []*domain.AuditEntryWithUser{
		{
			AuditEntry: domain.AuditEntry{
				ID:        1,
				ItemID:    uuid.New(),
				Action:    domain.AuditInsert,
				ChangedBy: adminClaims.UserID,
				ChangedAt: time.Now(),
			},
			Username: "admin",
		},
	}

	filter := &domain.AuditFilter{}
	repo.EXPECT().List(mock.Anything, filter, maxExportRows, 0).Return(entries, int64(1), nil)

	var buf bytes.Buffer
	err := svc.ExportCSV(context.Background(), adminClaims, filter, &buf)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "ID,Item ID,Action")
	assert.Contains(t, buf.String(), "admin")
}

func TestAuditService_ExportCSV_ViewerForbidden(t *testing.T) {
	svc, _ := newAuditService(t)

	var buf bytes.Buffer
	err := svc.ExportCSV(context.Background(), viewerClaims, &domain.AuditFilter{}, &buf)

	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAuditService_ExportCSV_RepoError(t *testing.T) {
	svc, repo := newAuditService(t)

	filter := &domain.AuditFilter{}
	repo.EXPECT().List(mock.Anything, filter, maxExportRows, 0).Return(nil, int64(0), errors.New("db error"))

	var buf bytes.Buffer
	err := svc.ExportCSV(context.Background(), adminClaims, filter, &buf)

	assert.Error(t, err)
}
