package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (h *Handlers) ListTestPlans(c echo.Context) error {
	paging := parsePaging(c)
	filter := domain.Filter{}
	if s := c.QueryParam("status"); s != "" {
		filter["status"] = s
	}
	if p := c.QueryParam("priority"); p != "" {
		filter["priority"] = p
	}
	ctx := c.Request().Context()
	plans, err := h.plans.List(ctx, appRepo(c), filter, paging)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, page(plans, paging))
}

func (h *Handlers) GetTestPlan(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	plan, err := h.plans.GetByID(ctx, appRepo(c), id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, plan)
}

func (h *Handlers) CreateTestPlan(c echo.Context) error {
	var dto domain.TestPlanDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	repo := appRepo(c)
	id, err := h.plans.Create(ctx, repo, dto)
	if err != nil {
		return httpErr(err)
	}
	plan, err := h.plans.GetByID(ctx, repo, id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusCreated, plan)
}

func (h *Handlers) UpdateTestPlan(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var dto domain.TestPlanDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	repo := appRepo(c)
	if err := h.plans.Update(ctx, repo, id, dto); err != nil {
		return httpErr(err)
	}
	plan, err := h.plans.GetByID(ctx, repo, id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, plan)
}

func (h *Handlers) DeleteTestPlan(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := h.plans.Delete(ctx, appRepo(c), id); err != nil {
		return httpErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}
