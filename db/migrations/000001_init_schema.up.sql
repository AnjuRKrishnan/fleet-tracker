-- Enable uuid-ossp extension for uuid_generate_v4()
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE vehicle (
    id BIGSERIAL PRIMARY KEY,
    plate_number TEXT NOT NULL,
    last_status JSONB NOT NULL DEFAULT '{}'  -- default empty JSON object
);

CREATE TABLE trips (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL REFERENCES vehicle(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    mileage DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_speed DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_trips_vehicle_start_time
  ON trips (vehicle_id, start_time DESC);
