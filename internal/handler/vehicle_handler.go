package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/domain"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type VehicleHandler struct {
	service services.VehicleServiceAPI
	logger  *zap.Logger
}

func NewVehicleHandler(s services.VehicleServiceAPI, l *zap.Logger) *VehicleHandler {
	return &VehicleHandler{service: s, logger: l}
}

func (h *VehicleHandler) IngestData(w http.ResponseWriter, r *http.Request) {
	var req domain.IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.IngestData(r.Context(), req); err != nil {
		h.logger.Error("Failed to ingest data", zap.Error(err))
		http.Error(w, "Failed to ingest data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *VehicleHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	vehicleIDStr := r.URL.Query().Get("vehicle_id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		http.Error(w, "Invalid vehicle_id", http.StatusBadRequest)
		return
	}

	status, err := h.service.GetVehicleStatus(r.Context(), vehicleID)
	if err != nil {
		h.logger.Error("Failed to get status", zap.Error(err))
		http.Error(w, "Failed to retrieve status", http.StatusInternalServerError)
		return
	}
	if status == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *VehicleHandler) GetTrips(w http.ResponseWriter, r *http.Request) {
	vehicleIDStr := r.URL.Query().Get("vehicle_id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		http.Error(w, "Invalid vehicle_id", http.StatusBadRequest)
		return
	}

	trips, err := h.service.GetVehicleTrips(r.Context(), vehicleID)
	if err != nil {
		h.logger.Error("Failed to get trips", zap.Error(err))
		http.Error(w, "Failed to retrieve trips", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trips)
}
