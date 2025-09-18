package infrastructure

import (
	"fmt"
	"log"
    "context"
    "blog-api/internal/config"

	"github.com/go-redis/redis/v8"
)

// CacheClient wraps Redis operations
type CacheClient struct {
	client *redis.Client;
    ctx    context.Context
}

// NewCacheClient creates a new Redis cache client
func NewCacheClient(cfg *config.Config) (*CacheClient, error) {

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     "", // no password for now
		DB:           0,  // default DB
		PoolSize:     10, // connection pool size
		MinIdleConns: 2,  // minimum idle connections
	})

	// Test connection
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Printf("✅ Redis connected successfully")

	return &CacheClient{
		client: rdb,
        ctx:    context.Background(),
	}, nil
}

// GetClient returns the Redis client
func (cc *CacheClient) GetClient() *redis.Client {
	return cc.client
}

// Close closes the Redis connection
func (cc *CacheClient) Close() error {
	if cc.client != nil {
		err := cc.client.Close()
		if err != nil {
			return fmt.Errorf("failed to close Redis: %v", err)
		}
		log.Printf("✅ Redis connection closed")
	}
	return nil
}

// Health checks Redis health
func (cc *CacheClient) Health() bool {
	if cc.client == nil {
		return false
	}
	_, err := cc.client.Ping(cc.ctx).Result()
	return err == nil
}

// Stats returns Redis connection statistics
func (cc *CacheClient) Stats() map[string]interface{} {
	if cc.client == nil {
		return map[string]interface{}{"status": "disconnected"}
	}

	poolStats := cc.client.PoolStats()
	return map[string]interface{}{
		"hits":        poolStats.Hits,
		"misses":      poolStats.Misses,
		"timeouts":    poolStats.Timeouts,
		"total_conns": poolStats.TotalConns,
		"idle_conns":  poolStats.IdleConns,
		"stale_conns": poolStats.StaleConns,
	}
}
