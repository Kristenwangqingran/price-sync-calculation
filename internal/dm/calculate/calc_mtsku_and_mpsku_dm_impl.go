package calculate

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/data"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/factors"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

const (
	ErrMsgGlobalDiscountRateTooLarge = "the discount rate in the input is larger than 100%"
	ErrMsgGlobalDiscountRateTooSmall = "the discount rate in the input is less than 0%"
)

type CalculateMtskuAndMpskuDmImpl struct {
	logisticService service.LogisticService
	factorsRepo     factors.CalculationFactorsRepo
}

func NewMtskuAndMpskuCalculateDm(logisticService service.LogisticService, factorsRepo factors.CalculationFactorsRepo) CalculateMtskuAndMpskuDm {
	return &CalculateMtskuAndMpskuDmImpl{
		logisticService: logisticService,
		factorsRepo:     factorsRepo,
	}
}

func (dm *CalculateMtskuAndMpskuDmImpl) CalcGlobalDiscount(ctx context.Context,
	queries []*priceSyncPriceCalculationPb.GlobalDiscountQueryId, calculateData *data.CalcFactorDataForMtskuAndMpsku) (
	[]*priceSyncPriceCalculationPb.GlobalDiscountInfo, error) {

	if calculateData == nil {
		return nil, cerr.New("calculateData is empty", uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	globalDiscountInfoList := make([]*priceSyncPriceCalculationPb.GlobalDiscountInfo, len(queries))

	for idx, query := range queries {
		queryDataForMtskuAndMpsku := &QueryDataForMtskuAndMpsku{
			MerchantId:         query.GetMerchantId(),
			MpskuShopId:        query.GetMpskuShopId(),
			MpskuItemId:        query.GetMpskuItemId(),
			MpskuModelId:       query.GetMpskuModelId(),
			MpskuRegion:        query.GetMpskuRegion(),
			MtskuOriginalPrice: query.GetMtskuOriginalPrice(),
		}

		globalDiscountInfoList[idx] = &priceSyncPriceCalculationPb.GlobalDiscountInfo{
			MerchantId:              query.MerchantId,
			MpskuShopId:             query.MpskuShopId,
			MpskuItemId:             query.MpskuItemId,
			MpskuModelId:            query.MpskuModelId,
			MpskuRegion:             query.MpskuRegion,
			MtskuOriginalPrice:      query.MtskuOriginalPrice,
			GlobalDiscountInputType: query.GlobalDiscountInputType,
			GlobalDiscountQueryData: query.GlobalDiscountQueryData,
		}

		// calculate based on input
		switch query.GetGlobalDiscountInputType() {
		case uint32(priceSyncPriceCalculationPb.Constant_DISCOUNT_RATE): // mtsku original price + discount rate -> mpsku promo price
			// checked with discount side, no need return calculate result if the discount rate in the request is unexpected
			if query.GetGlobalDiscountQueryData() <= 0 {
				globalDiscountInfoList[idx].ErrMsg = proto.String(ErrMsgGlobalDiscountRateTooSmall)
				globalDiscountInfoList[idx].ErrCode = proto.Uint32(uint32(priceSyncPriceCalculationPb.Constant_ERROR_GLOBAL_DISCOUNT_UNEXPECTED))
				continue
			}
			if query.GetGlobalDiscountQueryData() >= constant.PercentPrecisionBetweenMtskuAndMpsku {
				globalDiscountInfoList[idx].ErrMsg = proto.String(ErrMsgGlobalDiscountRateTooLarge)
				globalDiscountInfoList[idx].ErrCode = proto.Uint32(uint32(priceSyncPriceCalculationPb.Constant_ERROR_GLOBAL_DISCOUNT_UNEXPECTED))
				continue
			}

			inflatedMtskuPromoPrice := int64(float64(query.GetMtskuOriginalPrice()) * (float64(constant.PercentPrecisionBetweenMtskuAndMpsku-query.GetGlobalDiscountQueryData()) / float64(constant.PercentPrecisionBetweenMtskuAndMpsku)))
			queryDataForMtskuAndMpsku.QueryPriceData = inflatedMtskuPromoPrice
			queryDataForMtskuAndMpsku.QueryMtskuToMpsku = true

			inflatedMpskuPrice, err := dm.calcMtskuAndMpsku(ctx, queryDataForMtskuAndMpsku, calculateData)
			if err != nil {
				globalDiscountInfoList[idx].ErrMsg = proto.String(err.Error())
				globalDiscountInfoList[idx].ErrCode = proto.Uint32(cerr.Code(err))
				continue
			}

			globalDiscountInfoList[idx].GlobalDiscountQueryResult = proto.Int64(inflatedMpskuPrice)
		case uint32(priceSyncPriceCalculationPb.Constant_MPSKU_PRICE): // mpsku promo price + mtsku original price -> discount rate
			queryDataForMtskuAndMpsku.QueryPriceData = query.GetGlobalDiscountQueryData()
			queryDataForMtskuAndMpsku.QueryMtskuToMpsku = false

			inflatedMtskuPromoPrice, err := dm.calcMtskuAndMpsku(ctx, queryDataForMtskuAndMpsku, calculateData)
			if err != nil {
				globalDiscountInfoList[idx].ErrMsg = proto.String(err.Error())
				globalDiscountInfoList[idx].ErrCode = proto.Uint32(cerr.Code(err))
				continue
			}

			discountRate := float64(query.GetMtskuOriginalPrice()-inflatedMtskuPromoPrice) / float64(query.GetMtskuOriginalPrice())
			inflatedDiscountRate := calcutil.RoundFloatToInt(discountRate, constant.PercentPrecisionBetweenMtskuAndMpsku, 2)

			// checked with discount side, still need to return calculate result if the calculated discount rate is unexpected
			if inflatedDiscountRate <= 0 {
				globalDiscountInfoList[idx].ErrMsg = proto.String(ErrMsgGlobalDiscountRateTooSmall)
				globalDiscountInfoList[idx].ErrCode = proto.Uint32(uint32(priceSyncPriceCalculationPb.Constant_ERROR_GLOBAL_DISCOUNT_UNEXPECTED))
			}
			if inflatedDiscountRate >= constant.PercentPrecisionBetweenMtskuAndMpsku {
				globalDiscountInfoList[idx].ErrMsg = proto.String(ErrMsgGlobalDiscountRateTooLarge)
				globalDiscountInfoList[idx].ErrCode = proto.Uint32(uint32(priceSyncPriceCalculationPb.Constant_ERROR_GLOBAL_DISCOUNT_UNEXPECTED))
			}

			globalDiscountInfoList[idx].GlobalDiscountQueryResult = proto.Int64(inflatedDiscountRate)
		}
	}

	return globalDiscountInfoList, nil
}

func (dm *CalculateMtskuAndMpskuDmImpl) calcMtskuAndMpsku(ctx context.Context, query *QueryDataForMtskuAndMpsku, calcFactorData *data.CalcFactorDataForMtskuAndMpsku) (int64, error) {
	if query == nil || calcFactorData == nil {
		return 0, cerr.New("query or calculateData is empty", uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	// get merchant region
	merchantRegion, err := calcFactorData.GetMerchantRegion(query.MerchantId)
	if err != nil {
		return 0, err
	}

	// get item related info
	mpskuLeafCatId, err := calcFactorData.GetMpskuItemLeafCategoryId(query.MpskuItemId)
	if err != nil {
		return 0, err
	}
	mpskuItemWeight, err := calcFactorData.GetMpskuItemWeight(query.MpskuItemId)
	if err != nil {
		return 0, err
	}

	// get exchange rate
	merchantCurrency, exchangeRate, err := calcFactorData.GetMerchantExchangeRate(query.MerchantId, query.MpskuRegion)
	if err != nil {
		return 0, err
	}

	// get profit rate
	merchantConfigSetting, err := calcFactorData.GetMerchantConfigSetting(query.MerchantId, query.MpskuShopId)
	if err != nil {
		return 0, err
	}
	if merchantConfigSetting == nil || merchantConfigSetting.ProfitRate == nil {
		return 0, cerr.New(fmt.Sprintf(
			"failed to get profit rate in merchant config setting, MerchantId=%d, MpskuShopId=%d",
			query.MerchantId, query.MpskuShopId),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_MERCHANT_CONFIG_SETTING))
	}
	inflatedProfitRate := merchantConfigSetting.GetProfitRate()
	profitRate := calcutil.RoundIntToFloat(int64(inflatedProfitRate), constant.PercentPrecisionBetweenMtskuAndMpsku, 4)

	// calculate hidden fee
	mpskuEnabledChannelIdList, err := calcFactorData.GetMpskuItemShopEnabledChannelIds(query.MpskuItemId)
	if err != nil {
		return 0, err
	}
	hidePriceRests, err := dm.factorsRepo.GetHidePriceForCbsc(ctx,
		[]model.GetHidePriceForCbscRequest{
			{
				QueryId:              0,
				Region:               query.MpskuRegion,
				Weight:               mpskuItemWeight,
				IsMtskuToMpsku:       query.QueryMtskuToMpsku,
				ShopId:               query.MpskuShopId,
				ItemId:               query.MpskuItemId,
				LeafCategoryId:       uint64(mpskuLeafCatId),
				EnabledChannelIdList: mpskuEnabledChannelIdList,
				IgnoreChannelErr:     false,
			},
		},
	)
	if err != nil {
		return 0, cerr.New(fmt.Sprintf(
			"failed to calculate hidden fee, itemId=%d, err=%s", query.MpskuItemId, err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CALCULATE_HIDDEN_FEE))
	}
	if len(hidePriceRests) != 1 {
		return 0, cerr.New(fmt.Sprintf(
			"failed to calculate hidden fee, itemId=%d, hidden price result is not as expected (1)", query.MpskuItemId),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}
	hidePriceRest := hidePriceRests[0]
	if hidePriceRest.Err != nil {
		return 0, cerr.New(fmt.Sprintf(
			"failed to calculate hidden fee, itemId=%d, err=%s", query.MpskuItemId, hidePriceRest.Err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_CALCULATE_HIDDEN_FEE))
	}
	hidePrice := hidePriceRest.HidePrice

	// calculate cbscDenominatorPriceRate
	cbscPriceFeeConfig := calcFactorData.GetMerchantCbscPriceFeeConfig(merchantRegion)
	inflatedCommissionRate, err := calcFactorData.GetMpskuShopCommissionRate(query.MpskuShopId)
	if err != nil {
		return 0, err
	}
	cbscDenominatorPriceRate := calcutil.GetCBSCDenominatorPriceRate(ctx, cbscPriceFeeConfig, merchantConfigSetting, inflatedCommissionRate)

	// get result price precision
	var pricePrecision int32
	if query.QueryMtskuToMpsku {
		pricePrecision = config.GetPricePrecision(query.MpskuRegion)
	} else {
		pricePrecision = config.GetPricePrecision(merchantCurrency)
	}

	// calculate mtsku <-> mpsku
	var actualResPrice float64
	var inflatedResPrice int64
	calcPrice := calcutil.RoundIntToFloat(query.QueryPriceData, constant.PricePrecision, 2)
	if query.QueryMtskuToMpsku {
		actualResPrice = calcutil.CalculateMpskuPrice(calcPrice, exchangeRate, profitRate, hidePrice, cbscDenominatorPriceRate)
		inflatedResPrice = calcutil.RoundFloatToInt(actualResPrice, constant.PricePrecision, pricePrecision)
	} else {
		actualResPrice = calcutil.CalculateMtskuPrice(calcPrice, exchangeRate, profitRate, hidePrice, cbscDenominatorPriceRate)
		inflatedResPrice = calcutil.RoundFloatToInt(actualResPrice, constant.PricePrecision, pricePrecision)
	}

	logging.GetLogger(ctx).Info(fmt.Sprintf(
		"calculation factor for query=%s. "+
			"actualCalcPrice=%v, actualExchangeRate=%v, actualProfitRate(Market Adjustment Rate)=%v, actualHiddenPrice=%v, actualCbscDenominatorPriceRate=%v, "+
			"actualResPrice=%v, inflatedResPrice=%v, "+
			"mpskuRegion=%s, merchantCurrency=%s, pricePrecision=%d",
		cutil.JSONEncode(query), calcPrice, exchangeRate, profitRate, hidePrice, cbscDenominatorPriceRate,
		actualResPrice, inflatedResPrice,
		query.MpskuRegion, merchantCurrency, pricePrecision))

	return inflatedResPrice, nil
}
