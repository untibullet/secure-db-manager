package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (h *Handlers) ListUsers(c echo.Context) error {
	paging := parsePaging(c)
	filter := domain.Filter{}
	if v := c.QueryParam("is_active"); v != "" {
		filter["is_active"] = v == "true"
	}
	ctx := c.Request().Context()
	users, err := h.admin.ListUsers(ctx, appRepo(c), filter, paging)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, page(users, paging))
}

func (h *Handlers) GetUser(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	user, err := h.admin.GetUser(ctx, appRepo(c), id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, user)
}

type createUserReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	RoleID   int    `json:"role_id"`
}

// CreateUser создаёт пользователя + PG login-роль, затем назначает начальную групповую роль (AD-10).
func (h *Handlers) CreateUser(c echo.Context) error {
	var req createUserReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	dto := domain.CreateUserDTO{
		Username: req.Username,
		Password: req.Password,
		FullName: req.FullName,
		Email:    req.Email,
	}
	userID, err := h.admin.CreateUser(ctx, dto)
	if err != nil {
		return httpErr(err)
	}
	repo := appRepo(c)
	if err := h.admin.AssignRole(ctx, repo, userID, req.RoleID, nil); err != nil {
		return httpErr(err)
	}
	user, err := h.admin.GetUser(ctx, repo, userID)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusCreated, user)
}

func (h *Handlers) SetActive(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	repo := appRepo(c)
	if err := h.admin.SetActive(ctx, repo, id, req.IsActive); err != nil {
		return httpErr(err)
	}
	user, err := h.admin.GetUser(ctx, repo, id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handlers) SetLock(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var req struct {
		LockedUntil *time.Time `json:"locked_until"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	repo := appRepo(c)
	if err := h.admin.SetLock(ctx, repo, id, req.LockedUntil); err != nil {
		return httpErr(err)
	}
	user, err := h.admin.GetUser(ctx, repo, id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handlers) AssignRole(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var req struct {
		RoleID     int        `json:"role_id"`
		ValidUntil *time.Time `json:"valid_until"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	if err := h.admin.AssignRole(ctx, appRepo(c), id, req.RoleID, req.ValidUntil); err != nil {
		return httpErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) RevokeRole(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	roleID, err := pathInt(c, "roleId")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := h.admin.RevokeRole(ctx, appRepo(c), id, roleID); err != nil {
		return httpErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) ListRoles(c echo.Context) error {
	ctx := c.Request().Context()
	roles, err := h.admin.ListRoles(ctx)
	if err != nil {
		return httpErr(err)
	}
	if roles == nil {
		roles = []domain.Role{}
	}
	return c.JSON(http.StatusOK, roles)
}

func (h *Handlers) AuditLog(c echo.Context) error {
	paging := parsePaging(c)
	filter := domain.Filter{}
	if v := c.QueryParam("table_name"); v != "" {
		filter["table_name"] = v
	}
	if v := c.QueryParam("operation"); v != "" {
		filter["operation"] = v
	}
	ctx := c.Request().Context()
	entries, err := h.admin.AuditLog(ctx, appRepo(c), filter, paging)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, page(entries, paging))
}
