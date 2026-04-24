package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/untibullet/secure-db-manager/internal/domain"
	"github.com/untibullet/secure-db-manager/internal/middleware"
	"github.com/untibullet/secure-db-manager/internal/repository"
	"github.com/untibullet/secure-db-manager/internal/service"
)

type Handlers struct {
	auth       *service.AuthService
	plans      *service.TestPlanService
	cases      *service.TestCaseService
	runs       *service.RunService
	results    *service.ResultService
	autotests  *service.AutotestService
	reference  *service.ReferenceService
	stats      *service.StatsService
	admin      *service.AdminService
	roleRepo   *repository.RoleRepository
	sessionTTL time.Duration
}

func New(
	auth *service.AuthService,
	plans *service.TestPlanService,
	cases *service.TestCaseService,
	runs *service.RunService,
	results *service.ResultService,
	autotests *service.AutotestService,
	reference *service.ReferenceService,
	stats *service.StatsService,
	admin *service.AdminService,
	roleRepo *repository.RoleRepository,
	sessionTTL time.Duration,
) *Handlers {
	return &Handlers{
		auth: auth, plans: plans, cases: cases, runs: runs,
		results: results, autotests: autotests, reference: reference,
		stats: stats, admin: admin, roleRepo: roleRepo,
		sessionTTL: sessionTTL,
	}
}

// Register регистрирует все маршруты согласно api.yaml.
// authMW создаётся вызывающей стороной через middleware.Auth(secret, store).
func (h *Handlers) Register(e *echo.Echo, authMW echo.MiddlewareFunc) {
	e.POST("/auth/login", h.Login)
	e.POST("/auth/logout", h.Logout, authMW)

	plans := e.Group("/test-plans", authMW)
	plans.GET("", h.ListTestPlans)
	plans.POST("", h.CreateTestPlan)
	plans.GET("/:id", h.GetTestPlan)
	plans.PUT("/:id", h.UpdateTestPlan)
	plans.DELETE("/:id", h.DeleteTestPlan)

	cases := e.Group("/test-cases", authMW)
	cases.GET("", h.ListTestCases)
	cases.POST("", h.CreateTestCase)
	cases.GET("/:id", h.GetTestCase)
	cases.PUT("/:id", h.UpdateTestCase)
	cases.DELETE("/:id", h.DeleteTestCase)
	cases.GET("/:id/steps", h.ListSteps)
	cases.POST("/:id/steps", h.AddStep)
	cases.PUT("/:id/steps/:stepId", h.UpdateStep)
	cases.DELETE("/:id/steps/:stepId", h.DeleteStep)

	r := e.Group("/runs", authMW)
	r.GET("", h.ListRuns)
	r.POST("", h.CreateRun)
	r.GET("/:id", h.GetRun)
	r.DELETE("/:id", h.DeleteRun)
	r.PATCH("/:id/status", h.UpdateRunStatus)
	r.GET("/:id/items", h.ListRunItems)
	r.POST("/:id/items", h.AddRunItem)
	r.DELETE("/:id/items/:itemId", h.DeleteRunItem)

	res := e.Group("/results", authMW)
	res.GET("", h.ListResults)
	res.POST("", h.CreateResult)
	res.GET("/:id", h.GetResult)
	res.PUT("/:id", h.UpdateResult)
	res.DELETE("/:id", h.DeleteResult)
	res.GET("/:id/artifacts", h.ListArtifacts)
	res.POST("/:id/artifacts", h.AddArtifact)

	at := e.Group("/autotests", authMW)
	at.GET("", h.ListAutotests)
	at.POST("", h.CreateAutotest)
	at.GET("/:id", h.GetAutotest)
	at.PUT("/:id", h.UpdateAutotest)
	at.DELETE("/:id", h.DeleteAutotest)
	at.POST("/:id/versions", h.AddVersion)

	e.GET("/environments", h.ListEnvironments, authMW)
	e.GET("/reports", h.ListReports, authMW)

	st := e.Group("/stats", authMW)
	st.GET("/runs", h.RunSummaries)
	st.GET("/cases", h.CaseStats)

	adm := e.Group("/admin", authMW)
	adm.GET("/users", h.ListUsers)
	adm.POST("/users", h.CreateUser)
	adm.GET("/users/:id", h.GetUser)
	adm.PATCH("/users/:id/active", h.SetActive)
	adm.PATCH("/users/:id/lock", h.SetLock)
	adm.POST("/users/:id/roles", h.AssignRole)
	adm.DELETE("/users/:id/roles/:roleId", h.RevokeRole)
	adm.GET("/audit-log", h.AuditLog)
}

// ErrorHandler — кастомный обработчик ошибок echo.
// Устанавливается через e.HTTPErrorHandler = handler.ErrorHandler.
func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	var he *echo.HTTPError
	if errors.As(err, &he) {
		msg, _ := he.Message.(string)
		if msg == "" {
			msg = http.StatusText(he.Code)
		}
		_ = c.JSON(he.Code, map[string]string{"code": statusCode(he.Code), "message": msg})
	} else {
		_ = c.JSON(http.StatusInternalServerError, map[string]string{
			"code": "internal_error", "message": "internal server error",
		})
	}
}

func statusCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	default:
		return "internal_error"
	}
}

// ── Хелперы ────────────────────────────────────────────────────────────────────

func parsePaging(c echo.Context) domain.Paging {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return domain.Paging{Limit: limit, Offset: offset}
}

func pathInt(c echo.Context, name string) (int, error) {
	v, err := strconv.Atoi(c.Param(name))
	if err != nil || v <= 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+name)
	}
	return v, nil
}

func httpErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrConflict):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrBadRequest):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}

// appRepo создаёт AppRepository из соединения текущей сессии (AD-9, AD-13).
func appRepo(c echo.Context) *repository.AppRepository {
	conn := middleware.ConnFromCtx(c.Request().Context())
	return repository.NewAppRepository(conn)
}

type pageResp[T any] struct {
	Data []T             `json:"data"`
	Meta domain.PageMeta `json:"meta"`
}

func page[T any](data []T, p domain.Paging) pageResp[T] {
	if data == nil {
		data = []T{}
	}
	return pageResp[T]{
		Data: data,
		Meta: domain.PageMeta{Total: p.Offset + len(data), Limit: p.Limit, Offset: p.Offset},
	}
}
