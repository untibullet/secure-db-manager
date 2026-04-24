package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/untibullet/secure-db-manager/internal/domain"
	"github.com/untibullet/secure-db-manager/internal/middleware"
)

func (h *Handlers) ListRuns(c echo.Context) error {
	paging := parsePaging(c)
	filter := domain.Filter{}
	if s := c.QueryParam("status"); s != "" {
		filter["status"] = s
	}
	ctx := c.Request().Context()
	runs, err := h.runs.List(ctx, appRepo(c), filter, paging)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, page(runs, paging))
}

func (h *Handlers) GetRun(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	run, err := h.runs.GetByID(ctx, appRepo(c), id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, run)
}

type createRunReq struct {
	Name         string `json:"name"`
	PlanID       int    `json:"plan_id"`
	VersionID    int    `json:"version_id"`
	EnvID        int    `json:"env_id"`
	ToolConfigID int    `json:"tool_config_id"`
}

func (h *Handlers) CreateRun(c echo.Context) error {
	var req createRunReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	sess := middleware.SessionFromCtx(ctx)
	dto := domain.RunDTO{
		Name:         req.Name,
		PlanID:       req.PlanID,
		VersionID:    req.VersionID,
		EnvID:        req.EnvID,
		ToolConfigID: req.ToolConfigID,
		UserID:       sess.UserID,
	}
	repo := appRepo(c)
	id, err := h.runs.Create(ctx, repo, dto)
	if err != nil {
		return httpErr(err)
	}
	run, err := h.runs.GetByID(ctx, repo, id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusCreated, run)
}

func (h *Handlers) DeleteRun(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := h.runs.Delete(ctx, appRepo(c), id); err != nil {
		return httpErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}

type updateRunStatusReq struct {
	Status string `json:"status"`
}

func (h *Handlers) UpdateRunStatus(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var req updateRunStatusReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	repo := appRepo(c)
	if err := h.runs.UpdateStatus(ctx, repo, id, req.Status); err != nil {
		return httpErr(err)
	}
	run, err := h.runs.GetByID(ctx, repo, id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, run)
}

func (h *Handlers) ListRunItems(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	items, err := h.runs.ListItems(ctx, appRepo(c), id)
	if err != nil {
		return httpErr(err)
	}
	if items == nil {
		items = []domain.RunItem{}
	}
	return c.JSON(http.StatusOK, items)
}

type addRunItemReq struct {
	TestCaseID int `json:"test_case_id"`
	Order      int `json:"order"`
}

func (h *Handlers) AddRunItem(c echo.Context) error {
	runID, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var req addRunItemReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	if err := h.runs.AddItem(ctx, appRepo(c), runID, req.TestCaseID, req.Order); err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusCreated, domain.RunItem{
		RunID:     runID,
		CaseID:    req.TestCaseID,
		ExecOrder: req.Order,
	})
}

func (h *Handlers) DeleteRunItem(c echo.Context) error {
	if _, err := pathInt(c, "id"); err != nil {
		return err
	}
	itemID, err := pathInt(c, "itemId")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := h.runs.DeleteItem(ctx, appRepo(c), itemID); err != nil {
		return httpErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}
