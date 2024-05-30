package config

import (
	"context"
	"strings"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/common/uniconfig"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

const (
	defaultFulfillmentChannelDetailCacheExpireSecond           = 20 * 60 // 20 minutes
	defaultMerchantConfigSettingCacheExpireSecondForLocalCache = 1 * 60  // 1 minutes
	defaultMerchantConfigSettingCacheExpireSecondForRedis      = 100     // 100 seconds
	defaultExchangeRateInfoCacheExpireSecond                   = 60      // 60 seconds
	defaultCommissionRateCacheExpireSecond                     = 10 * 60 // 10 minutes
	defaultLocalSIPPriceConfigCacheExpireSecondForRedis        = 60
	defaultPricePrecision                                      = 2
	defaultDummyBuyerUserIdRemoteCacheExpireSeconds            = 30 * 60
	defaultShopDetailRemoteCacheExpireSeconds                  = 60
	defaultRegionChannelInfoCacheExpireSeconds                 = 1 * 60 * 60

	defaultOrderMartExchangeRateRefreshSeconds = 18 * 60 // 18 minutes
	defaultOrderMartExchangeRateRetrySeconds   = 1 * 60  // 1 minute

	expireTime10Minutes = 600
	expireTime5Minutes  = 300
	expireTime1Minute   = 60
	expireTime2Hours    = 2 * 60 * 60
)

// CommonConfig contains configures that is applicable at the service-level
type CommonConfig struct {
	CmdIgnoreReqRespInLogList []CmdIgnoreReqRespInLog `json:"cmd_ignore_req_resp_in_log_list"`

	// cache expired time
	FulfillmentChannelDetailCacheExpireSecond           int32 `json:"fulfillment_channel_detail_cache_expire_second"`
	MerchantConfigSettingCacheExpireSecondForLocalCache int32 `json:"merchant_config_setting_cache_expire_second_for_local_cache"`
	MerchantConfigSettingCacheExpireSecondForRedis      int32 `json:"merchant_config_setting_cache_expire_second_for_redis"`
	ExchangeRateInfoCacheExpireSecond                   int32 `json:"exchange_rate_info_cache_expire_second"`
	CommissionRateCacheExpireSecond                     int32 `json:"commission_rate_cache_expire_second"`
	ReferenceServiceFeeRateCacheExpireSecond            int32 `json:"reference_service_fee_rate_cache_expire_second"`
	LocalSIPPriceConfigCacheExpireSecondForRedis        int32 `json:"local_sip_price_config_cache_expire_second_for_redis"`
	CbSipShopFeeCacheExpireSeconds                      int32 `json:"cb_sip_shop_fee_cache_expire_seconds"`
	CbscProfitRateLimitCacheExpireSeconds               int32 `json:"cbsc_profit_rate_limit_cache_expire_seconds"`
	CnscShopsLocalCacheExpireSeconds                    int32 `json:"cnsc_shops_cache_expire_seconds"`
	CnscShopsRemoteCacheExpireSeconds                   int32 `json:"cnsc_shops_remote_cache_expire_seconds"`
	MerchantShopListLocalCacheExpireSeconds             int32 `json:"merchant_shop_list_local_cache_expire_seconds"`
	MerchantShopListRemoteCacheExpireSeconds            int32 `json:"merchant_shop_list_remote_cache_expire_seconds"`
	ShopDetailRemoteCacheExpireSeconds                  int32 `json:"shop_detail_remote_cache_expire_seconds"`
	DummyBuyerUserIdRemoteCacheExpireSeconds            int32 `json:"dummy_buyer_user_id_remote_cache_expire_seconds"`
	UserStatusMapLocalCacheExpireSeconds                int32 `json:"user_status_map_local_cache_expire_seconds"`
	UserStatusMapRemoteCacheExpireSeconds               int32 `json:"user_status_map_remote_cache_expire_seconds"`
	ShopMapCacheExpireSeconds                           int32 `json:"shop_map_cache_expire_seconds"`
	MstShopInfoExpireSeconds                            int32 `json:"mst_shop_info_expire_seconds"`
	RegionChannelInfoCacheExpireSeconds                 int32 `json:"region_channel_info_cache_expire_seconds"`
	UserIdOfShopIdLocalCacheExpireSeconds               int32 `json:"user_id_of_shop_id_local_cache_expire_seconds"`
	IsSellerWarehouseShopLocalCacheExpireSeconds        int32 `json:"is_seller_warehouse_shop_local_cache_expire_seconds"`

	OrderMartExchangeRateRefreshSeconds int32 `json:"order_mart_exchange_rate_refresh_seconds"`
	OrderMartExchangeRateRetrySeconds   int32 `json:"order_mart_exchange_rate_retry_seconds"`

	// precision based on region or currency
	PricePrecisionMap map[string]int32 `json:"price_precision"`

	DeprecatedRegions map[string]bool `json:"deprecated_regions"`

	//shopee_seller_listing_db decouple
	//TODO: remove the toggles
	MerchantConstraintsTabMigrationToggle   *SellerListingDBMigrationConfig `json:"merchant_constraints_tab_migration_toggle"`
	MerchantConfigSettingTabMigrationToggle *SellerListingDBMigrationConfig `json:"merchant_config_setting_tab_migration_toggle"`

	CBShopMarginLimit    *ShopMarginLimit `json:"cb_shop_margin_limit"`
	LocalShopMarginLimit *ShopMarginLimit `json:"local_shop_margin_limit"`
}

type CmdIgnoreReqRespInLog struct {
	Command    string `json:"command"`
	IgnoreReq  bool   `json:"ignore_req"`
	IgnoreResp bool   `json:"ignore_resp"`
}

type SellerListingDBMigrationConfig struct {
	ReadOnly    bool `json:"read_only"`
	UseNewTable bool `json:"use_new_table"`
}

type ShopMarginLimit struct {
	MinAShopMargin int32 `json:"min_shop_margin"` //inclusive
	MaxAShopMargin int32 `json:"max_shop_margin"` //exclusive
}

type ItemMarginLimit struct {
	MinAItemMargin int32 `json:"min_a_item_margin"` //inclusive
	MaxAItemMargin int32 `json:"max_a_item_margin"` //exclusive
}

var defaultCBShopMarginLimit = &ShopMarginLimit{
	MinAShopMargin: -1,
	MaxAShopMargin: 3,
}

var defaultLocalShopMarginLimit = &ShopMarginLimit{
	MinAShopMargin: 1,
	MaxAShopMargin: 11,
}

func onCommonConfigUpdate(e uniconfig.Event) {
	rawNewConfig, err := e.New()
	if err != nil {
		logging.GetLogger(context.Background()).Warn("error getting updated common config value from uniconfig.Event", ulog.Error(err))
		return
	}

	newConfig, ok := rawNewConfig.(*CommonConfig)
	if !ok {
		logging.GetLogger(context.Background()).Warn("new config is not a CommonConfig",
			ulog.String("newVal", cutil.JSONEncode(rawNewConfig)))
		return
	}

	confVal.CommonCfg = newConfig
	applyCommonConfig()

	logging.GetLogger(context.Background()).Info("common config updated", ulog.String("common_config", cutil.JSONEncode(newConfig)))
}

func applyCommonConfig() {
	if confVal == nil || confVal.CommonCfg == nil {
		logging.GetLogger(context.Background()).Warn("common config is empty")
		return
	}

	applyDefaultValForCommonConfig()
}

func applyDefaultValForCommonConfig() {
	if confVal == nil {
		confVal = &Config{}
	}

	if confVal.CommonCfg == nil {
		confVal.CommonCfg = &CommonConfig{}
	}

	commonCfg := confVal.CommonCfg
	if commonCfg.FulfillmentChannelDetailCacheExpireSecond == 0 {
		commonCfg.FulfillmentChannelDetailCacheExpireSecond = defaultFulfillmentChannelDetailCacheExpireSecond
	}

	if commonCfg.MerchantConfigSettingCacheExpireSecondForLocalCache == 0 {
		commonCfg.MerchantConfigSettingCacheExpireSecondForLocalCache = defaultMerchantConfigSettingCacheExpireSecondForLocalCache
	}

	if commonCfg.MerchantConfigSettingCacheExpireSecondForRedis == 0 {
		commonCfg.MerchantConfigSettingCacheExpireSecondForRedis = defaultMerchantConfigSettingCacheExpireSecondForRedis
	}

	if commonCfg.ExchangeRateInfoCacheExpireSecond == 0 {
		commonCfg.ExchangeRateInfoCacheExpireSecond = defaultExchangeRateInfoCacheExpireSecond
	}

	if commonCfg.CommissionRateCacheExpireSecond == 0 {
		commonCfg.CommissionRateCacheExpireSecond = defaultCommissionRateCacheExpireSecond
	}

	if commonCfg.LocalSIPPriceConfigCacheExpireSecondForRedis == 0 {
		commonCfg.LocalSIPPriceConfigCacheExpireSecondForRedis = defaultLocalSIPPriceConfigCacheExpireSecondForRedis
	}

	if commonCfg.ShopDetailRemoteCacheExpireSeconds == 0 {
		commonCfg.ShopDetailRemoteCacheExpireSeconds = defaultShopDetailRemoteCacheExpireSeconds
	}

	if commonCfg.DummyBuyerUserIdRemoteCacheExpireSeconds == 0 {
		commonCfg.DummyBuyerUserIdRemoteCacheExpireSeconds = defaultDummyBuyerUserIdRemoteCacheExpireSeconds
	}

	if commonCfg.ReferenceServiceFeeRateCacheExpireSecond == 0 {
		commonCfg.ReferenceServiceFeeRateCacheExpireSecond = expireTime10Minutes
	}

	if commonCfg.CbSipShopFeeCacheExpireSeconds == 0 {
		commonCfg.CbSipShopFeeCacheExpireSeconds = expireTime10Minutes
	}

	if commonCfg.CbscProfitRateLimitCacheExpireSeconds == 0 {
		commonCfg.CbscProfitRateLimitCacheExpireSeconds = expireTime10Minutes
	}

	if commonCfg.CnscShopsLocalCacheExpireSeconds == 0 {
		commonCfg.CnscShopsLocalCacheExpireSeconds = expireTime1Minute
	}

	if commonCfg.CnscShopsRemoteCacheExpireSeconds == 0 {
		commonCfg.CnscShopsRemoteCacheExpireSeconds = expireTime10Minutes
	}

	if commonCfg.MerchantShopListLocalCacheExpireSeconds == 0 {
		commonCfg.MerchantShopListLocalCacheExpireSeconds = expireTime1Minute
	}

	if commonCfg.MerchantShopListRemoteCacheExpireSeconds == 0 {
		commonCfg.MerchantShopListRemoteCacheExpireSeconds = expireTime10Minutes
	}

	if commonCfg.UserStatusMapLocalCacheExpireSeconds == 0 {
		commonCfg.UserStatusMapLocalCacheExpireSeconds = expireTime1Minute
	}

	if commonCfg.UserStatusMapRemoteCacheExpireSeconds == 0 {
		commonCfg.UserStatusMapRemoteCacheExpireSeconds = expireTime5Minutes
	}

	if commonCfg.ShopMapCacheExpireSeconds == 0 {
		commonCfg.ShopMapCacheExpireSeconds = expireTime1Minute
	}

	if commonCfg.MstShopInfoExpireSeconds == 0 {
		commonCfg.MstShopInfoExpireSeconds = expireTime2Hours
	}

	if commonCfg.RegionChannelInfoCacheExpireSeconds <= 0 {
		commonCfg.RegionChannelInfoCacheExpireSeconds = defaultRegionChannelInfoCacheExpireSeconds
	}

	if commonCfg.OrderMartExchangeRateRefreshSeconds <= 0 {
		commonCfg.OrderMartExchangeRateRefreshSeconds = defaultOrderMartExchangeRateRefreshSeconds
	}

	if commonCfg.OrderMartExchangeRateRetrySeconds <= 0 {
		commonCfg.OrderMartExchangeRateRetrySeconds = defaultOrderMartExchangeRateRetrySeconds
	}

	if commonCfg.CBShopMarginLimit == nil {
		commonCfg.CBShopMarginLimit = defaultCBShopMarginLimit
	}

	if commonCfg.LocalShopMarginLimit == nil {
		commonCfg.LocalShopMarginLimit = defaultLocalShopMarginLimit
	}
}

func GetCommonConfig() *CommonConfig {
	if confVal == nil {
		return nil
	}
	return confVal.CommonCfg
}

func GetCmdIgnoreReqRespInLogListConfig() []CmdIgnoreReqRespInLog {
	if GetCommonConfig() == nil {
		return nil
	}
	return GetCommonConfig().CmdIgnoreReqRespInLogList
}

func GetPricePrecision(regionOrCurrency string) int32 {
	regionOrCurrency = strings.ToUpper(regionOrCurrency)

	if GetCommonConfig() == nil {
		return defaultPricePrecision
	}

	if _, exist := GetCommonConfig().PricePrecisionMap[regionOrCurrency]; !exist {
		return defaultPricePrecision
	}

	return GetCommonConfig().PricePrecisionMap[regionOrCurrency]
}

func IsRegionDeprecated(region string) bool {
	c := GetCommonConfig()
	if c == nil || len(c.DeprecatedRegions) == 0 {
		return false
	}

	return c.DeprecatedRegions[region]
}

func GetShopMarginLimit(isCB bool) (max int32, min int32) {
	var marginLimitCfg *ShopMarginLimit
	if isCB {
		marginLimitCfg = GetCommonConfig().CBShopMarginLimit
	} else {
		marginLimitCfg = GetCommonConfig().LocalShopMarginLimit
	}
	return marginLimitCfg.MaxAShopMargin, marginLimitCfg.MinAShopMargin
}
