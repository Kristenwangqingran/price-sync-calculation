package config

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"

	"git.garena.com/shopee/common/cache"
)

type LocalCacheSession cache.Cache

const localCacheNameForMerchantConfig = "merchant_config_local_cache"
const localCacheNameForShop = "shop_local_cache"

var (
	localCacheForMerchantConfigClient cache.Cache
	localCacheForShopClient           cache.Cache
)

func applyLocalCacheCacheConfig() {
	if confVal == nil || confVal.LocalCacheForMerchantConfig == nil {
		panic("failed to init LocalCacheForMerchant client, since LocalCacheForMerchantConfig is nil")
	}

	client, err := InitLocalCache(localCacheNameForMerchantConfig, confVal.LocalCacheForMerchantConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to init LocalCacheForMerchant client, err=%s", err.Error()))
	}

	localCacheForMerchantConfigClient = client
	logging.GetLogger(context.Background()).Info("success to init LocalCacheForMerchant client")

	if confVal == nil || confVal.LocalCacheForShop == nil {
		panic("failed to init LocalCacheForMerchant client, since LocalCacheForShop is nil")
	}

	shopClient, err := InitLocalCache(localCacheNameForShop, confVal.LocalCacheForShop)
	if err != nil {
		panic(fmt.Sprintf("failed to init LocalCacheForShop client, err=%s", err.Error()))
	}

	localCacheForShopClient = shopClient
	logging.GetLogger(context.Background()).Info("success to init LocalCacheForShop client")

}

func GetLocalCacheForMerchantConfigClient() LocalCacheSession {
	return localCacheForMerchantConfigClient
}

func GetLocalCacheForShopClient() LocalCacheSession {
	return localCacheForShopClient
}

func InitLocalCache(name string, config *cache.InMemoryCacheConfig) (*cache.InMemoryCache, error) {
	store, err := cache.NewInMemoryCache(name, *config)
	if err != nil {
		return nil, err
	}
	return store, nil
}
