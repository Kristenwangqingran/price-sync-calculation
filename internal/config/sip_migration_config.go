package config

import (
	"context"
	"fmt"
	"strings"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/common/uniconfig"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

// SIPMigrationConfig copy from sip-goservice config, need to keep sync with it until decouple finished
type SIPMigrationConfig struct {
	CurrencyCommonConf CurrencyCommonConf `json:"currency_common_conf"`
	RegionCommonConf   RegionCommonConf   `json:"region_common_conf"`

	LocalPriceSettingLimitConfig     map[string]LocalSipLimitConfig `json:"local_price_setting_limit_config"`
	DefaultRateTableForCbHiddenPrice []map[string]string            `json:"default_rate_table_for_cb_hidden_price"`
	HandlingFeeConfig                *int64                         `json:"handling_fee_config"`

	RegionList []string `json:"region_list"`
}

type LocalSipLimitConfig struct {
	ExchangeRateMax    float64 `json:"exchange_rate_max"`
	ExchangeRateMin    float64 `json:"exchange_rate_min"`
	InitHiddenPriceMax float64 `json:"init_hidden_price_max"`
	InitHiddenPriceMin float64 `json:"init_hidden_price_min"`
	BufferMax          float64 `json:"buffer_max"`
	BufferMin          float64 `json:"buffer_min"`
}

type CurrencyCommonConf struct {
	CommonSetting map[string]CurrencyCommonSetting `json:"common_setting"`
}

type CurrencyCommonSetting struct {
	ExchangeRateMaxLimit float64 `json:"exchange_rate_max_limit"`
	ExchangeRateMinLimit float64 `json:"exchange_rate_min_limit"`
	Precision            int     `json:"precision"`
	PrecisionForFee      int     `json:"precision_for_fee"` // TODO used for order snapshot, remove after migration
	UseSpecialRoundup    bool    `json:"use_special_roundup"`
}

func (ccc *CurrencyCommonConf) GetByCurrency(currency string) CurrencyCommonSetting {
	config, exist := ccc.CommonSetting[currency]
	if !exist {
		ulog.DefaultLogger().Error("currency common config not found", ulog.String("currency", currency))
	}
	return config
}

type RegionCommonConf struct {
	CommonSetting map[string]RegionCommonSetting `json:"common_setting"`
}

type RegionCommonSetting struct {
	Currency string `json:"currency"`
}

func (rcc *RegionCommonConf) GetByRegion(region string) RegionCommonSetting {
	config, exist := rcc.CommonSetting[strings.ToUpper(region)]
	if !exist {
		ulog.DefaultLogger().Error("region common config not found", ulog.String("region", region))
	}
	return config
}

func onSIPMigrationConfigUpdate(e uniconfig.Event) {
	rawNewConfig, err := e.New()
	if err != nil {
		logging.GetLogger(context.Background()).Warn("error getting updated SIPMigrationConfig value from uniconfig.Event", ulog.Error(err))
		return
	}

	newConfig, ok := rawNewConfig.(*SIPMigrationConfig)
	if !ok {
		logging.GetLogger(context.Background()).Warn("new config is not SIPMigrationConfig",
			ulog.String("newVal", cutil.JSONEncode(rawNewConfig)))
		return
	}

	confVal.SIPMigrationCfg = newConfig

	logging.GetLogger(context.Background()).Info("SIPMigrationConfig updated", ulog.String("updated", cutil.JSONEncode(newConfig)))
}

func GetSIPCurrencyCommonConf() *CurrencyCommonConf {
	if confVal == nil || confVal.SIPMigrationCfg == nil {
		return nil
	}

	return &confVal.SIPMigrationCfg.CurrencyCommonConf
}

func GetSIPRegionCommonConf() *RegionCommonConf {
	if confVal == nil || confVal.SIPMigrationCfg == nil {
		return nil
	}

	return &confVal.SIPMigrationCfg.RegionCommonConf
}

func GetLocalSipSettingLimitConfigByRegions(pRegion string, aRegion string) (LocalSipLimitConfig, error) {
	if confVal == nil || confVal.SIPMigrationCfg == nil {
		return LocalSipLimitConfig{}, cerr.New(fmt.Sprintf("cannof find setting limit config for pRegion=%s and aRegion=%s", pRegion, aRegion), uint32(pb.Constant_ERROR_NOT_FOUND))
	}

	key := fmt.Sprintf("%s_%s", strings.ToUpper(pRegion), strings.ToUpper(aRegion))
	v, ok := confVal.SIPMigrationCfg.LocalPriceSettingLimitConfig[key]
	if !ok {
		return LocalSipLimitConfig{}, cerr.New(fmt.Sprintf("cannof find setting limit config for pRegion=%s and aRegion=%s", pRegion, aRegion), uint32(pb.Constant_ERROR_NOT_FOUND))
	}
	return v, nil
}

func GetDefaultRateTableForCbHiddenPrice() ([]map[string]string, error) {
	if confVal == nil || confVal.SIPMigrationCfg == nil {
		return nil, cerr.New(fmt.Sprintf("cannot find default rate table for cb hidden price config"), uint32(pb.Constant_ERROR_NOT_FOUND))
	}
	return confVal.SIPMigrationCfg.DefaultRateTableForCbHiddenPrice, nil
}

func GetHandlingFeeConfig() (int64, error) {
	if confVal == nil || confVal.SIPMigrationCfg == nil || confVal.SIPMigrationCfg.HandlingFeeConfig == nil {
		return 0, cerr.New(fmt.Sprintf("cannot find handling fee config"), uint32(pb.Constant_ERROR_NOT_FOUND))
	}
	return *confVal.SIPMigrationCfg.HandlingFeeConfig, nil
}

func GetSipRegionList() []string {
	if confVal == nil || confVal.SIPMigrationCfg == nil {
		return nil
	}
	return confVal.SIPMigrationCfg.RegionList
}

func GetCurrencyByRegion(region string) (string, error) {
	conf := GetSIPRegionCommonConf()
	if conf == nil {
		return "", cerr.New(fmt.Sprintf("cannot find region common config"), uint32(pb.Constant_ERROR_NOT_FOUND))
	}

	currency := conf.GetByRegion(region).Currency
	return currency, nil
}
