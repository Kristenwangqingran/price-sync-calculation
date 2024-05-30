package cache

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewShopDataCacheManager,
	NewCommissionRateCacheManager,
	NewExchangeRateCacheManager,
	NewFulfillmentChannelDetailCacheManager,
	NewLogisticCacheManager,
	NewMerchantConfigSettingLocalCacheManager,
	NewMerchantConfigSettingRedisCacheManager,
	NewSystemConfigCacheManager,
	NewCommonCacheImpl,
	wire.Bind(new(CommonCache), new(*CommonCacheImpl)),
)
