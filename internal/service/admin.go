package service

import (
	"context"
	"fmt"
	"time"

	"github.com/untibullet/secure-db-manager/internal/domain"
	"github.com/untibullet/secure-db-manager/internal/logger"
	"golang.org/x/crypto/bcrypt"
)

// AdminAppRepo — per-request операции через подключение администратора (AD-9).
type AdminAppRepo interface {
	AdminListUsersAndRoles(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.UserAdminView, error)
	AdminGetUserByID(ctx context.Context, id int) (*domain.UserAdminView, error)
	AdminSetUserActive(ctx context.Context, userID int, isActive bool) error
	AdminLockUntil(ctx context.Context, userID int, ts *time.Time) error
	AdminInsertUserRole(ctx context.Context, userID, roleID int, validUntil *time.Time) (username, dbRoleName string, err error)
	AdminDeleteUserRole(ctx context.Context, userID, roleID int) (username, dbRoleName string, err error)
	AdminListAuditLog(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.AuditEntry, error)
}

// AdminRoleRepo — DDL-операции через суперпользовательский пул (AD-9, AD-10).
type AdminRoleRepo interface {
	AdminCreateUser(ctx context.Context, dto domain.CreateUserDTO, bcryptHash string) (int, error)
	DropRole(ctx context.Context, username string) error
	GrantGroupRole(ctx context.Context, username, dbRoleName string) error
	RevokeGroupRole(ctx context.Context, username, dbRoleName string) error
	SetPasswordAndHash(ctx context.Context, userID int, username, bcryptHash, plainPassword string) error
	ListAvailableRoles(ctx context.Context) ([]domain.Role, error)
}

// AdminService синглтон: roleRepo инжектируется при старте (AD-9).
type AdminService struct {
	roles AdminRoleRepo
}

func NewAdminService(roles AdminRoleRepo) *AdminService {
	return &AdminService{roles: roles}
}

func (s *AdminService) ListUsers(ctx context.Context, repo AdminAppRepo, filter domain.Filter, paging domain.Paging) ([]domain.UserAdminView, error) {
	return repo.AdminListUsersAndRoles(ctx, filter, paging)
}

func (s *AdminService) GetUser(ctx context.Context, repo AdminAppRepo, id int) (*domain.UserAdminView, error) {
	return repo.AdminGetUserByID(ctx, id)
}

// CreateUser атомарно создаёт пользователя и его PostgreSQL login-роль (AD-10, AD-11).
func (s *AdminService) CreateUser(ctx context.Context, dto domain.CreateUserDTO) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: hash: %w", err)
	}

	userID, err := s.roles.AdminCreateUser(ctx, dto, string(hash))
	if err != nil {
		return 0, fmt.Errorf("CreateUser: %w", err)
	}

	logger.FromCtx(ctx).Info("admin: user created", "user_id", userID, "username", dto.Username)
	return userID, nil
}

func (s *AdminService) SetActive(ctx context.Context, repo AdminAppRepo, userID int, isActive bool) error {
	return repo.AdminSetUserActive(ctx, userID, isActive)
}

func (s *AdminService) SetLock(ctx context.Context, repo AdminAppRepo, userID int, until *time.Time) error {
	return repo.AdminLockUntil(ctx, userID, until)
}

// AssignRole назначает роль и выполняет GRANT групповой DB-роли (AD-7, AD-9).
func (s *AdminService) AssignRole(ctx context.Context, repo AdminAppRepo, userID, roleID int, validUntil *time.Time) error {
	username, dbRoleName, err := repo.AdminInsertUserRole(ctx, userID, roleID, validUntil)
	if err != nil {
		return fmt.Errorf("AssignRole: %w", err)
	}
	if err := s.roles.GrantGroupRole(ctx, username, dbRoleName); err != nil {
		return fmt.Errorf("AssignRole: grant: %w", err)
	}
	logger.FromCtx(ctx).Info("admin: role assigned", "user_id", userID, "db_role", dbRoleName)
	return nil
}

// RevokeRole снимает роль и выполняет REVOKE групповой DB-роли (AD-7, AD-9).
func (s *AdminService) RevokeRole(ctx context.Context, repo AdminAppRepo, userID, roleID int) error {
	username, dbRoleName, err := repo.AdminDeleteUserRole(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("RevokeRole: %w", err)
	}
	if err := s.roles.RevokeGroupRole(ctx, username, dbRoleName); err != nil {
		return fmt.Errorf("RevokeRole: revoke: %w", err)
	}
	logger.FromCtx(ctx).Info("admin: role revoked", "user_id", userID, "db_role", dbRoleName)
	return nil
}

func (s *AdminService) AuditLog(ctx context.Context, repo AdminAppRepo, filter domain.Filter, paging domain.Paging) ([]domain.AuditEntry, error) {
	return repo.AdminListAuditLog(ctx, filter, paging)
}

func (s *AdminService) ListRoles(ctx context.Context) ([]domain.Role, error) {
	return s.roles.ListAvailableRoles(ctx)
}
