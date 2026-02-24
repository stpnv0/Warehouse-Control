package service

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stpnv0/WarehouseControl/internal/export"
	"github.com/wb-go/wbf/logger"
)

const maxExportRows = 10000

type auditRepository interface {
	GetByItemID(ctx context.Context, itemID uuid.UUID) ([]*domain.AuditEntryWithUser, error)
	List(ctx context.Context, filter *domain.AuditFilter, limit, offset int) ([]*domain.AuditEntryWithUser, int64, error)
}
type AuditService struct {
	auditRepo auditRepository
	log       logger.Logger
}

func NewAuditService(auditRepo auditRepository, log logger.Logger) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
		log:       log.With("component", "AuditService"),
	}
}

func (s *AuditService) GetByItemID(
	ctx context.Context,
	claims *domain.AuthClaims,
	itemID uuid.UUID,
) ([]*domain.AuditEntryWithUser, error) {
	const op = "AuditService.GetByItemID"

	if !claims.Role.CanViewAudit() {
		return nil, domain.ErrForbidden
	}

	audit, err := s.auditRepo.GetByItemID(ctx, itemID)
	if err != nil {
		s.log.Ctx(ctx).Error("failed to get audit by item",
			"error", err,
			"item_id", itemID,
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if audit == nil {
		audit = []*domain.AuditEntryWithUser{}
	}

	return audit, nil
}

func (s *AuditService) List(
	ctx context.Context,
	claims *domain.AuthClaims,
	filter *domain.AuditFilter,
	page, pageSize int,
) (*domain.AuditList, error) {
	const op = "AuditService.List"

	if !claims.Role.CanViewAudit() {
		return nil, domain.ErrForbidden
	}

	if filter.Action != nil && !filter.Action.IsValid() {
		return nil, domain.ErrValidation
	}

	page, pageSize = normalizePagination(page, pageSize)
	offset := (page - 1) * pageSize

	audit, total, err := s.auditRepo.List(ctx, filter, pageSize, offset)
	if err != nil {
		s.log.Ctx(ctx).Error("failed to list audit",
			"error", err,
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if audit == nil {
		audit = []*domain.AuditEntryWithUser{}
	}

	return &domain.AuditList{
		Entries:    audit,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: calcTotalPages(total, pageSize),
	}, nil
}

func (s *AuditService) ExportCSV(
	ctx context.Context,
	claims *domain.AuthClaims,
	filter *domain.AuditFilter,
	w io.Writer,
) error {
	const op = "AuditService.ExportCSV"

	if !claims.Role.CanExport() {
		return domain.ErrForbidden
	}

	entries, _, err := s.auditRepo.List(ctx, filter, maxExportRows, 0)
	if err != nil {
		s.log.Ctx(ctx).Error("failed to fetch audit for CSV export",
			"error", err,
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	return export.WriteAuditCSV(w, entries)
}
