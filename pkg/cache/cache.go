package cache

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitCacheConnection() *redis.Client {
	addr := fmt.Sprintf("%s:%d", viper.GetString("cache.host"), viper.GetInt("cache.port"))
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}
