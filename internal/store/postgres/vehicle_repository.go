package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/db"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type VehicleRepository struct {
	q *db.Queries
}

// NewVehicleRepository creates a new repository.
func NewVehicleRepository(dbtx db.DBTX) *VehicleRepository {
	return &VehicleRepository{
		q: db.New(dbtx),
	}
}

// UpdateVehicleStatus calls the generated method.
func (r *VehicleRepository) UpdateVehicleStatus(ctx context.Context, vehicleID uuid.UUID, plateNumber string, status domain.VehicleStatus) error {
	statusJSON, err := json.Marshal(status)
	if err != nil {
		return err
	}

	// 5. Use the generated parameter struct from the 'db' package
	params := db.UpsertVehicleStatusParams{
		Column3:     string(statusJSON),
		PlateNumber: plateNumber,
		ID:          pgtype.UUID{Bytes: vehicleID, Valid: true},
	}

	return r.q.UpsertVehicleStatus(ctx, params)
}

func (r *VehicleRepository) FindTripsByVehicleID(ctx context.Context, vehicleID uuid.UUID, since time.Time) ([]domain.Trip, error) {
	// Use the generated ListTripsByVehicleID method
	dbTrips, err := r.q.ListTripsByVehicle(ctx, db.ListTripsByVehicleParams{
		VehicleID: pgtype.UUID{Bytes: vehicleID, Valid: true},
		StartTime: pgtype.Timestamptz{Time: since, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Map the database models to our domain models
	var domainTrips []domain.Trip
	for _, dt := range dbTrips {
		var endTime *time.Time
		if dt.EndTime.Valid {
			endTime = &dt.EndTime.Time
		}
		domainTrips = append(domainTrips, domain.Trip{
			ID:        dt.ID,
			VehicleID: dt.VehicleID,
			StartTime: dt.StartTime,
			EndTime:   endTime,
			Mileage:   dt.Mileage.Float64,
			AvgSpeed:  dt.AvgSpeed.Float64,
		})
	}

	return domainTrips, nil
}

func (r *VehicleRepository) GetVehicleStatus(ctx context.Context, vehicleID uuid.UUID) (*domain.VehicleStatus, error) {
	row, err := r.q.GetVehicleStatus(ctx, pgtype.UUID{Bytes: vehicleID, Valid: true})
	if err != nil {
		return nil, err
	}
	var status domain.VehicleStatus
	if err := json.Unmarshal([]byte(row), &status); err != nil {
		return nil, err
	}

	return &status, nil
}
