-- name: InsertTrip :exec
INSERT INTO trips (id, vehicle_id, start_time, end_time, mileage, avg_speed)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetTripsLast24Hours :many
SELECT id, vehicle_id, start_time, end_time, mileage, avg_speed
FROM trips
WHERE vehicle_id = $1
  AND start_time >= now() - interval '24 hours'
ORDER BY start_time DESC;

-- name: GetTripByID :one
SELECT *
FROM trips
WHERE id = $1;

-- name: ListTripsByVehicle :many
SELECT *
FROM trips
WHERE vehicle_id = $1
AND start_time >= $2 
ORDER BY start_time DESC;
