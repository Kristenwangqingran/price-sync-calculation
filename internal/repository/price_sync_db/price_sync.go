package price_sync_db

import (
	"context"

	internal_merchant_config_setting "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
	internal_merchant_constraints "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_constraints.pb"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/orm"
)

type MerchantConfigSettingRepo interface {
	orm.DbSessionFactory
	GetMerchantConfigSettingList(ctx context.Context, session orm.DbSession, merchantId uint64) ([]*internal_merchant_config_setting.MerchantConfigSetting, error)
	UpsertMerchantConfigSettingList(ctx context.Context, session orm.DbSession, settings []*internal_merchant_config_setting.MerchantConfigSetting) error
}

type MerchantConstraintsRepo interface {
	orm.DbSessionFactory
	GetProfitRateLimitListByMerchantRegion(ctx context.Context, session orm.DbSession, region string, merchantRegion string) ([]*internal_merchant_constraints.MerchantConstraints, error)
	UpdateProfitRateLimitByRegionMerchantRegion(ctx context.Context, session orm.DbSession, region, merchantRegion string, profitRateMin, profitRateMax *float64, operator string) error
}
