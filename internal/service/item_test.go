package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	adminClaims = &domain.AuthClaims{
		UserID:   uuid.New(),
		Username: "admin",
		Role:     domain.RoleAdmin,
	}
	managerClaims = &domain.AuthClaims{
		UserID:   uuid.New(),
		Username: "manager",
		Role:     domain.RoleManager,
	}
	viewerClaims = &domain.AuthClaims{
		UserID:   uuid.New(),
		Username: "viewer",
		Role:     domain.RoleViewer,
	}
)

func newItemService(t *testing.T) (*ItemService, *mockitemRepository) {
	repo := newMockitemRepository(t)
	svc := NewItemService(repo, newTestLogger())
	return svc, repo
}

func TestItemService_CreateItem_Success(t *testing.T) {
	svc, repo := newItemService(t)

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

	repo.EXPECT().Create(mock.Anything, adminClaims.UserID, input).Return(expected, nil)

	result, err := svc.CreateItem(context.Background(), adminClaims, input)

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Name, result.Name)
}

func TestItemService_CreateItem_ManagerAllowed(t *testing.T) {
	svc, repo := newItemService(t)

	input := &domain.CreateItemInput{Name: "Mouse", SKU: "MOU-001", Quantity: 5, Price: decimal.NewFromInt(25)}
	expected := &domain.Item{ID: uuid.New(), Name: "Mouse", SKU: "MOU-001"}

	repo.EXPECT().Create(mock.Anything, managerClaims.UserID, input).Return(expected, nil)

	result, err := svc.CreateItem(context.Background(), managerClaims, input)

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
}

func TestItemService_CreateItem_ViewerForbidden(t *testing.T) {
	svc, _ := newItemService(t)

	input := &domain.CreateItemInput{Name: "Mouse", SKU: "MOU-001", Quantity: 5, Price: decimal.NewFromInt(25)}

	_, err := svc.CreateItem(context.Background(), viewerClaims, input)

	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestItemService_CreateItem_DuplicateSKU(t *testing.T) {
	svc, repo := newItemService(t)

	input := &domain.CreateItemInput{Name: "Laptop", SKU: "LAP-001", Quantity: 10, Price: decimal.NewFromInt(999)}

	repo.EXPECT().Create(mock.Anything, adminClaims.UserID, input).Return(nil, domain.ErrDuplicateSKU)

	_, err := svc.CreateItem(context.Background(), adminClaims, input)

	assert.ErrorIs(t, err, domain.ErrDuplicateSKU)
}

func TestItemService_CreateItem_RepoError(t *testing.T) {
	svc, repo := newItemService(t)

	input := &domain.CreateItemInput{Name: "Laptop", SKU: "LAP-001", Quantity: 10, Price: decimal.NewFromInt(999)}

	repo.EXPECT().Create(mock.Anything, adminClaims.UserID, input).Return(nil, errors.New("db error"))

	_, err := svc.CreateItem(context.Background(), adminClaims, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ItemService.Create")
}

func TestItemService_GetByID_Success(t *testing.T) {
	svc, repo := newItemService(t)

	itemID := uuid.New()
	expected := &domain.Item{ID: itemID, Name: "Laptop", SKU: "LAP-001"}

	repo.EXPECT().GetByID(mock.Anything, itemID).Return(expected, nil)

	result, err := svc.GetByID(context.Background(), viewerClaims, itemID)

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
}

func TestItemService_GetByID_NotFound(t *testing.T) {
	svc, repo := newItemService(t)

	itemID := uuid.New()
	repo.EXPECT().GetByID(mock.Anything, itemID).Return(nil, domain.ErrNotFound)

	_, err := svc.GetByID(context.Background(), viewerClaims, itemID)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestItemService_ListItems_Success(t *testing.T) {
	svc, repo := newItemService(t)

	items := []*domain.Item{
		{ID: uuid.New(), Name: "Laptop"},
		{ID: uuid.New(), Name: "Mouse"},
	}

	filter := &domain.ItemFilter{}
	repo.EXPECT().List(mock.Anything, filter, 20, 0).Return(items, int64(2), nil)

	result, err := svc.ListItems(context.Background(), viewerClaims, filter, 1, 20)

	assert.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
	assert.Equal(t, 1, result.TotalPages)
}

func TestItemService_ListItems_PaginationNormalization(t *testing.T) {
	svc, repo := newItemService(t)

	filter := &domain.ItemFilter{}
	// page=0 should normalize to 1, pageSize=0 should normalize to 20
	repo.EXPECT().List(mock.Anything, filter, 20, 0).Return([]*domain.Item{}, int64(0), nil)

	result, err := svc.ListItems(context.Background(), viewerClaims, filter, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
}

func TestItemService_ListItems_MaxPageSize(t *testing.T) {
	svc, repo := newItemService(t)

	filter := &domain.ItemFilter{}
	// pageSize=200 should normalize to 20 (default)
	repo.EXPECT().List(mock.Anything, filter, 20, 0).Return([]*domain.Item{}, int64(0), nil)

	_, err := svc.ListItems(context.Background(), viewerClaims, filter, 1, 200)

	assert.NoError(t, err)
}

func TestItemService_Update_Success(t *testing.T) {
	svc, repo := newItemService(t)

	itemID := uuid.New()
	name := "Updated Laptop"
	input := &domain.UpdateItemInput{Name: &name}
	expected := &domain.Item{ID: itemID, Name: "Updated Laptop", SKU: "LAP-001"}

	repo.EXPECT().Update(mock.Anything, adminClaims.UserID, itemID, input).Return(expected, nil)

	result, err := svc.Update(context.Background(), adminClaims, itemID, input)

	assert.NoError(t, err)
	assert.Equal(t, "Updated Laptop", result.Name)
}

func TestItemService_Update_ViewerForbidden(t *testing.T) {
	svc, _ := newItemService(t)

	name := "Updated"
	input := &domain.UpdateItemInput{Name: &name}

	_, err := svc.Update(context.Background(), viewerClaims, uuid.New(), input)

	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestItemService_Update_NoChanges(t *testing.T) {
	svc, _ := newItemService(t)

	input := &domain.UpdateItemInput{} // all nil

	_, err := svc.Update(context.Background(), adminClaims, uuid.New(), input)

	assert.ErrorIs(t, err, domain.ErrNoChanges)
}

func TestItemService_Update_NotFound(t *testing.T) {
	svc, repo := newItemService(t)

	itemID := uuid.New()
	name := "Updated"
	input := &domain.UpdateItemInput{Name: &name}

	repo.EXPECT().Update(mock.Anything, adminClaims.UserID, itemID, input).Return(nil, domain.ErrNotFound)

	_, err := svc.Update(context.Background(), adminClaims, itemID, input)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestItemService_Update_DuplicateSKU(t *testing.T) {
	svc, repo := newItemService(t)

	itemID := uuid.New()
	sku := "EXISTING-SKU"
	input := &domain.UpdateItemInput{SKU: &sku}

	repo.EXPECT().Update(mock.Anything, adminClaims.UserID, itemID, input).Return(nil, domain.ErrDuplicateSKU)

	_, err := svc.Update(context.Background(), adminClaims, itemID, input)

	assert.ErrorIs(t, err, domain.ErrDuplicateSKU)
}

func TestItemService_Delete_AdminSuccess(t *testing.T) {
	svc, repo := newItemService(t)

	itemID := uuid.New()
	repo.EXPECT().Delete(mock.Anything, adminClaims.UserID, itemID).Return(nil)

	err := svc.Delete(context.Background(), adminClaims, itemID)

	assert.NoError(t, err)
}

func TestItemService_Delete_ManagerForbidden(t *testing.T) {
	svc, _ := newItemService(t)

	err := svc.Delete(context.Background(), managerClaims, uuid.New())

	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestItemService_Delete_ViewerForbidden(t *testing.T) {
	svc, _ := newItemService(t)

	err := svc.Delete(context.Background(), viewerClaims, uuid.New())

	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestItemService_Delete_NotFound(t *testing.T) {
	svc, repo := newItemService(t)

	itemID := uuid.New()
	repo.EXPECT().Delete(mock.Anything, adminClaims.UserID, itemID).Return(domain.ErrNotFound)

	err := svc.Delete(context.Background(), adminClaims, itemID)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestNormalizePagination(t *testing.T) {
	tests := []struct {
		name             string
		page, pageSize   int
		wantP, wantPS    int
	}{
		{"default values", 0, 0, 1, 20},
		{"negative page", -1, 10, 1, 10},
		{"negative page size", 1, -1, 1, 20},
		{"over max page size", 1, 200, 1, 20},
		{"normal", 3, 50, 3, 50},
		{"max page size", 1, 100, 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, ps := normalizePagination(tt.page, tt.pageSize)
			assert.Equal(t, tt.wantP, p)
			assert.Equal(t, tt.wantPS, ps)
		})
	}
}

func TestCalcTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		pageSize int
		want     int
	}{
		{"zero total", 0, 20, 0},
		{"exact fit", 20, 20, 1},
		{"one extra", 21, 20, 2},
		{"zero page size", 10, 0, 0},
		{"large", 100, 30, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, calcTotalPages(tt.total, tt.pageSize))
		})
	}
}
