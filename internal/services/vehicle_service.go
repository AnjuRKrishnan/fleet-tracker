package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/model"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/repository"
	"github.com/go-redis/redis/v8"
)

const statusCacheTTL = 5 * time.Minute

type VehicleService struct {
	q   *repository.Queries
	rdb *redis.Client
}

func NewVehicleService(q *repository.Queries, rdb *redis.Client) *VehicleService {
	return &VehicleService{q: q, rdb: rdb}
}

func (s *VehicleService) cacheKey(vehicleID string) string {
	return fmt.Sprintf("vehicle:status:%s", vehicleID)
}

func (s *VehicleService) GetVehicleStatus(ctx context.Context, vehicleID int64) (*model.VehicleStatus, error) {
	key := s.cacheKey(fmt.Sprint(vehicleID))

	// Try cache first
	if val, err := s.rdb.Get(ctx, key).Result(); err == nil {
		var vs model.VehicleStatus
		if err := json.Unmarshal([]byte(val), &vs); err == nil {
			return &vs, nil
		}
	}

	// Fetch from DB
	raw, err := s.q.GetVehicleStatus(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if raw == "" {
		return nil, nil
	}

	// Convert string to []byte
	var vs model.VehicleStatus
	if err := json.Unmarshal([]byte(raw), &vs); err != nil {
		return nil, err
	}

	// Save to cache
	if b, err := json.Marshal(vs); err == nil {
		_ = s.rdb.Set(ctx, key, b, statusCacheTTL).Err()
	}

	return &vs, nil
}

func (s *VehicleService) GetTrips(ctx context.Context, vehicleID int64) ([]model.Trip, error) {
	rows, err := s.q.GetTripsLast24Hours(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	trips := make([]model.Trip, len(rows))
	for i, r := range rows {
		trips[i] = model.Trip{
			ID:        r.ID,
			VehicleID: r.VehicleID,
			StartTime: r.StartTime,
			EndTime:   r.EndTime,
			Mileage:   r.Mileage,
			AvgSpeed:  r.AvgSpeed,
		}
	}
	return trips, nil
}

func (s *VehicleService) IngestStatus(ctx context.Context, status model.VehicleStatus) error {
	if status.VehicleID < 1 {
		return errors.New("vehicle_id required")
	}

	// Marshal status to JSON
	b, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	jsonStr := string(b)

	if err := s.q.UpsertVehicleStatus(ctx, repository.UpsertVehicleStatusParams{
		ID:      status.VehicleID,
		Column2: jsonStr, // <- pass string, not []byte
	}); err != nil {
		return err
	}

	// Cache in Redis
	if err := s.rdb.Set(ctx, s.cacheKey(fmt.Sprint(status.VehicleID)), b, statusCacheTTL).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}
