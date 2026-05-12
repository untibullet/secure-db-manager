package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/untibullet/secure-db-manager/internal/config"
	"github.com/untibullet/secure-db-manager/internal/handler"
	appmw "github.com/untibullet/secure-db-manager/internal/middleware"
	"github.com/untibullet/secure-db-manager/internal/repository"
	"github.com/untibullet/secure-db-manager/internal/seclog"
	"github.com/untibullet/secure-db-manager/internal/service"
	"github.com/untibullet/secure-db-manager/internal/session"
)

func main() {
	log := slog.Default()

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Суперпользовательский пул для DDL и логина (AD-10).
	sysPool, err := pgxpool.New(ctx, cfg.SuperuserDSN())
	if err != nil {
		log.Error("sysPool connect", "err", err)
		os.Exit(1)
	}
	defer sysPool.Close()

	roleRepo := repository.NewRoleRepository(sysPool)

	store := session.NewStore()
	store.StartReaper(ctx, cfg.ReaperInterval)

	sl, err := seclog.New(cfg.SeclogPath)
	if err != nil {
		log.Error("seclog init", "err", err)
		os.Exit(1)
	}
	defer sl.Close()

	// Сервисы.
	authSvc := service.NewAuthService(store, cfg, sl)
	adminSvc := service.NewAdminService(roleRepo)

	h := handler.New(
		authSvc,
		&service.TestPlanService{},
		&service.TestCaseService{},
		&service.RunService{},
		&service.ResultService{},
		&service.AutotestService{},
		&service.ReferenceService{},
		&service.StatsService{},
		adminSvc,
		roleRepo,
		cfg.SessionTTL,
		sl,
	)

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = h.ErrorHandler

	e.Use(appmw.RequestLogger(log))
	e.Use(echomw.Recover())

	e.Static("/static", "web/static")
	e.GET("/", func(c echo.Context) error { return c.File("web/templates/index.html") })
	e.GET("/login", func(c echo.Context) error { return c.File("web/templates/login.html") })

	authMW := appmw.Auth(cfg.JWTSecret, store, sl)
	h.Register(e, authMW)

	// SPA fallback: любой GET без совпадения с API-маршрутом отдаёт index.html,
	// чтобы прямая навигация по URL работала в браузере.
	e.RouteNotFound("/*", func(c echo.Context) error {
		if c.Request().Method == http.MethodGet {
			return c.File("web/templates/index.html")
		}
		return echo.ErrNotFound
	})

	// Graceful shutdown.
	go func() {
		if err := e.Start(cfg.ServerAddr); err != nil && err != http.ErrServerClosed {
			log.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", "err", err)
	}
	log.Info("server stopped")
}
