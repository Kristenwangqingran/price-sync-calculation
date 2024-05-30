package cache

import (
	"context"
	"fmt"
	"time"

	"git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	internalExchangeRatePb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_exchange_rate.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/serverutil"

	"github.com/vmihailenco/msgpack"
)

const ExchangeRateCacheKeyFormat = "SellerListingService:%s:GetCurrencyExchangeRate:%d"

type ExchangeRateCacheManagerImpl struct {
	store cache.Cache
}

func NewExchangeRateCacheManager(store config.RedisSession) ExchangeRateCacheManager {
	return &ExchangeRateCacheManagerImpl{
		store: store,
	}
}

func (m *ExchangeRateCacheManagerImpl) Key(merchantId uint64) string {
	key := fmt.Sprintf(ExchangeRateCacheKeyFormat, serverutil.GetEnv(), merchantId)
	return getFullCacheKey(config.GetRedisCacheKeyPrefix(), key)
}

func (m *ExchangeRateCacheManagerImpl) Get(ctx context.Context, key string) (*internalExchangeRatePb.ExchangeRateInfo, error) {
	if m.store == nil {
		return nil, cerr.New("cache client in ExchangeRateCacheManagerImpl is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	var data []byte
	err := m.store.Get(ctx, key, &data)
	if err != nil {
		if err == cache.ErrCacheMiss { // cache is missed
			return nil, err
		}
		return nil, cerr.Wrap(err, "Get() in ExchangeRateCacheManagerImpl is failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	var exchangeRateInfo *internalExchangeRatePb.ExchangeRateInfo
	err = msgpack.Unmarshal(data, &exchangeRateInfo)
	if err != nil {
		return nil, cerr.Wrap(err, "failed to unmarshal the result from cache in ExchangeRateCacheManagerImpl",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	logging.GetLogger(ctx).Debug(fmt.Sprintf(
		"success to fetch exchangeRateInfo from cache, key=%s, data=%v",
		key, cutil.JSONEncode(exchangeRateInfo)))

	return exchangeRateInfo, nil
}

func (m *ExchangeRateCacheManagerImpl) Set(ctx context.Context, key string, value *internalExchangeRatePb.ExchangeRateInfo) error {
	if m.store == nil {
		return cerr.New("cache client in ExchangeRateCacheManagerImpl is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	if value == nil {
		return nil
	}

	data, err := msgpack.Marshal(value)
	if err != nil {
		return cerr.Wrap(err, "failed to marshal in ExchangeRateCacheManagerImpl Set()",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	expireTime := time.Duration(config.GetCommonConfig().ExchangeRateInfoCacheExpireSecond) * time.Second
	err = m.store.Set(ctx, key, data, expireTime)
	if err != nil {
		return cerr.Wrap(err, "failed in ExchangeRateCacheManagerImpl Set()",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	logging.GetLogger(ctx).Debug(fmt.Sprintf(
		"success to set exchangeRateInfo into cache, key=%s, data=%v",
		key, cutil.JSONEncode(value)))

	return nil
}
