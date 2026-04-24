package service

import (
	"context"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

type TestCaseRepo interface {
	ListLeadAllCases(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.TestCase, error)
	GetTestCaseByID(ctx context.Context, id int) (*domain.TestCase, error)
	CreateTestCase(ctx context.Context, dto domain.TestCaseDTO) (int, error)
	UpdateTestCase(ctx context.Context, id int, dto domain.TestCaseDTO) error
	DeleteTestCase(ctx context.Context, id int) error
	ListTestCaseSteps(ctx context.Context, testCaseID int) ([]domain.Step, error)
	CreateTestCaseStep(ctx context.Context, tcID int, dto domain.StepDTO) (int, error)
	UpdateTestCaseStep(ctx context.Context, id int, dto domain.StepDTO) error
	DeleteTestCaseStep(ctx context.Context, id int) error
}

type TestCaseService struct{}

func (s *TestCaseService) List(ctx context.Context, repo TestCaseRepo, filter domain.Filter, paging domain.Paging) ([]domain.TestCase, error) {
	return repo.ListLeadAllCases(ctx, filter, paging)
}

func (s *TestCaseService) GetByID(ctx context.Context, repo TestCaseRepo, id int) (*domain.TestCase, error) {
	return repo.GetTestCaseByID(ctx, id)
}

func (s *TestCaseService) Create(ctx context.Context, repo TestCaseRepo, dto domain.TestCaseDTO) (int, error) {
	return repo.CreateTestCase(ctx, dto)
}

func (s *TestCaseService) Update(ctx context.Context, repo TestCaseRepo, id int, dto domain.TestCaseDTO) error {
	return repo.UpdateTestCase(ctx, id, dto)
}

func (s *TestCaseService) Delete(ctx context.Context, repo TestCaseRepo, id int) error {
	return repo.DeleteTestCase(ctx, id)
}

func (s *TestCaseService) ListSteps(ctx context.Context, repo TestCaseRepo, testCaseID int) ([]domain.Step, error) {
	return repo.ListTestCaseSteps(ctx, testCaseID)
}

func (s *TestCaseService) AddStep(ctx context.Context, repo TestCaseRepo, tcID int, dto domain.StepDTO) (int, error) {
	return repo.CreateTestCaseStep(ctx, tcID, dto)
}

func (s *TestCaseService) UpdateStep(ctx context.Context, repo TestCaseRepo, id int, dto domain.StepDTO) error {
	return repo.UpdateTestCaseStep(ctx, id, dto)
}

func (s *TestCaseService) DeleteStep(ctx context.Context, repo TestCaseRepo, id int) error {
	return repo.DeleteTestCaseStep(ctx, id)
}
