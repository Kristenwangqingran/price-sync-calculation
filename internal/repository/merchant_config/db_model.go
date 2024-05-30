package merchant_config

import (
	"git.garena.com/shopee/common/gdbc/gdbc/tablereflect"
	internalMerchantConfigSettingPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
)

const (
	TableMerchantConstraints = "merchant_constraints_tab"
)

type ProfitRateLimit struct {
	ID             uint64 `gdbc:"column=id"`
	Region         string `gdbc:"column=region"`
	ProfitRateMin  int64  `gdbc:"column=profit_rate_min"`
	ProfitRateMax  int64  `gdbc:"column=profit_rate_max"`
	Operator       string `gdbc:"column=operator"`
	MerchantRegion string `gdbc:"column=merchant_region"`
}

func init() {
	tablereflect.TypeInit(
		&internalMerchantConfigSettingPb.MerchantConfigSetting{},
	)
	tablereflect.TypeInit(
		&ProfitRateLimit{},
		tablereflect.Table(TableMerchantConstraints),
	)
}
