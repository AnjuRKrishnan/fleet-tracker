package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/domain"
	"github.com/go-redis/redis/v5"

	"github.com/google/uuid"
)

type VehicleCache struct {
	client *redis.Client
}

func NewRedisCache(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

func NewVehicleCache(client *redis.Client) *VehicleCache {
	return &VehicleCache{client: client}
}

func (c *VehicleCache) key(vehicleID uuid.UUID) string {
	return fmt.Sprintf("vehicle:%s:status", vehicleID.String())
}

func (c *VehicleCache) SetStatus(ctx context.Context, vehicleID uuid.UUID, status *domain.VehicleStatus, expiration time.Duration) error {
	statusJSON, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(vehicleID), statusJSON, expiration).Err()
}

func (c *VehicleCache) GetStatus(ctx context.Context, vehicleID uuid.UUID) (*domain.VehicleStatus, error) {
	val, err := c.client.Get(ctx, c.key(vehicleID)).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	} else if err != nil {
		return nil, err
	}

	var status domain.VehicleStatus
	if err := json.Unmarshal([]byte(val), &status); err != nil {
		return nil, err
	}
	return &status, nil
}
