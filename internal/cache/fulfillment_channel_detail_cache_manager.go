package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/vmihailenco/msgpack"

	"git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	internalFulfillmentChannelPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_fulfillment_channel.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/serverutil"
)

const FulfillmentChannelDetailCachePrefix = "GetShopChannelIdFromFulfillment"

type FulfillmentChannelDetailCacheManagerImpl struct {
	store cache.Cache
}

func NewFulfillmentChannelDetailCacheManager(store config.RedisSession) FulfillmentChannelDetailCacheManager {
	return &FulfillmentChannelDetailCacheManagerImpl{
		store: store,
	}
}

func (m *FulfillmentChannelDetailCacheManagerImpl) Key(shopId uint64) string {
	key := FulfillmentChannelDetailCachePrefix + strconv.FormatUint(shopId, 10) + serverutil.GetEnv()
	return getFullCacheKey(config.GetRedisCacheKeyPrefix(), key)
}

func (m *FulfillmentChannelDetailCacheManagerImpl) Get(ctx context.Context, key string) ([]*internalFulfillmentChannelPb.FulfillmentChannelDetail, error) {
	if m.store == nil {
		return nil, cerr.New("cache client in FulfillmentChannelDetailCacheManagerImpl is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	var data []byte
	var fulfillmentChannelDetailList []*internalFulfillmentChannelPb.FulfillmentChannelDetail
	err := m.store.Get(ctx, key, &data)
	if err != nil {
		if err == cache.ErrCacheMiss { // cache is missed
			return nil, err
		}
		return nil, cerr.Wrap(err, "Get() in FulfillmentChannelDetailCacheManagerImpl is failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	err = msgpack.Unmarshal(data, &fulfillmentChannelDetailList)
	if err != nil {
		return nil, cerr.Wrap(err, "failed to unmarshal the result from cache in FulfillmentChannelDetailCacheManagerImpl",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	logging.GetLogger(ctx).Debug(fmt.Sprintf(
		"success to fetch fulfillmentChannelDetailList from cache, key=%s, data=%v",
		key, cutil.JSONEncode(fulfillmentChannelDetailList)))

	return fulfillmentChannelDetailList, nil
}

func (m *FulfillmentChannelDetailCacheManagerImpl) Set(ctx context.Context, key string, value []*internalFulfillmentChannelPb.FulfillmentChannelDetail) error {
	if m.store == nil {
		return cerr.New("cache client in FulfillmentChannelDetailCacheManagerImpl is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	if value == nil {
		return nil
	}

	data, err := msgpack.Marshal(value)
	if err != nil {
		return cerr.Wrap(err, "failed to marshal in FulfillmentChannelDetailCacheManagerImpl Set()",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	expireTime := time.Duration(config.GetCommonConfig().FulfillmentChannelDetailCacheExpireSecond) * time.Second
	err = m.store.Set(ctx, key, data, expireTime)
	if err != nil {
		return cerr.Wrap(err, "failed in FulfillmentChannelDetailCacheManagerImpl Set()",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	logging.GetLogger(ctx).Debug(fmt.Sprintf(
		"success to set fulfillmentChannelDetailList into cache, key=%s, data=%v",
		key, cutil.JSONEncode(value)))

	return nil
}
