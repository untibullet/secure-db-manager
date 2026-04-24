package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/untibullet/secure-db-manager/internal/domain"
	"github.com/untibullet/secure-db-manager/internal/service"
)

// MockResultRepo implements service.ResultRepo.
// Only methods under test call m.Called(); the rest return zero values so they
// won't panic if unexpectedly called but also won't satisfy AssertExpectations.
type MockResultRepo struct{ mock.Mock }

func (m *MockResultRepo) ListLeadAllResults(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.TestResult, error) {
	args := m.Called(ctx, filter, paging)
	return args.Get(0).([]domain.TestResult), args.Error(1)
}

func (m *MockResultRepo) ListTesterResults(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.PublicResult, error) {
	args := m.Called(ctx, filter, paging)
	return args.Get(0).([]domain.PublicResult), args.Error(1)
}

func (m *MockResultRepo) GetLeadResultByID(ctx context.Context, id int) (*domain.TestResult, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TestResult), args.Error(1)
}

func (m *MockResultRepo) GetTesterResultByID(ctx context.Context, id int) (*domain.PublicResult, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PublicResult), args.Error(1)
}

func (m *MockResultRepo) CreateResult(ctx context.Context, dto domain.ResultDTO) (int, error) {
	args := m.Called(ctx, dto)
	return args.Int(0), args.Error(1)
}

func (m *MockResultRepo) UpdateResult(ctx context.Context, id int, dto domain.ResultDTO) error {
	return m.Called(ctx, id, dto).Error(0)
}

func (m *MockResultRepo) DeleteResult(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockResultRepo) ListResultArtifacts(ctx context.Context, resultID int) ([]domain.Artifact, error) {
	args := m.Called(ctx, resultID)
	return args.Get(0).([]domain.Artifact), args.Error(1)
}

func (m *MockResultRepo) AddResultArtifact(ctx context.Context, resultID int, dto domain.ArtifactDTO) (int, error) {
	args := m.Called(ctx, resultID, dto)
	return args.Int(0), args.Error(1)
}

func TestResultService_ListByRole(t *testing.T) {
	svc := new(service.ResultService)
	ctx := context.Background()
	filter := domain.Filter{}
	paging := domain.Paging{}

	// dbRole="TESTER" → must route to ListTesterResults (v_tester_my_results, RLS by CURRENT_USER)
	t.Run("tester_calls_tester_results", func(t *testing.T) {
		repo := new(MockResultRepo)
		repo.On("ListTesterResults", ctx, filter, paging).
			Return([]domain.PublicResult{{ResultID: 1}}, nil)

		result, err := svc.ListByRole(ctx, repo, filter, paging, "TESTER")
		require.NoError(t, err)

		results, ok := result.([]domain.PublicResult)
		require.True(t, ok, "expected []domain.PublicResult for TESTER role")
		assert.Len(t, results, 1)
		repo.AssertExpectations(t)
	})

	// dbRole != "TESTER" → must route to ListLeadAllResults
	t.Run("non_tester_calls_lead_results", func(t *testing.T) {
		repo := new(MockResultRepo)
		repo.On("ListLeadAllResults", ctx, filter, paging).
			Return([]domain.TestResult{{ResultID: 2}}, nil)

		result, err := svc.ListByRole(ctx, repo, filter, paging, "TEST_LEAD")
		require.NoError(t, err)

		results, ok := result.([]domain.TestResult)
		require.True(t, ok, "expected []domain.TestResult for TEST_LEAD role")
		assert.Len(t, results, 1)
		repo.AssertExpectations(t)
	})
}

func TestResultService_GetByIDForRole(t *testing.T) {
	svc := new(service.ResultService)
	ctx := context.Background()

	// dbRole="TESTER" → must route to GetTesterResultByID (PublicResult)
	t.Run("tester_calls_tester_view", func(t *testing.T) {
		repo := new(MockResultRepo)
		repo.On("GetTesterResultByID", ctx, 42).
			Return(&domain.PublicResult{ResultID: 42}, nil)

		result, err := svc.GetByIDForRole(ctx, repo, 42, "TESTER")
		require.NoError(t, err)

		res, ok := result.(*domain.PublicResult)
		require.True(t, ok, "expected *domain.PublicResult for TESTER role")
		assert.Equal(t, 42, res.ResultID)
		repo.AssertExpectations(t)
	})

	// dbRole != "TESTER" → must route to GetLeadResultByID (TestResult)
	t.Run("lead_calls_lead_view", func(t *testing.T) {
		repo := new(MockResultRepo)
		repo.On("GetLeadResultByID", ctx, 99).
			Return(&domain.TestResult{ResultID: 99}, nil)

		result, err := svc.GetByIDForRole(ctx, repo, 99, "TEST_LEAD")
		require.NoError(t, err)

		res, ok := result.(*domain.TestResult)
		require.True(t, ok, "expected *domain.TestResult for TEST_LEAD role")
		assert.Equal(t, 99, res.ResultID)
		repo.AssertExpectations(t)
	})
}
