package cache

import (
	"context"
	"time"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"

	internalExchangeRatePb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_exchange_rate.pb"
	internalFulfillmentChannelPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_fulfillment_channel.pb"
	internalMerchantConfigSettingPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
)

type FulfillmentChannelDetailCacheManager interface {
	// Key will return cache key.
	Key(shopId uint64) string

	// Get will retrieve data from cache.
	// If not found, ErrCacheMiss will be returned.
	Get(ctx context.Context, key string) ([]*internalFulfillmentChannelPb.FulfillmentChannelDetail, error)

	// Set will store the provided data to the cache.
	Set(ctx context.Context, key string, value []*internalFulfillmentChannelPb.FulfillmentChannelDetail) error
}

type MerchantConfigSettingCacheManager interface {
	// Key will return cache key.
	Key(merchantId uint64) string

	// Get will retrieve data from cache.
	// If not found, ErrCacheMiss will be returned.
	Get(ctx context.Context, key string) ([]*internalMerchantConfigSettingPb.MerchantConfigSetting, error)

	// Set will store the provided data to the cache.
	Set(ctx context.Context, key string, value []*internalMerchantConfigSettingPb.MerchantConfigSetting, expireTime time.Duration) error

	Del(ctx context.Context, key string) error
}

type ExchangeRateCacheManager interface {
	// Key will return cache key.
	Key(merchantId uint64) string

	// Get will retrieve data from cache.
	// If not found, ErrCacheMiss will be returned.
	Get(ctx context.Context, key string) (*internalExchangeRatePb.ExchangeRateInfo, error)

	// Set will store the provided data to the cache.
	Set(ctx context.Context, key string, value *internalExchangeRatePb.ExchangeRateInfo) error
}

type CommissionRateCacheManager interface {
	// Key will return cache key.
	Key(shopId uint64) string

	// Get will retrieve data from cache.
	// If not found, ErrCacheMiss will be returned.
	Get(ctx context.Context, key string) (uint64, error)

	// Set will store the provided data to the cache.
	Set(ctx context.Context, key string, value uint64) error
}

type SystemConfigCacheManager interface {
	LocalSIPConfigKey(primaryRegion, affiRegion string) string

	GetLocalSIPConfig(ctx context.Context, key string) (*model.CommonPriceConfig, error)

	SetLocalSIPConfig(ctx context.Context, key string, value *model.CommonPriceConfig) error

	ChannelWhiteListKey() string

	GetChannelWhiteList(ctx context.Context, key string) ([]int64, error)

	SetChannelWhiteList(ctx context.Context, key string, value []int64) error
}

type ShopDataCacheManager interface {
	ShopDetailKey(shopId int64) string

	GetShopDetail(ctx context.Context, key string) (*model.ShopDetail, error)
	SetShopDetail(ctx context.Context, key string, value *model.ShopDetail) error

	GetUserIdByShopIdFromLocalCache(ctx context.Context, shopId int64) (int64, error)
	SetUserIdOfShopIdToLocalCache(ctx context.Context, shopId int64, userId int64) error

	CheckIsSellerWarehouseShopFromLocalCache(ctx context.Context, shopId int64) (*ShopSellerWarehouseInfo, error)
	SetIsSellerWarehouseShopToLocalCache(ctx context.Context, shopId int64, shopSellerWarehouseInfo *ShopSellerWarehouseInfo) error
}

type LogisticCacheManager interface {
	DummyBuyerUserIdKey(region string) string

	GetDummyBuyerUserId(ctx context.Context, key string) (int64, error)
	SetDummyBuyerUserId(ctx context.Context, key string, value int64) error
}
