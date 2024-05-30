package config

import (
	"context"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/common/uniconfig"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type CbscPriceConfig struct {
	CbscPriceFeeConfigMap map[string]*CBSCPriceFeeConfig `json:"cbsc_price_fee_config_map"` // merchant Region -> CBSCPriceFeeConfig
}

type CBSCPriceFeeConfig struct {
	TransactionFeeRate uint64 `json:"transaction_fee_rate"`
	UseServiceFeeRate  bool   `json:"use_service_fee_rate"`
	CommissionFeeRate  uint64 `json:"commission_fee_rate"`
	ServiceFeeRateMax  uint64 `json:"service_fee_rate_max"`
	ServiceFeeRateMin  uint64 `json:"service_fee_rate_min"`
}

func onCbscPriceConfigUpdate(e uniconfig.Event) {
	rawNewConfig, err := e.New()
	if err != nil {
		logging.GetLogger(context.Background()).Warn("error getting updated CbscPriceConfig value from uniconfig.Event", ulog.Error(err))
		return
	}

	newConfig, ok := rawNewConfig.(*CbscPriceConfig)
	if !ok {
		logging.GetLogger(context.Background()).Warn("new config is not a CbscPriceConfig",
			ulog.String("newVal", cutil.JSONEncode(rawNewConfig)))
		return
	}

	confVal.CbscPriceConfig = newConfig

	logging.GetLogger(context.Background()).Info("CbscPriceConfig is updated", ulog.String("cbsc_price_config", cutil.JSONEncode(newConfig)))
}

func GetCbscPriceFeeConfigMap() map[string]*CBSCPriceFeeConfig {
	if confVal == nil || confVal.CbscPriceConfig == nil {
		return nil
	}
	return confVal.CbscPriceConfig.CbscPriceFeeConfigMap
}

func GetCbscPriceFeeConfig(merchantRegion string) *CBSCPriceFeeConfig {
	if confVal == nil || confVal.CbscPriceConfig == nil {
		return nil
	}
	return confVal.CbscPriceConfig.CbscPriceFeeConfigMap[merchantRegion]
}
