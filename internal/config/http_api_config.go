package config

import (
	"context"
	"strings"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/common/uniconfig"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type HTTPApiConfig struct { // TOFuture use own token and client name
	SlsFulfillmentUrl    string `json:"sls_fulfillment_url"`
	SlsFulfillmentToken  string `json:"sls_fulfillment_token"`
	SlsFulfillmentClient string `json:"sls_fulfillment_client"`
	SlsRetryTimeoutMs    int32  `json:"sls_retry_timeout_ms"`

	SlsOldUrlFormat string `json:"sls_old_url_format"`
	SlsOldToken     string `json:"sls_old_token"`

	SlsLpsUrlFormat string `json:"sls_lps_url_format"`
	SlsLpsToken     string `json:"sls_lps_token"`
	SlsLpsTimeoutMs int32  `json:"sls_lps_timeout_ms"`

	ShopeeRegionalDomainUrlMap map[string]string `json:"shopee_regional_domain_url_map"` // region -> domain url
	ShopLogisticsTimeoutMs     int32             `json:"shop_logistics_timeout_ms"`

	SellerAdminUrl       string `json:"seller_admin_url"`
	SellerAdminKey       string `json:"seller_admin_key"`
	SellerAdminAppId     string `json:"seller_admin_app_id"`
	SellerAdminTimeoutMs int32  `json:"seller_admin_timeout_ms"`

	ChargeCoreHost      string `json:"charge_core_host"`
	ChargeCoreToken     string `json:"charge_core_token"`
	ChargeCoreTimeoutMs int32  `json:"charge_core_timeout_ms"`

	PlatformMiscellaneousinnerHspexRouterPath string `json:"platform_miscellaneousinner_hspex_router_path"`

	SubAccountServer *SubAccountServer `json:"sub_account_server"`

	SellerManagerUrl   string `json:"seller_manager_url"`
	SellerManagerToken string `json:"seller_manager_token"`
}

type SubAccountServer struct {
	Host          string `json:"host,omitempty"`
	Token         string `json:"token,omitempty"`
	Salt          string `json:"salt,omitempty"`
	RemoteService string `json:"remote_service,omitempty"`
	Timeout       int    `json:"timeout,omitempty"`
}

func onHTTPApiConfigUpdate(e uniconfig.Event) {
	rawNewConfig, err := e.New()
	if err != nil {
		logging.GetLogger(context.Background()).Warn("error getting updated HTTPApiConfig value from uniconfig.Event", ulog.Error(err))
		return
	}

	newConfig, ok := rawNewConfig.(*HTTPApiConfig)
	if !ok {
		logging.GetLogger(context.Background()).Warn("new config is not a HTTPApiConfig",
			ulog.String("newVal", cutil.JSONEncode(rawNewConfig)))
		return
	}

	confVal.HTTPApiConfig = newConfig

	logging.GetLogger(context.Background()).Info("HTTPApiConfig is updated", ulog.String("http_api_config", cutil.JSONEncode(newConfig)))
}

func GetHTTPApiConfig() *HTTPApiConfig {
	if confVal == nil {
		return nil
	}
	return confVal.HTTPApiConfig
}

func GetShopeeRegionDomainUrl(region string) string {
	region = strings.ToUpper(region)

	if GetHTTPApiConfig() == nil || GetHTTPApiConfig().ShopeeRegionalDomainUrlMap[region] == "" {
		return ""
	}
	return GetHTTPApiConfig().ShopeeRegionalDomainUrlMap[region]
}
