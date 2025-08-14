package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/domain"
	handler "github.com/AnjuRKrishnan/fleet-tracker/internal/handler"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockVehicleService is a mock type for VehicleService
type MockVehicleService struct {
	mock.Mock
}

func (m *MockVehicleService) GetVehicleStatus(ctx context.Context, vehicleID uuid.UUID) (*domain.VehicleStatus, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VehicleStatus), args.Error(1)
}

func (m *MockVehicleService) IngestData(ctx context.Context, data domain.IngestRequest) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockVehicleService) GetVehicleTrips(ctx context.Context, vehicleID uuid.UUID) ([]domain.Trip, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Trip), args.Error(1)
}

// helper to get pointer to time
func ptrTime(t time.Time) *time.Time {
	return &t
}

func TestVehicleHandler_GetTrips(t *testing.T) {
	testVehicleUUID := uuid.New()
	pgVehicleUUID := pgtype.UUID{Bytes: testVehicleUUID, Valid: true}

	now := time.Now()
	testTrips := []domain.Trip{
		{
			ID:        pgtype.UUID{Bytes: testVehicleUUID, Valid: true},
			VehicleID: pgVehicleUUID,
			StartTime: pgtype.Timestamptz{Time: now, Valid: true},
			EndTime:   ptrTime(now.Add(1 * time.Hour)),
			Mileage:   120.5,
			AvgSpeed:  60,
		},
	}

	tests := []struct {
		name               string
		vehicleID          string
		setupMock          func(m *MockVehicleService)
		expectedStatusCode int
		validateBody       func(body []byte)
	}{
		{
			name:      "Success",
			vehicleID: testVehicleUUID.String(),
			setupMock: func(m *MockVehicleService) {
				m.On("GetVehicleTrips", mock.Anything, testVehicleUUID).Return(testTrips, nil)
			},
			expectedStatusCode: http.StatusOK,
			validateBody: func(body []byte) {
				var got []domain.Trip
				err := json.Unmarshal(body, &got)
				assert.NoError(t, err)
				assert.Len(t, got, 1)
				assert.Equal(t, testTrips[0].Mileage, got[0].Mileage)
				assert.Equal(t, testTrips[0].AvgSpeed, got[0].AvgSpeed)
			},
		},
		{
			name:               "Invalid UUID",
			vehicleID:          "invalid-uuid",
			setupMock:          func(m *MockVehicleService) {},
			expectedStatusCode: http.StatusBadRequest,
			validateBody: func(body []byte) {
				assert.Equal(t, "Invalid vehicle_id\n", string(body))
			},
		},
		{
			name:      "Internal Server Error",
			vehicleID: testVehicleUUID.String(),
			setupMock: func(m *MockVehicleService) {
				m.On("GetVehicleTrips", mock.Anything, testVehicleUUID).Return(nil, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			validateBody: func(body []byte) {
				assert.Equal(t, "Failed to retrieve trips\n", string(body))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockVehicleService)
			tc.setupMock(mockService)

			h := handler.NewVehicleHandler(mockService, zap.NewNop())
			req := httptest.NewRequest("GET", "/trips?vehicle_id="+tc.vehicleID, nil)
			rr := httptest.NewRecorder()
			h.GetTrips(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			tc.validateBody(rr.Body.Bytes())
			mockService.AssertExpectations(t)
		})
	}
}

func TestVehicleHandler_IngestData(t *testing.T) {
	testVehicleUUID := uuid.New()
	pgVehicleUUID := pgtype.UUID{Bytes: testVehicleUUID, Valid: true}

	tests := []struct {
		name               string
		body               []byte
		setupMock          func(m *MockVehicleService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Success",
			body: func() []byte {
				req := domain.IngestRequest{
					VehicleID: pgVehicleUUID,
					Status: domain.VehicleStatus{
						Speed: 50,
					},
				}
				b, _ := json.Marshal(req)
				return b
			}(),
			setupMock: func(m *MockVehicleService) {
				m.On("IngestData", mock.Anything, mock.AnythingOfType("domain.IngestRequest")).Return(nil)
			},
			expectedStatusCode: http.StatusAccepted,
			expectedBody:       "",
		},
		{
			name:               "Invalid JSON",
			body:               []byte(`{"vehicle_id":`), // malformed
			setupMock:          func(m *MockVehicleService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid request body\n",
		},
		{
			name: "Service Error",
			body: func() []byte {
				req := domain.IngestRequest{
					VehicleID: pgVehicleUUID,
					Status: domain.VehicleStatus{
						Speed: 80,
					},
				}
				b, _ := json.Marshal(req)
				return b
			}(),
			setupMock: func(m *MockVehicleService) {
				m.On("IngestData", mock.Anything, mock.AnythingOfType("domain.IngestRequest")).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Failed to ingest data\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockVehicleService)
			tc.setupMock(mockService)

			h := handler.NewVehicleHandler(mockService, zap.NewNop())
			req := httptest.NewRequest("POST", "/ingest", bytes.NewReader(tc.body))
			rr := httptest.NewRecorder()
			h.IngestData(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			assert.Equal(t, tc.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
