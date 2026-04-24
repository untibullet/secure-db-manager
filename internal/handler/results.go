package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/untibullet/secure-db-manager/internal/domain"
	"github.com/untibullet/secure-db-manager/internal/middleware"
)

func (h *Handlers) ListResults(c echo.Context) error {
	paging := parsePaging(c)
	filter := domain.Filter{}
	if s := c.QueryParam("status"); s != "" {
		filter["status"] = s
	}
	ctx := c.Request().Context()
	sess := middleware.SessionFromCtx(ctx)
	data, err := h.results.ListByRole(ctx, appRepo(c), filter, paging, sess.DBRole)
	if err != nil {
		return httpErr(err)
	}
	switch v := data.(type) {
	case []domain.TestResult:
		return c.JSON(http.StatusOK, page(v, paging))
	case []domain.PublicResult:
		return c.JSON(http.StatusOK, page(v, paging))
	default:
		return c.JSON(http.StatusOK, pageResp[any]{
			Data: nil,
			Meta: domain.PageMeta{Total: 0, Limit: paging.Limit, Offset: paging.Offset},
		})
	}
}

func (h *Handlers) GetResult(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	sess := middleware.SessionFromCtx(ctx)
	data, err := h.results.GetByIDForRole(ctx, appRepo(c), id, sess.DBRole)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handlers) CreateResult(c echo.Context) error {
	var dto domain.ResultDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	sess := middleware.SessionFromCtx(ctx)
	repo := appRepo(c)
	id, err := h.results.Create(ctx, repo, dto)
	if err != nil {
		return httpErr(err)
	}
	result, err := h.results.GetByIDForRole(ctx, repo, id, sess.DBRole)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *Handlers) UpdateResult(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var req struct {
		StatusID int    `json:"status_id"`
		Summary  string `json:"summary"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	sess := middleware.SessionFromCtx(ctx)
	repo := appRepo(c)
	dto := domain.ResultDTO{StatusID: req.StatusID, Summary: req.Summary}
	if err := h.results.Update(ctx, repo, id, dto); err != nil {
		return httpErr(err)
	}
	result, err := h.results.GetByIDForRole(ctx, repo, id, sess.DBRole)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handlers) DeleteResult(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := h.results.Delete(ctx, appRepo(c), id); err != nil {
		return httpErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) ListArtifacts(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	artifacts, err := h.results.ListArtifacts(ctx, appRepo(c), id)
	if err != nil {
		return httpErr(err)
	}
	if artifacts == nil {
		artifacts = []domain.Artifact{}
	}
	return c.JSON(http.StatusOK, artifacts)
}

func (h *Handlers) AddArtifact(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var dto domain.ArtifactDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	artID, err := h.results.AddArtifact(ctx, appRepo(c), id, dto)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusCreated, domain.Artifact{
		ID:        artID,
		ResultID:  id,
		Kind:      dto.Kind,
		FilePath:  dto.Path,
		CreatedAt: time.Now().UTC(),
	})
}
