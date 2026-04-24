package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (h *Handlers) ListTestCases(c echo.Context) error {
	paging := parsePaging(c)
	filter := domain.Filter{}
	if v := c.QueryParam("is_active"); v != "" {
		filter["is_active"] = v == "true"
	}
	if v := c.QueryParam("is_automated"); v != "" {
		filter["is_automated"] = v == "true"
	}
	ctx := c.Request().Context()
	cases, err := h.cases.List(ctx, appRepo(c), filter, paging)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, page(cases, paging))
}

func (h *Handlers) GetTestCase(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	tc, err := h.cases.GetByID(ctx, appRepo(c), id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, tc)
}

func (h *Handlers) CreateTestCase(c echo.Context) error {
	var dto domain.TestCaseDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	repo := appRepo(c)
	id, err := h.cases.Create(ctx, repo, dto)
	if err != nil {
		return httpErr(err)
	}
	tc, err := h.cases.GetByID(ctx, repo, id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusCreated, tc)
}

func (h *Handlers) UpdateTestCase(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var dto domain.TestCaseDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	repo := appRepo(c)
	if err := h.cases.Update(ctx, repo, id, dto); err != nil {
		return httpErr(err)
	}
	tc, err := h.cases.GetByID(ctx, repo, id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, tc)
}

func (h *Handlers) DeleteTestCase(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := h.cases.Delete(ctx, appRepo(c), id); err != nil {
		return httpErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) ListSteps(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	steps, err := h.cases.ListSteps(ctx, appRepo(c), id)
	if err != nil {
		return httpErr(err)
	}
	if steps == nil {
		steps = []domain.Step{}
	}
	return c.JSON(http.StatusOK, steps)
}

func (h *Handlers) AddStep(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var dto domain.StepDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	stepID, err := h.cases.AddStep(ctx, appRepo(c), id, dto)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusCreated, domain.Step{
		ID:       stepID,
		CaseID:   id,
		Order:    dto.Order,
		Action:   dto.Action,
		Expected: dto.Expected,
	})
}

func (h *Handlers) UpdateStep(c echo.Context) error {
	caseID, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	stepID, err := pathInt(c, "stepId")
	if err != nil {
		return err
	}
	var dto domain.StepDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	if err := h.cases.UpdateStep(ctx, appRepo(c), stepID, dto); err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, domain.Step{
		ID:       stepID,
		CaseID:   caseID,
		Order:    dto.Order,
		Action:   dto.Action,
		Expected: dto.Expected,
	})
}

func (h *Handlers) DeleteStep(c echo.Context) error {
	if _, err := pathInt(c, "id"); err != nil {
		return err
	}
	stepID, err := pathInt(c, "stepId")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := h.cases.DeleteStep(ctx, appRepo(c), stepID); err != nil {
		return httpErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}
