package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/domain"
	"github.com/google/uuid"
)

type VehicleRepository struct {
	db *sql.DB
}

func NewPostgresDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func NewVehicleRepository(db *sql.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

func (r *VehicleRepository) UpdateVehicleStatus(ctx context.Context, vehicleID uuid.UUID, status domain.VehicleStatus) error {
	statusJSON, err := json.Marshal(status)
	if err != nil {
		return err
	}

	query := `UPDATE vehicle SET last_status = $1 WHERE id = $2`
	_, err = r.db.ExecContext(ctx, query, statusJSON, vehicleID)
	// In a real app, check if the vehicle exists and handle accordingly, e.g., by inserting it.
	return err
}

func (r *VehicleRepository) FindTripsByVehicleID(ctx context.Context, vehicleID uuid.UUID, since time.Time) ([]domain.Trip, error) {
	query := `
		SELECT id, vehicle_id, start_time, end_time, mileage, avg_speed
		FROM trips
		WHERE vehicle_id = $1 AND start_time >= $2
		ORDER BY start_time DESC
	`
	rows, err := r.db.QueryContext(ctx, query, vehicleID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []domain.Trip
	for rows.Next() {
		var t domain.Trip
		if err := rows.Scan(&t.ID, &t.VehicleID, &t.StartTime, &t.EndTime, &t.Mileage, &t.AvgSpeed); err != nil {
			return nil, err
		}
		trips = append(trips, t)
	}

	return trips, nil
}
