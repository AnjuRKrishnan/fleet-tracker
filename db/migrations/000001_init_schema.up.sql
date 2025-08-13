-- Enable uuid-ossp extension for uuid_generate_v4()
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE vehicle (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plate_number TEXT NOT NULL,
    last_status JSONB NOT NULL DEFAULT '{}'  -- default empty JSON object
);

CREATE TABLE trips (
       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicle(id) ON DELETE CASCADE,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    mileage FLOAT,
    avg_speed FLOAT
);

--indexes

CREATE INDEX idx_trips_vehicle_id ON trips(vehicle_id);

CREATE INDEX idx_trips_start_time ON trips(start_time DESC);
