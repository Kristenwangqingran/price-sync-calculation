package data

import (
	"fmt"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	internalMerchantConfigSettingPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
)

type CalcFactorDataForMtskuAndMpsku struct {
	// merchant level
	merchantRegionMap             map[uint64]*service.MerchantRegionInfo        // merchantId -> merchantRegion
	merchantCbscPriceFeeConfigMap map[string]*config.CBSCPriceFeeConfig         // merchantRegion -> CBSCPriceFeeConfig
	merchantConfigSettingMap      map[uint64]*service.MerchantConfigSettingInfo // merchantId -> (shopId -> merchantConfigSetting)
	merchantExchangeRateMap       map[uint64]*service.MerchantExchangeRateInfo  // merchantId -> (merchant currency && region -> exchangeRate(float))

	// mpsku shop level
	mpskuShopCommissionRateMap map[uint64]*service.ShopCommissionRateInfo // shopId -> commissionRate(inflated)

	// mpsku item level
	mpskuItemWeightMap                map[uint64]*service.ItemWeightInfo        // itemId -> item weight
	mpskuItemLeafCatMap               map[uint64]*service.ItemLeafCategoryInfo  // itemId -> item leaf category id
	mpskuItemShopEnabledChannelIdsMap map[uint64]*service.EnabledChannelIdsInfo // itemId -> enabled channel ids on shop level and item level both
}

func (data *CalcFactorDataForMtskuAndMpsku) GetMerchantRegion(merchantId uint64) (string, error) {
	merchantRegionMapInfo := data.merchantRegionMap[merchantId]

	if merchantRegionMapInfo == nil {
		return "", cerr.New(fmt.Sprintf(
			"failed to get merchant region, merchantId=%d", merchantId),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_MERCHANT_REGION))
	}

	if merchantRegionMapInfo.Err != nil {
		return "", merchantRegionMapInfo.Err
	}

	return merchantRegionMapInfo.MerchantRegion, nil
}

func (data *CalcFactorDataForMtskuAndMpsku) GetMerchantCbscPriceFeeConfig(merchantRegion string) *config.CBSCPriceFeeConfig {
	return config.GetCbscPriceFeeConfig(merchantRegion)
}

func (data *CalcFactorDataForMtskuAndMpsku) GetMerchantConfigSetting(merchantId uint64, mpskuShopId uint64) (*internalMerchantConfigSettingPb.MerchantConfigSetting, error) {
	merchantConfigSettingMapInfo := data.merchantConfigSettingMap[merchantId]

	if merchantConfigSettingMapInfo == nil {
		return nil, cerr.New(fmt.Sprintf(
			"failed to get merchant config setting, merchantId=%d", merchantId),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_MERCHANT_CONFIG_SETTING))
	}

	if merchantConfigSettingMapInfo.Err != nil {
		return nil, merchantConfigSettingMapInfo.Err
	}

	if merchantConfigSettingMapInfo.MerchantConfigSettingMap[mpskuShopId] == nil {
		return nil, cerr.New(fmt.Sprintf(
			"failed to get merchant config setting, merchantId=%d, mpskuShopId=%d",
			merchantId, mpskuShopId),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_MERCHANT_CONFIG_SETTING))
	}

	return merchantConfigSettingMapInfo.MerchantConfigSettingMap[mpskuShopId], nil
}

func (data *CalcFactorDataForMtskuAndMpsku) GetMerchantExchangeRate(merchantId uint64, mpskuRegion string) (string, float64, error) {
	merchantExchangeRateMapInfo := data.merchantExchangeRateMap[merchantId]

	if merchantExchangeRateMapInfo == nil {
		return "", 0, cerr.New(fmt.Sprintf(
			"failed to get exchange rate, merchantId=%d", merchantId),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_MERCHANT_EXCHANGE_RATE))
	}

	if merchantExchangeRateMapInfo.Err != nil {
		return "", 0, merchantExchangeRateMapInfo.Err
	}

	if merchantExchangeRateMapInfo.MerchantExchangeRateMap[mpskuRegion] == 0 {
		return "", 0, cerr.New(fmt.Sprintf(
			"failed to get exchange rate, mpskuRegion=%s", mpskuRegion),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_MERCHANT_EXCHANGE_RATE))
	}

	return merchantExchangeRateMapInfo.MerchantCurrency, merchantExchangeRateMapInfo.MerchantExchangeRateMap[mpskuRegion], nil
}

func (data *CalcFactorDataForMtskuAndMpsku) GetMpskuShopCommissionRate(mpskuShopId uint64) (uint64, error) {
	mpskuShopCommissionRateMapInfo := data.mpskuShopCommissionRateMap[mpskuShopId]

	if mpskuShopCommissionRateMapInfo == nil {
		return 0, cerr.New(fmt.Sprintf(
			"failed to get commission rate, mpskuShopId=%d", mpskuShopId),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_SHOP_COMMISSION_RATE))
	}

	if mpskuShopCommissionRateMapInfo.Err != nil {
		return 0, mpskuShopCommissionRateMapInfo.Err
	}

	return mpskuShopCommissionRateMapInfo.CommissionRate, nil
}

func (data *CalcFactorDataForMtskuAndMpsku) GetMpskuItemWeight(mpskuItemId uint64) (uint64, error) {
	mpskuWeightInfo := data.mpskuItemWeightMap[mpskuItemId]

	if mpskuWeightInfo == nil {
		return 0, cerr.New(fmt.Sprintf(
			"failed to get mpsku item weight info, itemId=%d", mpskuItemId),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_ITEM_INFO))
	}

	if mpskuWeightInfo.Err != nil {
		return 0, mpskuWeightInfo.Err
	}

	return mpskuWeightInfo.ItemWeight, nil
}

func (data *CalcFactorDataForMtskuAndMpsku) GetMpskuItemLeafCategoryId(mpskuItemId uint64) (uint32, error) {
	mpskuCatInfo := data.mpskuItemLeafCatMap[mpskuItemId]

	if mpskuCatInfo == nil {
		return 0, cerr.New(fmt.Sprintf(
			"failed to get mpsku item category info, itemId=%d", mpskuItemId),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_ITEM_INFO))
	}

	if mpskuCatInfo.Err != nil {
		return 0, mpskuCatInfo.Err
	}

	return mpskuCatInfo.LeafCatId, nil
}

func (data *CalcFactorDataForMtskuAndMpsku) GetMpskuItemShopEnabledChannelIds(mpskuItemId uint64) ([]uint32, error) {
	mpskuEnabledChannelListInfo := data.mpskuItemShopEnabledChannelIdsMap[mpskuItemId]

	if mpskuEnabledChannelListInfo == nil {
		return nil, cerr.New(fmt.Sprintf(
			"failed to get enabled channel list, itemId=%d", mpskuItemId),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_ENABLED_CHANNELS))
	}

	if mpskuEnabledChannelListInfo.Err != nil {
		return nil, mpskuEnabledChannelListInfo.Err
	}

	if len(mpskuEnabledChannelListInfo.EnabledChannelIds) == 0 {
		return nil, cerr.New(fmt.Sprintf(
			"no available channel to calculate hidden fee, itemId=%d", mpskuItemId),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_EMPTY_ENABLED_CHANNELS))
	}

	return mpskuEnabledChannelListInfo.EnabledChannelIds, nil
}
