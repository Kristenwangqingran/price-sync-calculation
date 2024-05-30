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

type shopDataCacheManager struct {
	store      cache.Cache
	localStore cache.Cache
}

type ShopSellerWarehouseInfo struct {
	IsWarehouse    bool
	Is3PFWarehouse bool
}

func (m *shopDataCacheManager) ShopDetailKey(shopId int64) string {
	return fmt.Sprintf("shop_info_cache_%d", shopId)
}

func (m *shopDataCacheManager) UserIdKey(shopId int64) string {
	return fmt.Sprintf("user_id_of_%d", shopId)
}

func (m *shopDataCacheManager) isSellerWarehouseShopKey(shopId int64) string {
	return fmt.Sprintf("is_seller_warehouse_shop_%d", shopId)
}

func (m *shopDataCacheManager) GetShopDetail(ctx context.Context, key string) (*model.ShopDetail, error) {
	if m.store == nil {
		return nil, cerr.New("cache client in systemConfigCacheManager is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	var configVal model.ShopDetail
	err := m.store.Get(ctx, key, &configVal)
	if err != nil {
		if err == cache.ErrCacheMiss { // cache is missed
			return nil, err
		}
		logging.GetLogger(ctx).Error("GetShopDetail cache err", ulog.Error(err))

		return nil, cerr.Wrap(err, "GetShopDetail() in shopDataCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	return &configVal, nil
}

func (m *shopDataCacheManager) SetShopDetail(ctx context.Context, key string, value *model.ShopDetail) error {
	if m.store == nil {
		return cerr.New("cache client in shopDataCacheManager is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	if value == nil {
		return nil
	}

	expireTime := time.Duration(config.GetCommonConfig().ShopDetailRemoteCacheExpireSeconds) * time.Second
	err := m.store.Set(ctx, key, value, expireTime)
	if err != nil {
		logging.GetLogger(ctx).Error("SetShopDetail cache err", ulog.Error(err))

		return cerr.Wrap(err, "SetShopDetail() in shopDataCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	return nil
}

func (m *shopDataCacheManager) GetUserIdByShopIdFromLocalCache(ctx context.Context, shopId int64) (int64, error) {
	if m.localStore == nil {
		return 0, cerr.New("cache client in systemConfigCacheManager is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	var configVal int64
	err := m.localStore.Get(ctx, m.UserIdKey(shopId), &configVal)
	if err != nil {
		if err == cache.ErrCacheMiss { // cache is missed
			return 0, err
		}
		logging.GetLogger(ctx).Error("GetUserIdByShopIdFromLocalCache cache err", ulog.Error(err))

		return 0, cerr.Wrap(err, "GetShopDetail() in shopDataCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	return configVal, nil
}

func (m *shopDataCacheManager) SetUserIdOfShopIdToLocalCache(ctx context.Context, shopId int64, userId int64) error {
	if m.localStore == nil {
		return cerr.New("cache client in shopDataCacheManager is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	expireTime := time.Duration(config.GetCommonConfig().UserIdOfShopIdLocalCacheExpireSeconds) * time.Second
	err := m.localStore.Set(ctx, m.UserIdKey(shopId), userId, expireTime)
	if err != nil {
		logging.GetLogger(ctx).Error("SetUserIdOfShopIdToLocalCache err", ulog.Error(err))

		return cerr.Wrap(err, "SetUserIdOfShopIdToLocalCache() in shopDataCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	return nil
}

func (m *shopDataCacheManager) CheckIsSellerWarehouseShopFromLocalCache(ctx context.Context, shopId int64) (*ShopSellerWarehouseInfo, error) {
	if m.localStore == nil {
		return nil, cerr.New("cache client in systemConfigCacheManager is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	var configVal ShopSellerWarehouseInfo
	err := m.localStore.Get(ctx, m.isSellerWarehouseShopKey(shopId), &configVal)
	if err != nil {
		if err == cache.ErrCacheMiss { // cache is missed
			return nil, nil
		}
		logging.GetLogger(ctx).Error(fmt.Sprintf("CheckIsSellerWarehouseShopFromLocalCache cache err=%v", err))

		return nil, cerr.Wrap(err, "CheckIsSellerWarehouseShopFromLocalCache() in shopDataCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	return &configVal, nil
}

func (m *shopDataCacheManager) SetIsSellerWarehouseShopToLocalCache(ctx context.Context, shopId int64, shopSellerWarehouseInfo *ShopSellerWarehouseInfo) error {
	if m.localStore == nil {
		return cerr.New("cache client in shopDataCacheManager is not initialized",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	expireTime := time.Duration(config.GetCommonConfig().IsSellerWarehouseShopLocalCacheExpireSeconds) * time.Second
	err := m.localStore.Set(ctx, m.isSellerWarehouseShopKey(shopId), shopSellerWarehouseInfo, expireTime)
	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf("SetIsSellerWarehouseShopToLocalCache err=%v", err))

		return cerr.Wrap(err, "SetIsSellerWarehouseShopToLocalCache() in shopDataCacheManager failed",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CACHE))
	}

	return nil
}

func NewShopDataCacheManager(store config.RedisSession) ShopDataCacheManager {
	return &shopDataCacheManager{
		store:      store,
		localStore: config.GetLocalCacheForShopClient(),
	}
}
