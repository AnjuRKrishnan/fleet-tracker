# fleet-tracker
Fleet Tracking Backend Service A lightweight Go backend for real-time vehicle tracking and trip history with simulated sensor data ingestion, PostgreSQL persistence, Redis caching, JWT-secured REST APIs, and clean architecture. Dockerized for easy setup and extensibility.

## Design Decisions

### Concurrency Model

A worker pool pattern is used to process incoming data from the simulated sensor stream.

- **`StartDataSimulator`**: A single goroutine that acts as a data producer, sending data to a channel every 2 seconds.

- **`WorkerPool`**: Manages a fixed number of worker goroutines. Each worker listens on the shared channel. This prevents overwhelming the service with a burst of data and controls the concurrency level.

- **`select { for {}}`**: This pattern is used within the simulator goroutine to allow for a graceful shutdown via a context.

- **`SafeGo`**: A utility function wraps each critical goroutine (`WorkerPool`, `DataSimulator`, individual workers). It uses `recover()` to catch any panics, log them, and prevent the entire application from crashing.

### Caching Strategy

A **Write-Through** caching strategy was chosen for the `GET /vehicle/status` endpoint.

- **Decision**: When new vehicle data is ingested via `POST /api/vehicle/ingest`, the service writes the data to PostgreSQL *and then* immediately writes the same data to the Redis cache.

- **Justification**: This approach ensures that the cache is always consistent with the primary database. It simplifies the read logic, as `GET /vehicle/status` can read directly from the cache without needing to check the database first. The tradeoff is slightly higher latency on writes, which is acceptable for this use case.

- **Invalidation**: The cache is invalidated in two ways:
  1.  **TTL (Time-To-Live)**: Each cache entry is set with a 5-minute expiration.
  2.  **On Write**: Every successful `ingest` operation overwrites the existing cache entry for that vehicle, ensuring the data is always fresh.

### PostgreSQL Indexing

Indexes have been created on the `trips` table to optimize query performance.

- `idx_trips_vehicle_id`: On `trips(vehicle_id)` because trip history is always fetched for a specific vehicle.
- `idx_trips_start_time`: On `trips(start_time)` because trips are filtered by a 24-hour time window.

#### `EXPLAIN ANALYZE` Output

EXPLAIN ANALYZE SELECT id, vehicle_id, start_time, end_time, mileage, avg_speed FROM trips WHERE vehicle_id = '90f8aed2-06bd-4abd-bce5-691c92b0cca7' AND start_time >= NOW() - INTERVAL '24 hours';

**Output:**

  QUERY PLAN
----------------------------------------------------------------------
<img width="980" height="442" alt="image" src="https://github.com/user-attachments/assets/8edc17a9-925f-4bc9-a672-eed77c8a9bbf" />

----------------------------------------------------------------------

The output shows that the query planner correctly chose an **Index Scan** on `idx_trips_vehicle_id`, resulting in very fast execution.


POSTMAN outputs:

<img width="2582" height="984" alt="image" src="https://github.com/user-attachments/assets/b04b6deb-8763-41f9-8dae-c653421120ec" />

<img width="1654" height="1328" alt="image" src="https://github.com/user-attachments/assets/31e0ea10-77de-49df-acf6-a2e38743948c" />

<img width="2626" height="1280" alt="image" src="https://github.com/user-attachments/assets/57df6b45-1430-48d7-844f-14a080788dde" />




