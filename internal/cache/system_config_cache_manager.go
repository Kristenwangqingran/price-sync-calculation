package cache

import (
	"context"
	"fmt"
	"time"

	"git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

const (
	localPriceCacheKey  = "-price-config-20211201" // weird key, keep same with sip-goservice
	channelWhitelistKey = "channel_whitelist_config"

	channelWhitelistExpireTimeInSec = 60
)

type systemConfigCacheManager struct {
	store cache.Cache
}

func NewSystemConfigCacheManager(store config.RedisSession) SystemConfigCacheManager {
	return &systemConfigCacheManager{
		store: store,
	}
}

func (m systemConfigCacheManager) LocalSIPConfigKey(primaryRegion, affiRegion string) string {
	return fmt.Sprintf("%s_%s_%s", localPriceCacheKey, primaryRegion, affiRegion)
}

func (m systemConfigCacheManager) GetLocalSipConfigMany(ctx context.Context, keys []string) (map[string]*model.CommonPriceConfig, error) {
	receiver := map[string]interface{}{}
	for _, key := range keys {
		receiver[key] = &model.CommonPriceConfig{}
	}
	err := m.store.GetMany(ctx, receiver)
	if err != nil {
		if err == cache.ErrCacheMiss { // cache is missed
			return nil, err
		}
		logging.GetLogger(ctx).Error("GetLocalSIPConfig cache err", ulog.Error(err))

		return nil, cerr.Wrap(err, "GetLocalSIPConfig() in systemConfigCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}
	res := make(map[string]*model.CommonPriceConfig)

	for k, v := range receiver {
		cfg, ok := v.(*model.CommonPriceConfig)
		if !ok {
			return nil, cerr.New(fmt.Sprintf("failed to assert common config price type, key=%v", k), uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
		}
		res[k] = cfg
	}
	return res, nil
}

func (m systemConfigCacheManager) GetLocalSIPConfig(ctx context.Context, key string) (*model.CommonPriceConfig, error) {
	results, err := m.GetLocalSipConfigMany(ctx, []string{key})
	if err != nil {
		return nil, err
	}
	if results[key] == nil {
		return nil, cerr.New(fmt.Sprintf("failed to get local sip config for key=%v", key), uint32(priceSyncPriceCalculationPb.Constant_ERROR_NOT_FOUND))
	}

	return results[key], nil
}

func (m systemConfigCacheManager) SetLocalSipConfigMany(ctx context.Context, kvs map[string]*model.CommonPriceConfig) error {
	if m.store == nil {
		return cerr.New("cache client in systemConfigCacheManager is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}
	if len(kvs) == 0 {
		return nil
	}
	expireTime := time.Duration(config.GetCommonConfig().LocalSIPPriceConfigCacheExpireSecondForRedis) * time.Second
	cacheKvs := map[string]interface{}{}
	for k, v := range kvs {
		cacheKvs[k] = v
	}

	err := m.store.SetMany(ctx, cacheKvs, expireTime)
	if err != nil {
		logging.GetLogger(ctx).Error("SetLocalSIPConfig cache err", ulog.Error(err))

		return cerr.Wrap(err, "SetLocalSIPConfig() in systemConfigCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}
	return nil
}

func (m systemConfigCacheManager) SetLocalSIPConfig(ctx context.Context, key string, value *model.CommonPriceConfig) error {
	kvs := make(map[string]*model.CommonPriceConfig)
	kvs[key] = value
	return m.SetLocalSipConfigMany(ctx, kvs)
}

func (m systemConfigCacheManager) ChannelWhiteListKey() string {
	return channelWhitelistKey
}

func (m systemConfigCacheManager) GetChannelWhiteList(ctx context.Context, key string) ([]int64, error) {
	if m.store == nil {
		return nil, cerr.New("cache client in systemConfigCacheManager is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	configVal := make([]int64, 0)
	err := m.store.Get(ctx, key, &configVal)
	if err != nil {
		if err == cache.ErrCacheMiss { // cache is missed
			return nil, err
		}
		logging.GetLogger(ctx).Error("GetChannelWhiteList cache err", ulog.Error(err))

		return nil, cerr.Wrap(err, "GetChannelWhiteList() in systemConfigCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	return configVal, nil
}

func (m systemConfigCacheManager) SetChannelWhiteList(ctx context.Context, key string, value []int64) error {
	if m.store == nil {
		return cerr.New("cache client in systemConfigCacheManager is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	if value == nil {
		return nil
	}

	expireTime := time.Duration(channelWhitelistExpireTimeInSec) * time.Second
	err := m.store.Set(ctx, key, value, expireTime)
	if err != nil {
		logging.GetLogger(ctx).Error("SetChannelWhiteList cache err", ulog.Error(err))

		return cerr.Wrap(err, "SetChannelWhiteList() in systemConfigCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	return nil
}
