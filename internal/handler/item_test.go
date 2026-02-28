package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stpnv0/WarehouseControl/internal/handler/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	testAdminClaims = &domain.AuthClaims{
		UserID:   uuid.New(),
		Username: "admin",
		Role:     domain.RoleAdmin,
	}
	testViewerClaims = &domain.AuthClaims{
		UserID:   uuid.New(),
		Username: "viewer",
		Role:     domain.RoleViewer,
	}
)

func TestItemHandler_Create_Success(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	input := &domain.CreateItemInput{
		Name:     "Laptop",
		SKU:      "LAP-001",
		Quantity: 10,
		Price:    decimal.NewFromFloat(999.99),
	}

	expected := &domain.Item{
		ID:       uuid.New(),
		Name:     "Laptop",
		SKU:      "LAP-001",
		Quantity: 10,
		Price:    decimal.NewFromFloat(999.99),
	}

	svc.EXPECT().CreateItem(mock.Anything, testAdminClaims, input).Return(expected, nil)

	body, _ := json.Marshal(dto.CreateItemRequest{
		Name:     "Laptop",
		SKU:      "LAP-001",
		Quantity: 10,
		Price:    decimal.NewFromFloat(999.99),
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/items", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setAuthClaims(c, testAdminClaims)

	h.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp dto.ItemResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Laptop", resp.Name)
}

func TestItemHandler_Create_NoClaims(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/items", nil)

	h.Create(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestItemHandler_Create_InvalidBody(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/items", bytes.NewReader([]byte(`{}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	setAuthClaims(c, testAdminClaims)

	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_Create_Forbidden(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	svc.EXPECT().CreateItem(mock.Anything, testViewerClaims, mock.Anything).Return(nil, domain.ErrForbidden)

	body, _ := json.Marshal(dto.CreateItemRequest{
		Name:     "Laptop",
		SKU:      "LAP-001",
		Quantity: 10,
		Price:    decimal.NewFromInt(999),
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/items", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setAuthClaims(c, testViewerClaims)

	h.Create(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestItemHandler_List_Success(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	list := &domain.ItemList{
		Items: []*domain.Item{
			{ID: uuid.New(), Name: "Laptop", SKU: "LAP-001", Price: decimal.NewFromInt(999)},
		},
		Total:      1,
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}

	svc.EXPECT().ListItems(mock.Anything, testViewerClaims, mock.Anything, 0, 0).Return(list, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/items", nil)
	setAuthClaims(c, testViewerClaims)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ItemListResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Items, 1)
}

func TestItemHandler_GetByID_Success(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	itemID := uuid.New()
	expected := &domain.Item{ID: itemID, Name: "Laptop", SKU: "LAP-001", Price: decimal.NewFromInt(999)}

	svc.EXPECT().GetByID(mock.Anything, testViewerClaims, itemID).Return(expected, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/items/%s", itemID), nil)
	c.Params = gin.Params{{Key: "id", Value: itemID.String()}}
	setAuthClaims(c, testViewerClaims)

	h.GetByID(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestItemHandler_GetByID_InvalidID(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/items/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}
	setAuthClaims(c, testViewerClaims)

	h.GetByID(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_GetByID_NotFound(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	itemID := uuid.New()
	svc.EXPECT().GetByID(mock.Anything, testViewerClaims, itemID).Return(nil, domain.ErrNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/items/%s", itemID), nil)
	c.Params = gin.Params{{Key: "id", Value: itemID.String()}}
	setAuthClaims(c, testViewerClaims)

	h.GetByID(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestItemHandler_Update_Success(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	itemID := uuid.New()
	name := "Updated Laptop"
	expected := &domain.Item{ID: itemID, Name: "Updated Laptop", SKU: "LAP-001", Price: decimal.NewFromInt(999)}

	svc.EXPECT().Update(mock.Anything, testAdminClaims, itemID, &domain.UpdateItemInput{Name: &name}).Return(expected, nil)

	body, _ := json.Marshal(dto.UpdateItemRequest{Name: &name})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/items/%s", itemID), bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: itemID.String()}}
	setAuthClaims(c, testAdminClaims)

	h.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestItemHandler_Update_InvalidID(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/items/bad", bytes.NewReader([]byte(`{"name":"x"}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	setAuthClaims(c, testAdminClaims)

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_Delete_Success(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	itemID := uuid.New()
	svc.EXPECT().Delete(mock.Anything, testAdminClaims, itemID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/items/%s", itemID), nil)
	c.Params = gin.Params{{Key: "id", Value: itemID.String()}}
	setAuthClaims(c, testAdminClaims)

	h.Delete(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
}

func TestItemHandler_Delete_InvalidID(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/items/bad-id", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
	setAuthClaims(c, testAdminClaims)

	h.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_Delete_NotFound(t *testing.T) {
	svc := newMockitemService(t)
	h := NewItemHandler(svc, newTestLogger())

	itemID := uuid.New()
	svc.EXPECT().Delete(mock.Anything, testAdminClaims, itemID).Return(domain.ErrNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/items/%s", itemID), nil)
	c.Params = gin.Params{{Key: "id", Value: itemID.String()}}
	setAuthClaims(c, testAdminClaims)

	h.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
