package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack"

	"git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/serverutil"
)

const CommissionRateCacheKeyFormat = "SellerListingService:%s:GetCommissionRate:%d"

type CommissionRateCacheManagerImpl struct {
	store cache.Cache
}

func NewCommissionRateCacheManager(store config.RedisSession) CommissionRateCacheManager {
	return &CommissionRateCacheManagerImpl{
		store: store,
	}
}

func (m *CommissionRateCacheManagerImpl) Key(shopId uint64) string {
	key := fmt.Sprintf(CommissionRateCacheKeyFormat, serverutil.GetEnv(), shopId)
	return getFullCacheKey(config.GetRedisCacheKeyPrefix(), key)
}

func (m *CommissionRateCacheManagerImpl) Get(ctx context.Context, key string) (uint64, error) {
	if m.store == nil {
		return 0, cerr.New("cache client in CommissionRateCacheManagerImpl is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	var data []byte
	var commissionRateInfo uint64
	err := m.store.Get(ctx, key, &data)
	if err != nil {
		if err == cache.ErrCacheMiss { // cache is missed
			return 0, err
		}
		return 0, cerr.Wrap(err, "Get() in CommissionRateCacheManagerImpl is failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	err = msgpack.Unmarshal(data, &commissionRateInfo)
	if err != nil {
		return 0, cerr.Wrap(err, "failed to unmarshal the result from cache in CommissionRateCacheManagerImpl",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	logging.GetLogger(ctx).Debug(fmt.Sprintf(
		"success to fetch commissionRateInfo from cache, key=%s, data=%v",
		key, commissionRateInfo))

	return commissionRateInfo, nil
}

func (m *CommissionRateCacheManagerImpl) Set(ctx context.Context, key string, value uint64) error {
	if m.store == nil {
		return cerr.New("cache client in CommissionRateCacheManagerImpl is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	data, err := msgpack.Marshal(value)
	if err != nil {
		return cerr.Wrap(err, "failed to marshal in CommissionRateCacheManagerImpl Set()",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	expireTime := time.Duration(config.GetCommonConfig().CommissionRateCacheExpireSecond) * time.Second
	err = m.store.Set(ctx, key, data, expireTime)
	if err != nil {
		errMsg := fmt.Sprintf("falied to set CommissionRateCache, err=%s, key=%s, value=%d",
			err.Error(), key, value)
		logging.GetLogger(ctx).Error(errMsg)
		return cerr.Wrap(err, errMsg,
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	logging.GetLogger(ctx).Debug(fmt.Sprintf(
		"success to set commissionRateInfo into cache, key=%s, data=%v",
		key, value))

	return nil
}
