package services

import (
	"context"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/domain"
	"github.com/google/uuid"
)

const (
	// CacheDuration defines how long vehicle status is cached.
	CacheDuration = 5 * time.Minute
)

// VehicleService encapsulates the business logic for vehicle operations.
type VehicleService struct {
	repo  domain.VehicleRepository
	cache domain.VehicleCache
}

// NewVehicleService creates a new VehicleService.
func NewVehicleService(repo domain.VehicleRepository, cache domain.VehicleCache) *VehicleService {
	return &VehicleService{
		repo:  repo,
		cache: cache,
	}
}

// IngestData processes new vehicle data, updating the database and cache.
func (s *VehicleService) IngestData(ctx context.Context, data domain.IngestRequest) error {
	// 1. Update the database (write-through)
	if err := s.repo.UpdateVehicleStatus(ctx, data.VehicleID, data.Status); err != nil {
		return err
	}

	// 2. Update the cache
	return s.cache.SetStatus(ctx, data.VehicleID, &data.Status, CacheDuration)
}

// GetVehicleStatus retrieves the current status of a vehicle, trying the cache first.
func (s *VehicleService) GetVehicleStatus(ctx context.Context, vehicleID uuid.UUID) (*domain.VehicleStatus, error) {
	// Note: With a write-through cache, we can just read from the cache.
	// If it were cache-aside, we'd have logic here to check the DB on a cache miss.
	return s.cache.GetStatus(ctx, vehicleID)
}

// GetVehicleTrips retrieves the trip history for a vehicle in the last 24 hours.
func (s *VehicleService) GetVehicleTrips(ctx context.Context, vehicleID uuid.UUID) ([]domain.Trip, error) {
	since := time.Now().Add(-24 * time.Hour)
	return s.repo.FindTripsByVehicleID(ctx, vehicleID, since)
}
