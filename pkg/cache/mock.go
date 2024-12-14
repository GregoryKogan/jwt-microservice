package cache

import (
	"context"
	"fmt"
	"os"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/redis/go-redis/v9"
)

type MockCache interface {
	Cache() *redis.Client
	Flush() error
	Cleanup() error
}

type MockCacheImpl struct {
	cache    *redis.Client
	resource *dockertest.Resource
}

func NewMockCache() MockCache {
	pool, err := dockertest.NewPool("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to Docker: %s", err)
		panic(err)
	}

	cache, resource := setupCache(pool)
	return &MockCacheImpl{cache, resource}
}

func (m *MockCacheImpl) Cache() *redis.Client {
	return m.cache
}

func (m *MockCacheImpl) Flush() error {
	return m.cache.FlushAll(context.Background()).Err()
}

func (m *MockCacheImpl) Cleanup() error {
	if err := m.resource.Close(); err != nil {
		panic(err)
	}

	return nil
}

func setupCache(pool *dockertest.Pool) (*redis.Client, *dockertest.Resource) {
	// Start Redis container
	redisResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "latest",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not start Redis container: %s", err)
		os.Exit(1)
	}

	// Exponential backoff-retry to connect to Redis
	var cache *redis.Client
	err = pool.Retry(func() error {
		cache = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", redisResource.GetPort("6379/tcp")),
		})
		_, err := cache.Ping(context.Background()).Result()
		return err
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to Redis: %s", err)
		os.Exit(1)
	}

	return cache, redisResource
}
