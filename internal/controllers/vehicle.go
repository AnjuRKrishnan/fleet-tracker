package controllers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/model"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/services"
)

type VehicleController struct {
	svc *services.VehicleService
}

func NewVehicleController(svc *services.VehicleService) *VehicleController {
	return &VehicleController{svc: svc}
}

// Logging middleware
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %v", r.Method, r.RequestURI, r.RemoteAddr, time.Since(start))
	})
}

func (vc *VehicleController) GetStatusHandler(w http.ResponseWriter, r *http.Request) {
	vehicleIDStr := r.URL.Query().Get("vehicle_id")
	if vehicleIDStr == "" {
		http.Error(w, "vehicle_id is required", http.StatusBadRequest)
		return
	}
	vehicleID, err := strconv.ParseInt(vehicleIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid vehicle_id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	vs, err := vc.svc.GetVehicleStatus(ctx, vehicleID)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if vs == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vs)
}

func (vc *VehicleController) GetTripsHandler(w http.ResponseWriter, r *http.Request) {
	vehicleIDStr := r.URL.Query().Get("vehicle_id")
	if vehicleIDStr == "" {
		http.Error(w, "vehicle_id is required", http.StatusBadRequest)
		return
	}
	vehicleID, err := strconv.ParseInt(vehicleIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid vehicle_id", http.StatusBadRequest)
		return
	}

	trips, err := vc.svc.GetTrips(r.Context(), vehicleID)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trips)
}

func (vc *VehicleController) IngestHandler(w http.ResponseWriter, r *http.Request) {
	var payload model.VehicleStatus
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if payload.VehicleID == 0 {
		http.Error(w, "vehicle_id required", http.StatusBadRequest)
		return
	}
	if err := vc.svc.IngestStatus(r.Context(), payload); err != nil {
		http.Error(w, "ingest error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"ingested"}`))
}
