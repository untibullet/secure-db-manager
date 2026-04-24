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

	// Сервисы.
	authSvc := service.NewAuthService(store, cfg)
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
	)

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = handler.ErrorHandler

	e.Use(appmw.RequestLogger(log))
	e.Use(echomw.Recover())

	authMW := appmw.Auth(cfg.JWTSecret, store)
	h.Register(e, authMW)

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
