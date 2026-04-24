package service

import (
	"context"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

type RunRepo interface {
	ListActiveRuns(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.ActiveRun, error)
	GetRunByID(ctx context.Context, id int) (*domain.ActiveRun, error)
	CreateRun(ctx context.Context, dto domain.RunDTO) (int, error)
	UpdateRunStatus(ctx context.Context, id int, status string) error
	DeleteRun(ctx context.Context, id int) error
	ListRunItems(ctx context.Context, runID int) ([]domain.RunItem, error)
	AddRunItem(ctx context.Context, runID, testCaseID, order int) error
	RemoveRunItem(ctx context.Context, itemID int) error
}

type RunService struct{}

func (s *RunService) List(ctx context.Context, repo RunRepo, filter domain.Filter, paging domain.Paging) ([]domain.ActiveRun, error) {
	return repo.ListActiveRuns(ctx, filter, paging)
}

func (s *RunService) GetByID(ctx context.Context, repo RunRepo, id int) (*domain.ActiveRun, error) {
	return repo.GetRunByID(ctx, id)
}

func (s *RunService) Create(ctx context.Context, repo RunRepo, dto domain.RunDTO) (int, error) {
	return repo.CreateRun(ctx, dto)
}

func (s *RunService) UpdateStatus(ctx context.Context, repo RunRepo, id int, status string) error {
	return repo.UpdateRunStatus(ctx, id, status)
}

func (s *RunService) Delete(ctx context.Context, repo RunRepo, id int) error {
	return repo.DeleteRun(ctx, id)
}

func (s *RunService) ListItems(ctx context.Context, repo RunRepo, runID int) ([]domain.RunItem, error) {
	return repo.ListRunItems(ctx, runID)
}

func (s *RunService) AddItem(ctx context.Context, repo RunRepo, runID, testCaseID, order int) error {
	return repo.AddRunItem(ctx, runID, testCaseID, order)
}

func (s *RunService) DeleteItem(ctx context.Context, repo RunRepo, itemID int) error {
	return repo.RemoveRunItem(ctx, itemID)
}
