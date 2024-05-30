package config

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"

	"git.garena.com/shopee/common/cache"
)

type RedisSession cache.Cache

var (
	redisClient cache.Cache
)

type CacheConfig struct {
	Address          string `json:"address"`
	PoolSize         int    `json:"pool_size"`
	MaxIdle          int    `json:"max_idle"`
	ExpirationSecond int    `json:"expiration_second"`
	TimeoutMs        int    `json:"timeout_ms"`
	Password         string `json:"password"`
	CacheKeyPrefix   string `json:"cache_key_prefix"`
}

func applyRedisCacheConfig() {
	if confVal == nil || confVal.RedisCacheConfig == nil {
		panic("failed to init redis client, since RedisCacheConfig is nil")
	}

	client, err := cache.NewRedisCache("price_sync_redis_cache", cache.RedisConfig{
		Host:                  confVal.RedisCacheConfig.Address,
		PoolSize:              confVal.RedisCacheConfig.PoolSize,
		DefaultExpirationSecs: confVal.RedisCacheConfig.ExpirationSecond,
		ConnectTimeoutMillis:  confVal.RedisCacheConfig.TimeoutMs,
		ReadTimeoutMillis:     confVal.RedisCacheConfig.TimeoutMs,
		WriteTimeoutMillis:    confVal.RedisCacheConfig.TimeoutMs,
		Password:              confVal.RedisCacheConfig.Password,
		EncodingConfig: cache.EncodingConfig{
			DisableEncoding: true, // disable compression and bytes protocol, and sent raw bytes into cache directly
		},
	})
	if err != nil {
		panic(fmt.Sprintf("failed to init redis client, err=%s", err.Error()))
	}

	redisClient = client
	logging.GetLogger(context.Background()).Info("success to init redis client")
}

func GetRedisClient() RedisSession {
	return redisClient
}

func GetRedisCacheKeyPrefix() string {
	if confVal == nil || confVal.RedisCacheConfig == nil {
		return ""
	}
	return confVal.RedisCacheConfig.CacheKeyPrefix
}

func GetRedisConfig() *CacheConfig {
	if confVal == nil {
		return nil
	}

	return confVal.RedisCacheConfig
}
