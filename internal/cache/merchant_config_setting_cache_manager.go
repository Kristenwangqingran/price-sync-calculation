package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	internalMerchantConfigSettingPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/serverutil"
)

type MerchantConfigSettingLocalCacheManager MerchantConfigSettingCacheManager
type MerchantConfigSettingRedisCacheManager MerchantConfigSettingCacheManager

// MerchantConfigSettingCachePrefix use different cache prefix with current listing service
const MerchantConfigSettingCachePrefix = "price-MerchantConfigSetting-"

type MerchantConfigSettingCacheManagerImpl struct {
	store cache.Cache
}

func NewMerchantConfigSettingLocalCacheManager(store config.LocalCacheSession) MerchantConfigSettingLocalCacheManager {
	return &MerchantConfigSettingCacheManagerImpl{
		store: store,
	}
}

func NewMerchantConfigSettingRedisCacheManager(store config.RedisSession) MerchantConfigSettingRedisCacheManager {
	return &MerchantConfigSettingCacheManagerImpl{
		store: store,
	}
}

func (m *MerchantConfigSettingCacheManagerImpl) Key(merchantId uint64) string {
	key := MerchantConfigSettingCachePrefix + strconv.FormatUint(merchantId, 10) + serverutil.GetEnv()
	return getFullCacheKey(config.GetRedisCacheKeyPrefix(), key)
}

func (m *MerchantConfigSettingCacheManagerImpl) Get(ctx context.Context, key string) ([]*internalMerchantConfigSettingPb.MerchantConfigSetting, error) {
	if m.store == nil {
		return nil, cerr.New("cache client in MerchantConfigSettingCacheManagerImpl is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	var data []*internalMerchantConfigSettingPb.MerchantConfigSetting
	err := m.store.Get(ctx, key, &data)
	if err != nil {
		if err == cache.ErrCacheMiss { // cache is missed
			return nil, err
		}
		return nil, cerr.Wrap(err, "Get() in MerchantConfigSettingCacheManagerImpl is failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	logging.GetLogger(ctx).Debug(fmt.Sprintf(
		"success to fetch merchantConfigSettingList from cache, key=%s, data=%v",
		key, cutil.JSONEncode(data)))

	return data, nil
}

func (m *MerchantConfigSettingCacheManagerImpl) Set(ctx context.Context, key string, value []*internalMerchantConfigSettingPb.MerchantConfigSetting, expireTime time.Duration) error {
	if m.store == nil {
		return cerr.New("cache client in MerchantConfigSettingCacheManagerImpl is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	if value == nil {
		return nil
	}

	err := m.store.Set(ctx, key, value, expireTime)
	if err != nil {
		return cerr.Wrap(err, "failed in MerchantConfigSettingCacheManagerImpl Set()",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	logging.GetLogger(ctx).Debug(fmt.Sprintf(
		"success to set merchantConfigSettingList into cache, key=%s, data=%v",
		key, cutil.JSONEncode(value)))

	return nil
}

func (m *MerchantConfigSettingCacheManagerImpl) Del(ctx context.Context, key string) error {
	if m.store == nil {
		return cerr.New("cache client in MerchantConfigSettingCacheManagerImpl is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	err := m.store.Delete(ctx, key)
	if err != nil {
		return cerr.Wrap(err, "failed in MerchantConfigSettingCacheManagerImpl Delete()",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	logging.GetLogger(ctx).Debug(fmt.Sprintf(
		"success to del cache in merchantConfigSettingList cache, key=%s", key))

	return nil
}
