package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/controllers"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/ingest"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/middleware"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/repository"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/services"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbURL := os.Getenv("DATABASE_URL")
	redisAddr := os.Getenv("REDIS_ADDR")
	jwtSecret := os.Getenv("JWT_SECRET")

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}

	// ping with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dbConn.PingContext(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	queries := repository.New(dbConn)
	svc := services.NewVehicleService(queries, rdb)
	ctl := controllers.NewVehicleController(svc)

	r := mux.NewRouter()
	// logging middleware
	r.Use(controllers.LoggingMiddleware)
	// jwt middleware applied to required routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.NewJWTMiddleware(jwtSecret).Middleware)

	api.HandleFunc("/vehicle/status", ctl.GetStatusHandler).Methods("GET")
	api.HandleFunc("/vehicle/trips", ctl.GetTripsHandler).Methods("GET")
	api.HandleFunc("/vehicle/ingest", ctl.IngestHandler).Methods("POST")

	simVehicleIDStr := os.Getenv("SIMULATOR_VEHICLE_ID")
	var simVehicleID int64
	if simVehicleIDStr != "" {
		id, err := strconv.ParseInt(simVehicleIDStr, 10, 64)
		if err != nil {
			log.Fatalf("invalid SIMULATOR_VEHICLE_ID: %v", err)
		}
		simVehicleID = id
	}

	go ingest.StartSimulator(svc, simVehicleID)
	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("listening on %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
