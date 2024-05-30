package cache

import (
	"context"
	"fmt"
	"time"

	"git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

const (
	dummyBuyerUserIdByRegionCacheKey = "dummy_buyer_user_id"
)

type logisticCacheManager struct {
	store cache.Cache
}

func NewLogisticCacheManager(store config.RedisSession) LogisticCacheManager {
	return &logisticCacheManager{
		store: store,
	}
}

func (m *logisticCacheManager) DummyBuyerUserIdKey(region string) string {
	return fmt.Sprintf("%s:%s", dummyBuyerUserIdByRegionCacheKey, region)
}

func (m *logisticCacheManager) GetDummyBuyerUserId(ctx context.Context, key string) (int64, error) {
	if m.store == nil {
		return 0, cerr.New("cache client in logisticCacheManager is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	var configVal int64
	err := m.store.Get(ctx, key, &configVal)
	if err != nil {
		if err == cache.ErrCacheMiss { // cache is missed
			return 0, err
		}
		logging.GetLogger(ctx).Error("GetDummyBuyerUserId cache err", ulog.Error(err))

		return 0, cerr.Wrap(err, "GetDummyBuyerUserId() in logisticCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	return configVal, nil
}

func (m *logisticCacheManager) SetDummyBuyerUserId(ctx context.Context, key string, value int64) error {
	if m.store == nil {
		return cerr.New("cache client in logisticCacheManager is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	expireTime := time.Duration(config.GetCommonConfig().DummyBuyerUserIdRemoteCacheExpireSeconds) * time.Second
	err := m.store.Set(ctx, key, &value, expireTime)
	if err != nil {
		logging.GetLogger(ctx).Error("SetDummyBuyerUserId cache err", ulog.Error(err))

		return cerr.Wrap(err, "SetDummyBuyerUserId() in logisticCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	return nil
}
