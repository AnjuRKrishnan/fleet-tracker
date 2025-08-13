package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/auth"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/config"
	handlers "github.com/AnjuRKrishnan/fleet-tracker/internal/handler"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/middleware"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/services"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/store/postgres"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/store/redis"
	"github.com/AnjuRKrishnan/fleet-tracker/pkg/logger"
	"github.com/AnjuRKrishnan/fleet-tracker/pkg/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	zapLogger := logger.NewZapLogger()
	defer zapLogger.Sync()

	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		zapLogger.Fatal("Could not load configuration", zap.Error(err))
	}

	// Setup Database & Cache
	db, err := postgres.NewPostgresDB(cfg.PostgresURL)
	if err != nil {
		zapLogger.Fatal("Could not connect to PostgreSQL", zap.Error(err))
	}
	defer db.Close()

	cache, err := redis.NewRedisCache(cfg.RedisURL)
	if err != nil {
		zapLogger.Fatal("Could not connect to Redis", zap.Error(err))
	}

	// Setup Repositories
	vehicleRepo := postgres.NewVehicleRepository(db)
	vehicleCache := redis.NewVehicleCache(cache)

	// Setup Services
	vehicleService := services.NewVehicleService(vehicleRepo, vehicleCache)

	// Setup JWT Auth
	jwtAuth := auth.NewJWTAuth(cfg.JWTSecret)

	// Setup Handlers
	vehicleHandler := handlers.NewVehicleHandler(vehicleService, zapLogger)

	// Setup Router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestLogger(zapLogger))

	// Public routes (e.g., for generating a token if needed)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Private (authenticated) routes
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.JWTAuthenticator(jwtAuth))
		r.Post("/vehicle/ingest", vehicleHandler.IngestData)
		r.Get("/vehicle/status", vehicleHandler.GetStatus)
		r.Get("/vehicle/trips", vehicleHandler.GetTrips)
	})

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		zapLogger.Info("Starting server on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Could not start server", zap.Error(err))
		}
	}()

	// Start the simulated data stream and worker pool
	if cfg.SimulatorEnabled {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		dataChannel := services.StartDataSimulator(ctx)
		workerPool := services.NewWorkerPool(5, dataChannel, vehicleService, zapLogger)
		utils.SafeGo(workerPool.Run, "WorkerPool")
	}

	// Graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	zapLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server shutdown failed", zap.Error(err))
	}
	zapLogger.Info("Server stopped gracefully")

}
