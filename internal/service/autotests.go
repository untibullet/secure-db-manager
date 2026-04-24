package service

import (
	"context"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

type AutotestRepo interface {
	ListEngineerAutotests(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.Autotest, error)
	GetAutotestByID(ctx context.Context, id int) (*domain.Autotest, error)
	CreateAutotest(ctx context.Context, dto domain.AutotestDTO) (int, error)
	UpdateAutotest(ctx context.Context, id int, dto domain.AutotestDTO) error
	DeleteAutotest(ctx context.Context, id int) error
	ListAutotestVersions(ctx context.Context, autotestID int) ([]domain.AutotestVersion, error)
	CreateAutotestVersion(ctx context.Context, autotestID int, dto domain.VersionDTO) (int, error)
}

type AutotestService struct{}

func (s *AutotestService) List(ctx context.Context, repo AutotestRepo, filter domain.Filter, paging domain.Paging) ([]domain.Autotest, error) {
	return repo.ListEngineerAutotests(ctx, filter, paging)
}

func (s *AutotestService) GetByID(ctx context.Context, repo AutotestRepo, id int) (*domain.Autotest, error) {
	return repo.GetAutotestByID(ctx, id)
}

func (s *AutotestService) Create(ctx context.Context, repo AutotestRepo, dto domain.AutotestDTO) (int, error) {
	return repo.CreateAutotest(ctx, dto)
}

func (s *AutotestService) Update(ctx context.Context, repo AutotestRepo, id int, dto domain.AutotestDTO) error {
	return repo.UpdateAutotest(ctx, id, dto)
}

func (s *AutotestService) Delete(ctx context.Context, repo AutotestRepo, id int) error {
	return repo.DeleteAutotest(ctx, id)
}

func (s *AutotestService) ListVersions(ctx context.Context, repo AutotestRepo, autotestID int) ([]domain.AutotestVersion, error) {
	return repo.ListAutotestVersions(ctx, autotestID)
}

func (s *AutotestService) AddVersion(ctx context.Context, repo AutotestRepo, autotestID int, dto domain.VersionDTO) (int, error) {
	return repo.CreateAutotestVersion(ctx, autotestID, dto)
}
