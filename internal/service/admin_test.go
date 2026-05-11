package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/untibullet/secure-db-manager/internal/domain"
	"github.com/untibullet/secure-db-manager/internal/service"
)

// MockAdminRoleRepo implements service.AdminRoleRepo.
type MockAdminRoleRepo struct{ mock.Mock }

func (m *MockAdminRoleRepo) AdminCreateUser(ctx context.Context, dto domain.CreateUserDTO, bcryptHash string) (int, error) {
	args := m.Called(ctx, dto, bcryptHash)
	return args.Int(0), args.Error(1)
}

func (m *MockAdminRoleRepo) DropRole(ctx context.Context, username string) error {
	return m.Called(ctx, username).Error(0)
}

func (m *MockAdminRoleRepo) GrantGroupRole(ctx context.Context, username, dbRoleName string) error {
	return m.Called(ctx, username, dbRoleName).Error(0)
}

func (m *MockAdminRoleRepo) RevokeGroupRole(ctx context.Context, username, dbRoleName string) error {
	return m.Called(ctx, username, dbRoleName).Error(0)
}

func (m *MockAdminRoleRepo) ListAvailableRoles(ctx context.Context) ([]domain.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Role), args.Error(1)
}

func (m *MockAdminRoleRepo) SetPasswordAndHash(ctx context.Context, userID int, username, bcryptHash, plainPassword string) error {
	return m.Called(ctx, userID, username, bcryptHash, plainPassword).Error(0)
}

// MockAdminAppRepo implements service.AdminAppRepo.
type MockAdminAppRepo struct{ mock.Mock }

func (m *MockAdminAppRepo) AdminListUsersAndRoles(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.UserAdminView, error) {
	args := m.Called(ctx, filter, paging)
	return args.Get(0).([]domain.UserAdminView), args.Error(1)
}

func (m *MockAdminAppRepo) AdminGetUserByID(ctx context.Context, id int) (*domain.UserAdminView, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserAdminView), args.Error(1)
}

func (m *MockAdminAppRepo) AdminSetUserActive(ctx context.Context, userID int, isActive bool) error {
	return m.Called(ctx, userID, isActive).Error(0)
}

func (m *MockAdminAppRepo) AdminLockUntil(ctx context.Context, userID int, ts *time.Time) error {
	return m.Called(ctx, userID, ts).Error(0)
}

func (m *MockAdminAppRepo) AdminInsertUserRole(ctx context.Context, userID, roleID int, validUntil *time.Time) (string, string, error) {
	args := m.Called(ctx, userID, roleID, validUntil)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAdminAppRepo) AdminDeleteUserRole(ctx context.Context, userID, roleID int) (string, string, error) {
	args := m.Called(ctx, userID, roleID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAdminAppRepo) AdminListAuditLog(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.AuditEntry, error) {
	args := m.Called(ctx, filter, paging)
	return args.Get(0).([]domain.AuditEntry), args.Error(1)
}


var (
	errInsertRole = errors.New("AdminInsertUserRole failed")
	errGrantRole  = errors.New("GrantGroupRole failed")
	errDeleteRole = errors.New("AdminDeleteUserRole failed")
	errRevokeRole = errors.New("RevokeGroupRole failed")
)

func TestAdminService_CreateUser(t *testing.T) {
	ctx := context.Background()
	dto := domain.CreateUserDTO{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secret",
	}

	// AdminCreateUser returns error → CreateUser propagates it.
	t.Run("repo_error_propagated", func(t *testing.T) {
		roles := new(MockAdminRoleRepo)
		roles.On("AdminCreateUser", mock.Anything, dto, mock.AnythingOfType("string")).
			Return(0, errors.New("db error"))

		svc := service.NewAdminService(roles)
		_, err := svc.CreateUser(ctx, dto)
		require.Error(t, err)
		roles.AssertExpectations(t)
	})
}

func TestAdminService_AssignRole(t *testing.T) {
	ctx := context.Background()

	// AdminInsertUserRole fails → GrantGroupRole must NOT be called.
	t.Run("insert_error_propagated", func(t *testing.T) {
		roles := new(MockAdminRoleRepo)
		appRepo := new(MockAdminAppRepo)
		appRepo.On("AdminInsertUserRole", ctx, 1, 2, (*time.Time)(nil)).
			Return("", "", errInsertRole)

		svc := service.NewAdminService(roles)
		err := svc.AssignRole(ctx, appRepo, 1, 2, nil)
		require.ErrorIs(t, err, errInsertRole)
		roles.AssertNotCalled(t, "GrantGroupRole")
		appRepo.AssertExpectations(t)
	})

	// Insert succeeds but GrantGroupRole fails → error propagated.
	t.Run("grant_error_propagated", func(t *testing.T) {
		roles := new(MockAdminRoleRepo)
		appRepo := new(MockAdminAppRepo)
		appRepo.On("AdminInsertUserRole", ctx, 1, 2, (*time.Time)(nil)).
			Return("alice", "db_tester", nil)
		roles.On("GrantGroupRole", ctx, "alice", "db_tester").Return(errGrantRole)

		svc := service.NewAdminService(roles)
		err := svc.AssignRole(ctx, appRepo, 1, 2, nil)
		require.ErrorIs(t, err, errGrantRole)
		appRepo.AssertExpectations(t)
		roles.AssertExpectations(t)
	})
}

func TestAdminService_RevokeRole(t *testing.T) {
	ctx := context.Background()

	// AdminDeleteUserRole fails → RevokeGroupRole must NOT be called.
	t.Run("delete_error_propagated", func(t *testing.T) {
		roles := new(MockAdminRoleRepo)
		appRepo := new(MockAdminAppRepo)
		appRepo.On("AdminDeleteUserRole", ctx, 1, 2).Return("", "", errDeleteRole)

		svc := service.NewAdminService(roles)
		err := svc.RevokeRole(ctx, appRepo, 1, 2)
		require.ErrorIs(t, err, errDeleteRole)
		roles.AssertNotCalled(t, "RevokeGroupRole")
		appRepo.AssertExpectations(t)
	})

	// Delete succeeds but RevokeGroupRole fails → error propagated.
	t.Run("revoke_error_propagated", func(t *testing.T) {
		roles := new(MockAdminRoleRepo)
		appRepo := new(MockAdminAppRepo)
		appRepo.On("AdminDeleteUserRole", ctx, 1, 2).Return("alice", "db_tester", nil)
		roles.On("RevokeGroupRole", ctx, "alice", "db_tester").Return(errRevokeRole)

		svc := service.NewAdminService(roles)
		err := svc.RevokeRole(ctx, appRepo, 1, 2)
		require.ErrorIs(t, err, errRevokeRole)
		appRepo.AssertExpectations(t)
		roles.AssertExpectations(t)
	})
}
