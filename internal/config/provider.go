package config

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	GetListingDbClient,
	GetPriceSyncDBClient,
	GetSipDbClient,
	GetSipV2DbClient,
	GetAuditLogDBClient,
	GetLocalCacheForMerchantConfigClient,
	GetRedisClient,
)
