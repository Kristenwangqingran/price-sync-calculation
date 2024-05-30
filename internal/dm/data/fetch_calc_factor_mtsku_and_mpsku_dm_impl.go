package data

import (
	"context"

	"github.com/google/wire"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
)

var fetchCalcFactorMtskuAndMpskuProvicerSet = wire.NewSet(
	wire.Struct(new(FetchCalcFactorMtskuAndMpskuDmOpts), "*"),
	NewFetchCalcFactorMtskuAndMpskuDm,
)

type FetchCalcFactorMtskuAndMpskuDmOpts struct {
	LogisticService                  service.LogisticService
	ShopMerchantService              service.ShopMerchantService
	ItemService                      service.ItemService
	MerchantConfigService            service.MerchantConfigService
	ExchangeRateService              service.ExchangeRateService
	OrderAccountIntegratedFeeService service.OrderAccountIntegratedFeeService
}

type FetchCalcFactorMtskuAndMpskuDmImpl struct {
	logisticService                  service.LogisticService
	shopMerchantService              service.ShopMerchantService
	itemService                      service.ItemService
	merchantConfigService            service.MerchantConfigService
	exchangeRateService              service.ExchangeRateService
	orderAccountIntegratedFeeService service.OrderAccountIntegratedFeeService
}

func NewFetchCalcFactorMtskuAndMpskuDm(opts *FetchCalcFactorMtskuAndMpskuDmOpts) FetchCalcFactorForMtskuAndMpskuDm {
	return &FetchCalcFactorMtskuAndMpskuDmImpl{
		shopMerchantService:              opts.ShopMerchantService,
		itemService:                      opts.ItemService,
		merchantConfigService:            opts.MerchantConfigService,
		exchangeRateService:              opts.ExchangeRateService,
		orderAccountIntegratedFeeService: opts.OrderAccountIntegratedFeeService,
		logisticService:                  opts.LogisticService,
	}
}

func (dm *FetchCalcFactorMtskuAndMpskuDmImpl) FetchCalcFactorDataForGlobalDiscount(ctx context.Context, queries []*priceSyncPriceCalculationPb.GlobalDiscountQueryId) *CalcFactorDataForMtskuAndMpsku {
	merchantIdList := getUniqueMerchantIdListFromGlobalDiscountQueryIds(queries)
	mpskuShopIdRegionList := getUniqueMpskuShopIdListFromGlobalDiscountQueryIds(queries)
	mpskuItemIdList, mpskuRegionItemIdMap := getUniqueMpskuItemIdListFromGlobalDiscountQueryIds(queries)

	calcFactorDataForMtskuAndMpsku := &CalcFactorDataForMtskuAndMpsku{}
	// merchant level
	calcFactorDataForMtskuAndMpsku.merchantRegionMap = dm.shopMerchantService.GetMerchantRegionInfoMap(ctx, merchantIdList)
	calcFactorDataForMtskuAndMpsku.merchantConfigSettingMap = dm.merchantConfigService.GetMerchantConfigSettingInfoMap(ctx, merchantIdList)
	calcFactorDataForMtskuAndMpsku.merchantCbscPriceFeeConfigMap = config.GetCbscPriceFeeConfigMap()
	calcFactorDataForMtskuAndMpsku.merchantExchangeRateMap = dm.exchangeRateService.GetMerchantExchangeRateMap(ctx, merchantIdList)

	// mpsku shop level
	calcFactorDataForMtskuAndMpsku.mpskuShopCommissionRateMap = dm.orderAccountIntegratedFeeService.GetShopCommissionRateMap(ctx, mpskuShopIdRegionList)

	// mpsku item level
	mpskuItemInfoMap := dm.itemService.GetProductInfoMapForMixedRegion(ctx, mpskuRegionItemIdMap)
	calcFactorDataForMtskuAndMpsku.mpskuItemLeafCatMap = dm.itemService.GetItemLeafCatIdMap(ctx, mpskuItemIdList, mpskuItemInfoMap)
	calcFactorDataForMtskuAndMpsku.mpskuItemWeightMap = dm.itemService.GetItemWeightMap(ctx, mpskuItemIdList, mpskuItemInfoMap)

	// mpsku item and shop enabled channels
	fetchItemShopEnabledChannelIdsQueryList := make([]*FetchItemShopEnabledChannelIdsQuery, len(queries))
	for idx, q := range queries {
		fetchItemShopEnabledChannelIdsQueryList[idx] = &FetchItemShopEnabledChannelIdsQuery{
			ItemId: q.GetMpskuItemId(),
			ShopId: q.GetMpskuShopId(),
			Region: q.GetMpskuRegion(),
		}
	}
	calcFactorDataForMtskuAndMpsku.mpskuItemShopEnabledChannelIdsMap = FetchItemShopEnabledChannelIds(ctx,
		fetchItemShopEnabledChannelIdsQueryList, mpskuItemInfoMap, dm.itemService, dm.logisticService)

	return calcFactorDataForMtskuAndMpsku
}
