package service

import (
	"context"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

type TestPlanRepo interface {
	ListPublicTestPlans(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.TestPlan, error)
	GetTestPlanByID(ctx context.Context, id int) (*domain.TestPlan, error)
	CreateTestPlan(ctx context.Context, dto domain.TestPlanDTO) (int, error)
	UpdateTestPlan(ctx context.Context, id int, dto domain.TestPlanDTO) error
	DeleteTestPlan(ctx context.Context, id int) error
}

type TestPlanService struct{}

func (s *TestPlanService) List(ctx context.Context, repo TestPlanRepo, filter domain.Filter, paging domain.Paging) ([]domain.TestPlan, error) {
	return repo.ListPublicTestPlans(ctx, filter, paging)
}

func (s *TestPlanService) GetByID(ctx context.Context, repo TestPlanRepo, id int) (*domain.TestPlan, error) {
	return repo.GetTestPlanByID(ctx, id)
}

func (s *TestPlanService) Create(ctx context.Context, repo TestPlanRepo, dto domain.TestPlanDTO) (int, error) {
	return repo.CreateTestPlan(ctx, dto)
}

func (s *TestPlanService) Update(ctx context.Context, repo TestPlanRepo, id int, dto domain.TestPlanDTO) error {
	return repo.UpdateTestPlan(ctx, id, dto)
}

func (s *TestPlanService) Delete(ctx context.Context, repo TestPlanRepo, id int) error {
	return repo.DeleteTestPlan(ctx, id)
}
