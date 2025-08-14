package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// VehicleRepository defines the interface for database operations related to vehicles and trips.
type VehicleRepository interface {
	UpdateVehicleStatus(ctx context.Context, vehicleID uuid.UUID, plateNumber string, status VehicleStatus) error
	FindTripsByVehicleID(ctx context.Context, vehicleID uuid.UUID, since time.Time) ([]Trip, error)
	GetVehicleStatus(ctx context.Context, vehicleID uuid.UUID) (*VehicleStatus, error) // <-- new
}

// VehicleCache defines the interface for caching vehicle status.
type VehicleCache interface {
	SetStatus(ctx context.Context, vehicleID uuid.UUID, status *VehicleStatus, expiration time.Duration) error
	GetStatus(ctx context.Context, vehicleID uuid.UUID) (*VehicleStatus, error)
}
