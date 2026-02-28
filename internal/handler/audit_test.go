package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stpnv0/WarehouseControl/internal/handler/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuditHandler_GetByItemID_Success(t *testing.T) {
	svc := newMockauditService(t)
	h := NewAuditHandler(svc, newTestLogger())

	itemID := uuid.New()
	entries := []*domain.AuditEntryWithUser{
		{
			AuditEntry: domain.AuditEntry{
				ID:        1,
				ItemID:    itemID,
				Action:    domain.AuditInsert,
				ChangedBy: testAdminClaims.UserID,
				ChangedAt: time.Now(),
			},
			Username: "admin",
		},
	}

	svc.EXPECT().GetByItemID(mock.Anything, testAdminClaims, itemID).Return(entries, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/items/%s/audit", itemID), nil)
	c.Params = gin.Params{{Key: "id", Value: itemID.String()}}
	setAuthClaims(c, testAdminClaims)

	h.GetByItemID(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []*dto.AuditEntryResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
	assert.Equal(t, "INSERT", resp[0].Action)
}

func TestAuditHandler_GetByItemID_InvalidID(t *testing.T) {
	svc := newMockauditService(t)
	h := NewAuditHandler(svc, newTestLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/items/bad/audit", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	setAuthClaims(c, testAdminClaims)

	h.GetByItemID(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuditHandler_GetByItemID_Forbidden(t *testing.T) {
	svc := newMockauditService(t)
	h := NewAuditHandler(svc, newTestLogger())

	itemID := uuid.New()
	svc.EXPECT().GetByItemID(mock.Anything, testViewerClaims, itemID).Return(nil, domain.ErrForbidden)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/items/%s/audit", itemID), nil)
	c.Params = gin.Params{{Key: "id", Value: itemID.String()}}
	setAuthClaims(c, testViewerClaims)

	h.GetByItemID(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuditHandler_List_Success(t *testing.T) {
	svc := newMockauditService(t)
	h := NewAuditHandler(svc, newTestLogger())

	list := &domain.AuditList{
		Entries:    []*domain.AuditEntryWithUser{},
		Total:      0,
		Page:       1,
		PageSize:   20,
		TotalPages: 0,
	}

	svc.EXPECT().List(mock.Anything, testAdminClaims, mock.Anything, 0, 0).Return(list, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/audit", nil)
	setAuthClaims(c, testAdminClaims)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.AuditListResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, int64(0), resp.Total)
}

func TestAuditHandler_List_WithFilters(t *testing.T) {
	svc := newMockauditService(t)
	h := NewAuditHandler(svc, newTestLogger())

	list := &domain.AuditList{
		Entries:    []*domain.AuditEntryWithUser{},
		Total:      0,
		Page:       1,
		PageSize:   20,
		TotalPages: 0,
	}

	svc.EXPECT().List(mock.Anything, testAdminClaims, mock.Anything, 0, 0).Return(list, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/audit?action=INSERT", nil)
	setAuthClaims(c, testAdminClaims)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditHandler_List_InvalidAction(t *testing.T) {
	svc := newMockauditService(t)
	h := NewAuditHandler(svc, newTestLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/audit?action=INVALID", nil)
	setAuthClaims(c, testAdminClaims)

	h.List(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuditHandler_List_InvalidDateFrom(t *testing.T) {
	svc := newMockauditService(t)
	h := NewAuditHandler(svc, newTestLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/audit?date_from=not-a-date", nil)
	setAuthClaims(c, testAdminClaims)

	h.List(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuditHandler_ExportCSV_NoClaims(t *testing.T) {
	svc := newMockauditService(t)
	h := NewAuditHandler(svc, newTestLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/audit/export", nil)

	h.ExportCSV(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
