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

// VehicleServiceAPI defines the interface for vehicle service operations.
type VehicleServiceAPI interface {
	GetVehicleStatus(ctx context.Context, vehicleID uuid.UUID) (*domain.VehicleStatus, error)
	IngestData(ctx context.Context, data domain.IngestRequest) error
	GetVehicleTrips(ctx context.Context, vehicleID uuid.UUID) ([]domain.Trip, error)
}

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
	vehicleUUID := uuid.UUID(data.VehicleID.Bytes)
	if err := s.repo.UpdateVehicleStatus(ctx, vehicleUUID, data.PlateNumber, data.Status); err != nil {
		return err
	}

	// 2. Update the cache
	return s.cache.SetStatus(ctx, data.VehicleID.Bytes, &data.Status, CacheDuration)
}

// GetVehicleStatus retrieves the current status of a vehicle, trying the cache first.
func (s *VehicleService) GetVehicleStatus(ctx context.Context, vehicleID uuid.UUID) (*domain.VehicleStatus, error) {
	status, err := s.cache.GetStatus(ctx, vehicleID)
	if status != nil || err != nil {
		return status, err
	}
	// fallback to DB
	status, err = s.repo.GetVehicleStatus(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	// optionally write back to cache
	s.cache.SetStatus(ctx, vehicleID, status, time.Hour)
	return status, nil
}

// GetVehicleTrips retrieves the trip history for a vehicle in the last 24 hours.
func (s *VehicleService) GetVehicleTrips(ctx context.Context, vehicleID uuid.UUID) ([]domain.Trip, error) {
	since := time.Now().Add(-24 * time.Hour)
	return s.repo.FindTripsByVehicleID(ctx, vehicleID, since)
}
