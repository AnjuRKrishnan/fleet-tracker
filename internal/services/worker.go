package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/domain"
	"github.com/AnjuRKrishnan/fleet-tracker/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// The vehicle ID to simulate data for.
var simulatedVehicleID = uuid.MustParse("d9c1b442-fb2f-412a-9d2a-a3ab499cd91c")

// StartDataSimulator simulates incoming sensor data every 2 seconds.
func StartDataSimulator(ctx context.Context) <-chan domain.IngestRequest {
	dataChannel := make(chan domain.IngestRequest, 10) // Buffered channel

	utils.SafeGo(func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		defer close(dataChannel)

		for {
			select {
			case <-ticker.C:
				data := domain.IngestRequest{
					VehicleID: simulatedVehicleID,
					Status: domain.VehicleStatus{
						Location:  []float64{55.296249, 25.276987}, // Example location
						Speed:     60.5,
						Timestamp: time.Now().UTC(),
					},
				}
				dataChannel <- data
			case <-ctx.Done():
				return
			}
		}
	}, "DataSimulator")

	return dataChannel
}

// WorkerPool processes ingested data from a channel.
type WorkerPool struct {
	numWorkers int
	dataChan   <-chan domain.IngestRequest
	service    *VehicleService
	logger     *zap.Logger
}

func NewWorkerPool(numWorkers int, dataChan <-chan domain.IngestRequest, service *VehicleService, logger *zap.Logger) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		dataChan:   dataChan,
		service:    service,
		logger:     logger,
	}
}

// Run starts the workers.
func (wp *WorkerPool) Run() {
	for i := 0; i < wp.numWorkers; i++ {
		workerID := i + 1
		utils.SafeGo(func() {
			wp.worker(workerID)
		}, "Worker", workerID)
	}
}

func (wp *WorkerPool) worker(id int) {
	wp.logger.Info("Starting worker", "id", id)
	for data := range wp.dataChan {
		jsonData, _ := json.Marshal(data)
		wp.logger.Info("Worker processing data", "worker_id", id, "data", string(jsonData))

		// In a real scenario, you'd extract a user ID or other claims from a context.
		// For the simulator, we use a background context.
		err := wp.service.IngestData(context.Background(), data)
		if err != nil {
			wp.logger.Error("Worker failed to process data", "worker_id", id, "error", err)
		}
	}
	wp.logger.Info("Stopping worker", "id", id)
}
