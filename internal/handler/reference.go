package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (h *Handlers) ListEnvironments(c echo.Context) error {
	ctx := c.Request().Context()
	envs, err := h.reference.ListEnvironments(ctx, appRepo(c), domain.Filter{})
	if err != nil {
		return httpErr(err)
	}
	if envs == nil {
		envs = []domain.Environment{}
	}
	return c.JSON(http.StatusOK, envs)
}

func (h *Handlers) ListReports(c echo.Context) error {
	paging := parsePaging(c)
	ctx := c.Request().Context()
	reports, err := h.reference.ListReports(ctx, appRepo(c), domain.Filter{}, paging)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, page(reports, paging))
}
