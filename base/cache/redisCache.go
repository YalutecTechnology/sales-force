package cache

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
	"github.com/labstack/gommon/log"
)

var (
	// ErrGettingRedisClient this error is thrown when there is an error creating the REDIS client
	ErrGettingRedisClient = errors.New("error trying to get client for REDIS")
	// ErrGettingRedis this error happends when expected PONG response from Redis was not received
	ErrGettingRedis = errors.New("error trying to get redis PONG")
)

// CommonRedisCache interface that holds method to retrieve cached sessions
type CommonRedisCache interface {
	StoreData(string, []byte, time.Duration) error
	RetrieveData(string) (string, error)
}

// RedisCache implements a session cache with Redis
type RedisCache struct {
	client     *redis.Client
	sessionTTL time.Duration
}

// RedisOptions holds the required configurations for a Redis Cache
type RedisOptions struct {
	FailOverOptions *redis.FailoverOptions
	Options         *redis.Options
	SessionsTTL     time.Duration
}

// NewRedisCache creates a RedisCachedSessions verifying that the redis client is connected
func NewRedisCache(config *RedisOptions) (*RedisCache, error) {
	var client *redis.Client
	if config.FailOverOptions != nil && len(config.FailOverOptions.MasterName) > 0 {
		client = redis.NewFailoverClient(config.FailOverOptions)
	}
	if config.Options != nil && len(config.Options.Addr) > 0 {
		client = redis.NewClient(config.Options)
	}
	if client == nil {
		return nil, ErrGettingRedisClient
	}
	rc := &RedisCache{
		client:     client,
		sessionTTL: config.SessionsTTL,
	}
	err := rc.ping()
	return rc, err
}

// ping tests connectivity for redis (PONG should be returned)
func (rc *RedisCache) ping() error {
	pong, err := rc.client.Ping().Result()
	if err != nil {
		return err
	}
	if pong != "PONG" {
		return ErrGettingRedis
	}
	log.Info("Connected to redis")
	return nil
}

// StoreData saves a user session on the Session Cache
func (rc *RedisCache) StoreData(key string, data []byte, ttl time.Duration) error {
	_, err := rc.client.Set(key, data, ttl).Result()
	if err != nil {
		return err
	}
	return nil
}

// RetrieveData returns a user session from the Session Cache
func (rc *RedisCache) RetrieveData(key string) (string, error) {
	data, err := rc.client.Get(key).Result()
	if err != nil {
		return "", err
	}
	return data, nil
}

// DeleteAll delete all data from Redis Cache
func (rc *RedisCache) DeleteAll() error {
	_, err := rc.client.FlushDb().Result()
	if err != nil {
		return err
	}
	return nil
}
