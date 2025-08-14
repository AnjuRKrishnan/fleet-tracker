package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/domain"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/services"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Repository ---
type MockVehicleRepository struct {
	mock.Mock
}

func (m *MockVehicleRepository) UpdateVehicleStatus(ctx context.Context, vehicleID uuid.UUID, plate string, status domain.VehicleStatus) error {
	args := m.Called(ctx, vehicleID, plate, status)
	return args.Error(0)
}

func (m *MockVehicleRepository) GetVehicleStatus(ctx context.Context, vehicleID uuid.UUID) (*domain.VehicleStatus, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VehicleStatus), args.Error(1)
}

func (m *MockVehicleRepository) FindTripsByVehicleID(ctx context.Context, vehicleID uuid.UUID, since time.Time) ([]domain.Trip, error) {
	args := m.Called(ctx, vehicleID, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Trip), args.Error(1)
}

// --- Mock Cache ---
type MockVehicleCache struct {
	mock.Mock
}

func (m *MockVehicleCache) SetStatus(ctx context.Context, vehicleID uuid.UUID, status *domain.VehicleStatus, ttl time.Duration) error {
	args := m.Called(ctx, vehicleID, status, ttl)
	return args.Error(0)
}

func (m *MockVehicleCache) GetStatus(ctx context.Context, vehicleID uuid.UUID) (*domain.VehicleStatus, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VehicleStatus), args.Error(1)
}

// --- Tests for IngestData ---
func TestVehicleService_IngestData(t *testing.T) {
	vehicleID := uuid.New()
	status := domain.VehicleStatus{Speed: 60}

	tests := []struct {
		name       string
		setupMocks func(repo *MockVehicleRepository, cache *MockVehicleCache)
		wantErr    bool
	}{
		{
			name: "Success",
			setupMocks: func(repo *MockVehicleRepository, cache *MockVehicleCache) {
				repo.On("UpdateVehicleStatus", mock.Anything, vehicleID, "", status).Return(nil)
				cache.On("SetStatus", mock.Anything, vehicleID, &status, services.CacheDuration).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Repo Error",
			setupMocks: func(repo *MockVehicleRepository, cache *MockVehicleCache) {
				repo.On("UpdateVehicleStatus", mock.Anything, vehicleID, "", status).Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "Cache Error",
			setupMocks: func(repo *MockVehicleRepository, cache *MockVehicleCache) {
				repo.On("UpdateVehicleStatus", mock.Anything, vehicleID, "", status).Return(nil)
				cache.On("SetStatus", mock.Anything, vehicleID, &status, services.CacheDuration).Return(errors.New("cache error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockVehicleRepository)
			mockCache := new(MockVehicleCache)
			tt.setupMocks(mockRepo, mockCache)

			svc := services.NewVehicleService(mockRepo, mockCache)
			err := svc.IngestData(context.Background(), domain.IngestRequest{
				VehicleID: pgtype.UUID{Bytes: vehicleID, Valid: true},
				Status:    status,
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// --- Tests for GetVehicleTrips ---
func TestVehicleService_GetVehicleTrips(t *testing.T) {
	vehicleID := uuid.New()
	trips := []domain.Trip{
		{VehicleID: pgtype.UUID{Bytes: vehicleID, Valid: true}, Mileage: 100},
	}

	tests := []struct {
		name       string
		setupMocks func(repo *MockVehicleRepository)
		wantTrips  []domain.Trip
		wantErr    bool
	}{
		{
			name: "Success",
			setupMocks: func(repo *MockVehicleRepository) {
				repo.On("FindTripsByVehicleID", mock.Anything, vehicleID, mock.Anything).Return(trips, nil)
			},
			wantTrips: trips,
			wantErr:   false,
		},
		{
			name: "Repo Error",
			setupMocks: func(repo *MockVehicleRepository) {
				repo.On("FindTripsByVehicleID", mock.Anything, vehicleID, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantTrips: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockVehicleRepository)
			tt.setupMocks(mockRepo)

			svc := services.NewVehicleService(mockRepo, nil)
			got, err := svc.GetVehicleTrips(context.Background(), vehicleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTrips, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
