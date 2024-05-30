package constant

import (
	"fmt"
	"strconv"
	"strings"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/serverutil"
)

const (
	CachePreFix = "PriceCalculationService"
	CacheSplit  = ":"
)

const (
	CountryMarginConfigCacheKey    = "new_country_margin_config"
	shopMapByMstShoIdKey           = "shop_map_by_mst_shopid"
	mstShopCacheKey                = "mstshop"
	regionLevelChannelInfoCacheKey = "region_channel_info"
	orderMartExchangeRateCacheKey  = "order_mart_exchange_rate"
)

func WrapRedisKey(funcName string, selfKey string) string {
	var sb strings.Builder
	const baseLen = len(CachePreFix) + len(CacheSplit)*3 + 4
	sb.Grow(baseLen + len(funcName) + len(selfKey))
	sb.WriteString(CachePreFix)
	sb.WriteString(CacheSplit)
	sb.WriteString(serverutil.GetEnv())
	sb.WriteString(CacheSplit)
	sb.WriteString(funcName)
	sb.WriteString(CacheSplit)
	sb.WriteString(selfKey)
	return sb.String()
}

func GetReferenceServiceFeeRateCacheKey(shopId uint64) string {
	cacheKey := WrapRedisKey("GetReferenceServiceFeeRate", strconv.FormatUint(shopId, 10))
	return cacheKey
}

func GetProfitRateLimitCacheKey(merchantRegion string) string {
	cacheKey := "GetProfitRateLimitList-v2" + merchantRegion + serverutil.GetEnv()
	return cacheKey
}

func GetMerchantShopListCacheKey(merchantId uint64) string {
	cacheKey := WrapRedisKey("GetMerchantShopList-v2", strconv.FormatUint(merchantId, 10))
	return cacheKey
}

func GetAllCnscShopsCacheKey(merchantId uint64) string {
	return WrapRedisKey("GetAllCNSCShops-v2", strconv.FormatUint(merchantId, 10))
}

func GetExchangeRateCacheKey(currencyPair string) string {
	return fmt.Sprintf("exchange_rate#%s", currencyPair)
}

func GetShopMapCacheKey(pShopId uint64) string {
	key := fmt.Sprintf("%s:%d", shopMapByMstShoIdKey, pShopId)
	return key
}

func GetMstShopCacheKey(pShopId uint64) string {
	return fmt.Sprintf("%s:%v", mstShopCacheKey, pShopId)
}

func GetRegionChannelInfoMapCacheKey(region string) string {
	return fmt.Sprintf("%s:%v", regionLevelChannelInfoCacheKey, region)
}

func GetOrderMartExchangeRateCacheKey(currency string) string {
	return fmt.Sprintf("%s:%s", orderMartExchangeRateCacheKey, strings.ToUpper(currency))
}
