package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (h *Handlers) RunSummaries(c echo.Context) error {
	paging := parsePaging(c)
	ctx := c.Request().Context()
	summaries, err := h.stats.RunSummaries(ctx, appRepo(c), domain.Filter{}, paging)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, page(summaries, paging))
}

func (h *Handlers) CaseStats(c echo.Context) error {
	paging := parsePaging(c)
	ctx := c.Request().Context()
	stats, err := h.stats.CaseStats(ctx, appRepo(c), domain.Filter{}, paging)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, page(stats, paging))
}
