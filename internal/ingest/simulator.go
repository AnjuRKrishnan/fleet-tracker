package ingest

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/model"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/services"
)

func StartSimulator(svc *services.VehicleService, vehicleID int64) {
	if vehicleID == 0 {
		log.Println("simulator: no vehicle id configured")
		return
	}

	ch := make(chan model.VehicleStatus)
	go producer(ch, vehicleID)
	go consumer(svc, ch)
}

func producer(ch chan<- model.VehicleStatus, vehicleID int64) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for t := range ticker.C {
		lat := 25.276987 + (rand.Float64()-0.5)/100.0
		lon := 55.296249 + (rand.Float64()-0.5)/100.0
		speed := rand.Float64()*30 + 20 // 20-50
		ch <- model.VehicleStatus{
			VehicleID: vehicleID,
			Location:  [2]float64{lat, lon},
			Speed:     speed,
			Timestamp: t.UTC(),
		}
	}
}

func consumer(svc *services.VehicleService, ch <-chan model.VehicleStatus) {
	for s := range ch {
		ctx := context.Background()
		err := svc.IngestStatus(ctx, s)
		if err != nil {
			log.Printf("sim ingest failed: %v", err)
		} else {
			log.Printf("sim ingest ok: vehicle=%d time=%s", s.VehicleID, s.Timestamp.String())
		}
	}
}
