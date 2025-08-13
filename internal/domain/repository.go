package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// VehicleRepository defines the interface for database operations related to vehicles and trips.
type VehicleRepository interface {
	UpdateVehicleStatus(ctx context.Context, vehicleID uuid.UUID, status VehicleStatus) error
	FindTripsByVehicleID(ctx context.Context, vehicleID uuid.UUID, since time.Time) ([]Trip, error)
	// In a real application, you'd also have methods like:
	// CreateVehicle(ctx context.Context, vehicle Vehicle) error
	// FindVehicleByID(ctx context.Context, vehicleID uuid.UUID) (*Vehicle, error)
}

// VehicleCache defines the interface for caching vehicle status.
type VehicleCache interface {
	SetStatus(ctx context.Context, vehicleID uuid.UUID, status *VehicleStatus, expiration time.Duration) error
	GetStatus(ctx context.Context, vehicleID uuid.UUID) (*VehicleStatus, error)
}
