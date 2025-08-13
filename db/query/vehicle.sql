-- name: UpsertVehicleStatus :exec
INSERT INTO vehicle (id, plate_number, last_status)
VALUES ($1, $2, $3::JSONB)
ON CONFLICT (id) DO UPDATE
SET plate_number = EXCLUDED.plate_number,
    last_status  = EXCLUDED.last_status;


-- name: GetVehicleStatus :one
SELECT last_status
FROM vehicle
WHERE id = $1;

-- name: CreateVehicle :one
INSERT INTO vehicle (id, plate_number, last_status)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
SET last_status = EXCLUDED.last_status
RETURNING *;

-- name: GetVehicleByPlate :one
SELECT *
FROM vehicle
WHERE plate_number = $1;

-- name: ListVehicles :many
SELECT *
FROM vehicle
ORDER BY ID DESC
LIMIT $1 OFFSET $2;

