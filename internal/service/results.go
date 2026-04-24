package service

import (
	"context"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

type ResultRepo interface {
	ListLeadAllResults(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.TestResult, error)
	ListTesterResults(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.PublicResult, error)
	GetLeadResultByID(ctx context.Context, id int) (*domain.TestResult, error)
	GetTesterResultByID(ctx context.Context, id int) (*domain.PublicResult, error)
	CreateResult(ctx context.Context, dto domain.ResultDTO) (int, error)
	UpdateResult(ctx context.Context, id int, dto domain.ResultDTO) error
	DeleteResult(ctx context.Context, id int) error
	ListResultArtifacts(ctx context.Context, resultID int) ([]domain.Artifact, error)
	AddResultArtifact(ctx context.Context, resultID int, dto domain.ArtifactDTO) (int, error)
}

type ResultService struct{}

// ListByRole возвращает результаты с учётом прав доступа по роли (AD-4).
// TESTER видит только собственные данные через v_tester_my_results (RLS по CURRENT_USER).
func (s *ResultService) ListByRole(ctx context.Context, repo ResultRepo, filter domain.Filter, paging domain.Paging, dbRole string) (any, error) {
	if dbRole == "TESTER" {
		return repo.ListTesterResults(ctx, filter, paging)
	}
	return repo.ListLeadAllResults(ctx, filter, paging)
}

// GetByIDForRole выбирает view по роли (AD-4).
func (s *ResultService) GetByIDForRole(ctx context.Context, repo ResultRepo, id int, dbRole string) (any, error) {
	if dbRole == "TESTER" {
		return repo.GetTesterResultByID(ctx, id)
	}
	return repo.GetLeadResultByID(ctx, id)
}

func (s *ResultService) Create(ctx context.Context, repo ResultRepo, dto domain.ResultDTO) (int, error) {
	return repo.CreateResult(ctx, dto)
}

func (s *ResultService) Update(ctx context.Context, repo ResultRepo, id int, dto domain.ResultDTO) error {
	return repo.UpdateResult(ctx, id, dto)
}

func (s *ResultService) Delete(ctx context.Context, repo ResultRepo, id int) error {
	return repo.DeleteResult(ctx, id)
}

func (s *ResultService) ListArtifacts(ctx context.Context, repo ResultRepo, resultID int) ([]domain.Artifact, error) {
	return repo.ListResultArtifacts(ctx, resultID)
}

func (s *ResultService) AddArtifact(ctx context.Context, repo ResultRepo, resultID int, dto domain.ArtifactDTO) (int, error) {
	return repo.AddResultArtifact(ctx, resultID, dto)
}
