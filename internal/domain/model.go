package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Vehicle represents a vehicle in the system.
type Vehicle struct {
	ID          uuid.UUID      `json:"id"`
	PlateNumber string         `json:"plate_number"`
	LastStatus  *VehicleStatus `json:"last_status,omitempty"`
}

// VehicleStatus represents the real-time status of a vehicle.
type VehicleStatus struct {
	Location  []float64 `json:"location"` // [longitude, latitude]
	Speed     float64   `json:"speed"`
	Timestamp time.Time `json:"timestamp"`
}

// Trip represents a single journey made by a vehicle.
type Trip struct {
	ID        pgtype.UUID        `json:"id"`
	VehicleID pgtype.UUID        `json:"vehicle_id"`
	StartTime pgtype.Timestamptz `json:"start_time"`
	EndTime   *time.Time         `json:"end_time,omitempty"`
	Mileage   float64            `json:"mileage"`
	AvgSpeed  float64            `json:"avg_speed"`
}

// IngestRequest is the structure for incoming data from the /ingest endpoint.
type IngestRequest struct {
	VehicleID   pgtype.UUID   `json:"vehicle_id"`
	Status      VehicleStatus `json:"status"`
	PlateNumber string        `json:"plate_number"`
}
