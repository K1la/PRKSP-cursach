package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/username/parking-service/internal/config"
	"github.com/username/parking-service/internal/handler"
	"github.com/username/parking-service/internal/migration"
	"github.com/username/parking-service/internal/repository"
	"github.com/username/parking-service/internal/service"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel(),
	}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		logger.Warn("database is not reachable at startup", "error", err)
	} else if cfg.AutoMigrate {
		if err := migration.NewMigrator(db, "migrations").Up(ctx); err != nil {
			logger.Error("database migration failed", "error", err)
			os.Exit(1)
		}
		logger.Info("database migrations applied")
	}

	router := buildRouter(cfg, db, logger)
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server started", "addr", server.Addr, "env", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}

func buildRouter(cfg config.Config, db *pgxpool.Pool, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(handler.Recoverer(logger))
	r.Use(handler.RequestLogger(logger))
	r.Use(handler.CORS(cfg.CORSOrigins))

	healthHandler := handler.NewHealthHandler(db)
	r.Get("/health", healthHandler.Check)

	userRepository := repository.NewUserRepository(db)
	parkingRepository := repository.NewParkingRepository(db)
	bookingRepository := repository.NewBookingRepository(db)
	reviewRepository := repository.NewReviewRepository(db)
	adminRepository := repository.NewAdminRepository(db)

	authService := service.NewAuthService(userRepository, cfg)
	userService := service.NewUserService(userRepository)
	parkingService := service.NewParkingService(parkingRepository)
	bookingService := service.NewBookingService(bookingRepository, parkingRepository)
	reviewService := service.NewReviewService(reviewRepository, parkingRepository)
	adminService := service.NewAdminService(adminRepository)

	api := handler.NewAPI(authService, userService, parkingService, bookingService, reviewService, adminService)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler.Check)
		api.Routes(r)
	})

	return r
}
