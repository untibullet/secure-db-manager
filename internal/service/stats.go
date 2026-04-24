package service

import (
	"context"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

type StatsRepo interface {
	ListRunSummaries(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.RunSummary, error)
	ListTestCaseStatistics(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.CaseStats, error)
}

type StatsService struct{}

func (s *StatsService) RunSummaries(ctx context.Context, repo StatsRepo, filter domain.Filter, paging domain.Paging) ([]domain.RunSummary, error) {
	return repo.ListRunSummaries(ctx, filter, paging)
}

func (s *StatsService) CaseStats(ctx context.Context, repo StatsRepo, filter domain.Filter, paging domain.Paging) ([]domain.CaseStats, error) {
	return repo.ListTestCaseStatistics(ctx, filter, paging)
}
