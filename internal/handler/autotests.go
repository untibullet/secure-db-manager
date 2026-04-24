package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (h *Handlers) ListAutotests(c echo.Context) error {
	paging := parsePaging(c)
	filter := domain.Filter{}
	if v := c.QueryParam("is_active"); v != "" {
		filter["is_active"] = v == "true"
	}
	ctx := c.Request().Context()
	autotests, err := h.autotests.List(ctx, appRepo(c), filter, paging)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, page(autotests, paging))
}

func (h *Handlers) GetAutotest(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	at, err := h.autotests.GetByID(ctx, appRepo(c), id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, at)
}

func (h *Handlers) CreateAutotest(c echo.Context) error {
	var dto domain.AutotestDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	repo := appRepo(c)
	id, err := h.autotests.Create(ctx, repo, dto)
	if err != nil {
		return httpErr(err)
	}
	at, err := h.autotests.GetByID(ctx, repo, id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusCreated, at)
}

func (h *Handlers) UpdateAutotest(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var dto domain.AutotestDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	repo := appRepo(c)
	if err := h.autotests.Update(ctx, repo, id, dto); err != nil {
		return httpErr(err)
	}
	at, err := h.autotests.GetByID(ctx, repo, id)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusOK, at)
}

func (h *Handlers) DeleteAutotest(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := h.autotests.Delete(ctx, appRepo(c), id); err != nil {
		return httpErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) AddVersion(c echo.Context) error {
	id, err := pathInt(c, "id")
	if err != nil {
		return err
	}
	var dto domain.VersionDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()
	versionID, err := h.autotests.AddVersion(ctx, appRepo(c), id, dto)
	if err != nil {
		return httpErr(err)
	}
	return c.JSON(http.StatusCreated, domain.AutotestVersion{
		ID:            versionID,
		AutotestID:    id,
		VersionString: dto.VersionString,
		CommitHash:    sql.NullString{String: dto.CommitHash, Valid: dto.CommitHash != ""},
		CreatedAt:     time.Now().UTC(),
	})
}
