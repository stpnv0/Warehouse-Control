package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/wb-go/wbf/logger"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

type itemRepository interface {
	Create(ctx context.Context, userID uuid.UUID, input *domain.CreateItemInput) (*domain.Item, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Item, error)
	List(ctx context.Context, filter *domain.ItemFilter, limit, offset int) ([]*domain.Item, int64, error)
	Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input *domain.UpdateItemInput) (*domain.Item, error)
	Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
}
type ItemService struct {
	itemRepo itemRepository
	log      logger.Logger
}

func NewItemService(itemRepo itemRepository, log logger.Logger) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
		log:      log.With("component", "ItemService"),
	}
}

func (s *ItemService) CreateItem(
	ctx context.Context,
	claims *domain.AuthClaims,
	input *domain.CreateItemInput,
) (*domain.Item, error) {
	const op = "ItemService.Create"

	if !claims.Role.CanCreate() {
		return nil, domain.ErrForbidden
	}

	item, err := s.itemRepo.Create(ctx, claims.UserID, input)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateSKU) {
			return nil, domain.ErrDuplicateSKU
		}
		s.log.Ctx(ctx).Error("failed to create item",
			"error", err,
			"user_id", claims.UserID,
			"sku", input.SKU,
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return item, nil
}

func (s *ItemService) GetByID(ctx context.Context, claims *domain.AuthClaims, id uuid.UUID) (*domain.Item, error) {
	const op = "ItemService.GetByID"

	if !claims.Role.CanView() {
		return nil, domain.ErrForbidden
	}

	item, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
		s.log.Ctx(ctx).Error("failed to get item",
			"error", err,
			"item_id", id,
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return item, nil
}

func (s *ItemService) ListItems(
	ctx context.Context,
	claims *domain.AuthClaims,
	filter *domain.ItemFilter,
	page, pageSize int,
) (*domain.ItemList, error) {
	const op = "ItemService.ListItems"

	if !claims.Role.CanView() {
		return nil, domain.ErrForbidden
	}

	page, pageSize = normalizePagination(page, pageSize)
	offset := (page - 1) * pageSize

	items, total, err := s.itemRepo.List(ctx, filter, pageSize, offset)
	if err != nil {
		s.log.Ctx(ctx).Error("failed to list items",
			"error", err,
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.ItemList{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: calcTotalPages(total, pageSize),
	}, nil
}

func (s *ItemService) Update(
	ctx context.Context,
	claims *domain.AuthClaims,
	id uuid.UUID,
	input *domain.UpdateItemInput,
) (*domain.Item, error) {
	const op = "ItemService.Update"

	if !claims.Role.CanUpdate() {
		return nil, domain.ErrForbidden
	}

	if !input.HasChanges() {
		return nil, domain.ErrNoChanges
	}

	item, err := s.itemRepo.Update(ctx, claims.UserID, id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		if errors.Is(err, domain.ErrDuplicateSKU) {
			return nil, domain.ErrDuplicateSKU
		}
		s.log.Ctx(ctx).Error("failed to update item",
			"error", err,
			"item_id", id,
			"user_id", claims.UserID,
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return item, nil
}

func (s *ItemService) Delete(
	ctx context.Context,
	claims *domain.AuthClaims,
	id uuid.UUID,
) error {
	const op = "ItemService.Delete"

	if !claims.Role.CanDelete() {
		return domain.ErrForbidden
	}

	err := s.itemRepo.Delete(ctx, claims.UserID, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		s.log.Ctx(ctx).Error("failed to delete item",
			"error", err,
			"item_id", id,
			"user_id", claims.UserID,
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = defaultPage
	}
	if pageSize < 1 || pageSize > maxPageSize {
		pageSize = defaultPageSize
	}
	return page, pageSize
}

func calcTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	pages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		pages++
	}
	return pages
}
