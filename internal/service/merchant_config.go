package service

import (
	"context"

	internalMerchantConfigSettingPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_constraints.pb"
)

type MerchantConfigService interface {
	// GetMerchantConfigSettingInfoMap fetch merchant config map by merchant ids.
	// If failed to get for one merchant id, then only record error in the response and continue to handle remaining part.
	GetMerchantConfigSettingInfoMap(ctx context.Context, merchantIdMap []uint64) map[uint64]*MerchantConfigSettingInfo
	GetSingleMerchantConfigSettingInfoMap(ctx context.Context, merchantId uint64) (map[uint64]*internalMerchantConfigSettingPb.MerchantConfigSetting, error)
	SetMerchantConfigSettings(ctx context.Context,
		merchantId uint64, settings []*internalMerchantConfigSettingPb.MerchantConfigSetting) error
	GetProfitRateLimit(ctx context.Context, region, merchantRegion string) ([]*internal.MerchantConstraints, error)
	UpdateProfitRateLimit(ctx context.Context, region, merchantRegion string, profitRateMin, profitRateMax *float64, operator string) error
}

type MerchantConfigSettingInfo struct {
	Err error
	// shopId => setting
	MerchantConfigSettingMap map[uint64]*internalMerchantConfigSettingPb.MerchantConfigSetting
}
