package test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/domain"
	handler "github.com/AnjuRKrishnan/fleet-tracker/internal/handler"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockVehicleService is a mock type for the VehicleService
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

// Unused mocked methods
func (m *MockVehicleService) IngestData(ctx context.Context, data domain.IngestRequest) error {
	return nil
}
func (m *MockVehicleService) GetVehicleTrips(ctx context.Context, vehicleID uuid.UUID) ([]domain.Trip, error) {
	return nil, nil
}

func TestVehicleHandler_GetStatus(t *testing.T) {
	testVehicleID := uuid.New()
	testStatus := &domain.VehicleStatus{
		Location:  []float64{10, 20},
		Speed:     100,
		Timestamp: time.Now(),
	}

	testCases := []struct {
		name               string
		vehicleID          string
		setupMock          func(mockService *MockVehicleService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:      "Success",
			vehicleID: testVehicleID.String(),
			setupMock: func(mockService *MockVehicleService) {
				mockService.On("GetVehicleStatus", mock.Anything, testVehicleID).Return(testStatus, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"location":[10,20],"speed":100,`, // Partial check
		},
		{
			name:               "Invalid UUID",
			vehicleID:          "not-a-uuid",
			setupMock:          func(mockService *MockVehicleService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid vehicle_id\n",
		},
		{
			name:      "Not Found",
			vehicleID: testVehicleID.String(),
			setupMock: func(mockService *MockVehicleService) {
				mockService.On("GetVehicleStatus", mock.Anything, testVehicleID).Return(nil, nil)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "Not Found\n",
		},
		{
			name:      "Internal Server Error",
			vehicleID: testVehicleID.String(),
			setupMock: func(mockService *MockVehicleService) {
				mockService.On("GetVehicleStatus", mock.Anything, testVehicleID).Return(nil, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Failed to retrieve status\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockVehicleService)
			tc.setupMock(mockService)

			// Using a nop logger for tests
			handler := handler.NewVehicleHandler(mockService, zap.NewNop())

			req := httptest.NewRequest("GET", "/api/vehicle/status?vehicle_id="+tc.vehicleID, nil)
			rr := httptest.NewRecorder()

			handler.GetStatus(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}
