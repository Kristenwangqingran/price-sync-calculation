package price_sync_db

import (
	"git.garena.com/shopee/common/gdbc/gdbc/tablereflect"
)

const (
	TableMerchantConstraints   = "merchant_constraints_tab"
	TableMerchantConfigSetting = "merchant_config_setting_tab"
)

type ProfitRateLimit struct {
	ID             uint64 `gdbc:"column=id"`
	Region         string `gdbc:"column=region"`
	ProfitRateMin  int64  `gdbc:"column=profit_rate_min"`
	ProfitRateMax  int64  `gdbc:"column=profit_rate_max"`
	Operator       string `gdbc:"column=operator"`
	CreateTime     uint64 `json:"ctime" gdbc:"column=ctime"`
	ModifyTime     uint64 `json:"mtime" gdbc:"column=mtime"`
	MerchantRegion string `gdbc:"column=merchant_region"`
}

type MerchantConfigSetting struct {
	ID             uint64 `gdbc:"column=id"`
	MerchantID     uint64 `gdbc:"column=merchant_id"`
	ShopId         uint64 `gdbc:"column=shop_id"`
	Region         string `gdbc:"column=region"`
	ProfitRate     uint64 `gdbc:"column=profit_rate"`
	ServiceFeeRate uint64 `gdbc:"column=service_fee_rate"`
	CreateTime     uint32 `gdbc:"column=ctime"`
	ModifyTime     uint32 `gdbc:"column=mtime"`
}

func init() {
	tablereflect.TypeInit(
		&MerchantConfigSetting{},
		tablereflect.Table(TableMerchantConfigSetting),
	)
	tablereflect.TypeInit(
		&ProfitRateLimit{},
		tablereflect.Table(TableMerchantConstraints),
	)
}
