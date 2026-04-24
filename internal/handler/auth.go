package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/untibullet/secure-db-manager/internal/middleware"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handlers) Login(c echo.Context) error {
	var req loginReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Username == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "username and password are required")
	}
	ctx := c.Request().Context()
	// roleRepo реализует AuthRepo: GetUserForLogin читает через суперпользовательский пул (AD-11).
	token, err := h.auth.Login(ctx, h.roleRepo, req.Username, req.Password)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": time.Now().Add(h.sessionTTL).UTC(),
	})
}

func (h *Handlers) Logout(c echo.Context) error {
	ctx := c.Request().Context()
	sess := middleware.SessionFromCtx(ctx)
	h.auth.Logout(ctx, sess.UserID)
	return c.NoContent(http.StatusNoContent)
}
