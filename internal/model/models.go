package model

import "time"

type VehicleStatus struct {
	VehicleID int64      `json:"vehicle_id"`
	Location  [2]float64 `json:"location"` // [lat, lng]
	Speed     float64    `json:"speed"`
	Timestamp time.Time  `json:"timestamp"`
}

type Vehicle struct {
	ID          string      `json:"id"`
	PlateNumber string      `json:"plate_number"`
	LastStatus  interface{} `json:"last_status"`
}

type Trip struct {
	ID        int64     `json:"id"`
	VehicleID int64     `json:"vehicle_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Mileage   float64   `json:"mileage"`
	AvgSpeed  float64   `json:"avg_speed"`
}
