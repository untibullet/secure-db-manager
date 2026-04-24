package service

import (
	"context"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

type ReferenceRepo interface {
	ListSharedEnvironments(ctx context.Context, filter domain.Filter) ([]domain.Environment, error)
	ListSharedReports(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.Report, error)
}

type ReferenceService struct{}

func (s *ReferenceService) ListEnvironments(ctx context.Context, repo ReferenceRepo, filter domain.Filter) ([]domain.Environment, error) {
	return repo.ListSharedEnvironments(ctx, filter)
}

func (s *ReferenceService) ListReports(ctx context.Context, repo ReferenceRepo, filter domain.Filter, paging domain.Paging) ([]domain.Report, error) {
	return repo.ListSharedReports(ctx, filter, paging)
}
