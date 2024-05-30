package cbsc_logic

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

func (c *CbscLogicImpl) CalculatePriceForCbsc(ctx context.Context, merchantId uint64, isMtskuToMpsku bool, queries []model.MtskuMpskuPriceQuery) ([]model.MtskuMpskuPriceCalcResult, error) {
	if len(queries) == 0 {
		return nil, nil
	}

	// get merchant region
	merchantRegion, err := c.shopMerchantService.GetMerchantRegion(ctx, merchantId)
	if err != nil {
		return nil, err
	}

	// get merchant config
	merchantConfigMap, err := c.merchantConfigService.GetSingleMerchantConfigSettingInfoMap(ctx, merchantId)
	if err != nil {
		return nil, err
	}

	// get factors.
	// get exchange rates
	merchantCurrency, exchangeRateMap, err := c.factorsRepo.GetExchangeRateMapForCbsc(ctx, merchantId)
	if err != nil {
		return nil, err
	}

	// get hidden fees
	hidePriceQueries := make([]model.GetHidePriceForCbscRequest, 0)
	for idx, query := range queries {
		hidePriceQueries = append(hidePriceQueries, model.GetHidePriceForCbscRequest{
			QueryId:              idx,
			Region:               query.MpskuRegion,
			Weight:               query.Weight,
			IsMtskuToMpsku:       isMtskuToMpsku,
			ShopId:               query.MpskuShopId,
			ItemId:               query.MpskuItemId,
			LeafCategoryId:       query.LeafCategoryId,
			EnabledChannelIdList: query.EnabledChannelIds,
			IgnoreChannelErr:     !isMtskuToMpsku,
		})
	}
	hidePriceList, err := c.factorsRepo.GetHidePriceForCbsc(ctx, hidePriceQueries)
	if err != nil {
		return nil, err
	}

	// get commission rates
	commissionRateQueries := make([]model.GetCommissionRateRequest, 0)
	for _, query := range queries {
		commissionRateQueries = append(commissionRateQueries, model.GetCommissionRateRequest{
			ShopId:      query.MpskuShopId,
			MpskuRegion: query.MpskuRegion,
		})
	}
	commissionRates := c.factorsRepo.GetCommissionRateBatchForCbsc(ctx, commissionRateQueries)
	if len(commissionRates) != len(commissionRateQueries) {
		return nil, cerr.New(fmt.Sprintf("the length of returned commissionRates is unexpected, length=%d, expected=%d",
			len(commissionRates), len(commissionRateQueries)), uint32(pb.Constant_ERROR_INTERNAL))
	}

	// get cbsc price rate
	cbscPriceRateQueries := make([]model.GetCbscPriceRateRequest, 0)
	for i, query := range queries {
		cbscPriceRateQueries = append(cbscPriceRateQueries, model.GetCbscPriceRateRequest{
			ShopId:            query.MpskuShopId,
			CommissionRate:    commissionRates[i].CommissionRate,
			CommissionRateErr: commissionRates[i].Err,
		})
	}
	cbscPriceRates, err := c.factorsRepo.GetCbscPriceRateBatchForCbsc(ctx, merchantRegion, merchantConfigMap, cbscPriceRateQueries)
	if err != nil {
		return nil, err
	}

	// get profit rates
	profitRateQueries := make([]model.GetCbscProfitRateRequest, 0)
	for _, query := range queries {
		profitRateQueries = append(profitRateQueries, model.GetCbscProfitRateRequest{
			ShopId: query.MpskuShopId,
		})
	}
	profitRates, err := c.factorsRepo.GetProfitRateBatchForCbsc(ctx, merchantConfigMap, profitRateQueries)
	if err != nil {
		return nil, err
	}

	finalResult := make([]model.MtskuMpskuPriceCalcResult, len(queries))
	for i, query := range queries {
		calcPrice := calcutil.RoundIntToFloat(query.SourcePrice, constant.PricePrecision, 2)
		exchangeRate, ok := exchangeRateMap[query.MpskuRegion]
		if !ok {
			finalResult[i] = model.MtskuMpskuPriceCalcResult{
				Err: cerr.New(fmt.Sprintf("failed to get exchange rate for region=%v", query.MpskuRegion), uint32(pb.Constant_ERROR_GET_MERCHANT_EXCHANGE_RATE)),
			}
			continue
		}
		profitRateRes := profitRates[i]
		if profitRateRes.Err != nil {
			finalResult[i] = model.MtskuMpskuPriceCalcResult{
				Err: profitRateRes.Err,
			}
			continue
		}
		profitRate := profitRateRes.ProfitRate

		hidePriceRes := hidePriceList[i]
		if hidePriceRes.Err != nil {
			finalResult[i] = model.MtskuMpskuPriceCalcResult{
				Err:                hidePriceRes.Err,
				HidePriceErrorCode: int32(cerr.Code(hidePriceRes.Err)),
			}
			continue
		}
		hidePrice := hidePriceRes.HidePrice

		cbscPriceRateRes := cbscPriceRates[i]
		if cbscPriceRateRes.Err != nil {
			finalResult[i] = model.MtskuMpskuPriceCalcResult{
				Err: cbscPriceRateRes.Err,
			}
			continue
		}
		cbscPriceRate := cbscPriceRateRes.CbscPriceRate

		var dstPrice float64
		if isMtskuToMpsku {
			dstPrice = calcutil.CalculateMpskuPrice(calcPrice, exchangeRate, profitRate, hidePrice, cbscPriceRate)
		} else {
			dstPrice = calcutil.CalculateMtskuPrice(calcPrice, exchangeRate, profitRate, hidePrice, cbscPriceRate)
		}

		var pricePrecision int32
		if isMtskuToMpsku {
			pricePrecision = config.GetPricePrecision(query.MpskuRegion)
		} else {
			pricePrecision = config.GetPricePrecision(merchantCurrency)
		}

		inflatedHidePrice := calcutil.RoundFloatToInt(hidePrice, constant.PricePrecision, pricePrecision)
		inflatedDstPrice := calcutil.RoundFloatToInt(dstPrice, constant.PricePrecision, pricePrecision)
		if !isMtskuToMpsku && inflatedDstPrice == 0 {
			inflatedDstPrice = calcutil.GetMiniPriceValueBasedOnCurrency(pricePrecision)
		}

		finalResult[i] = model.MtskuMpskuPriceCalcResult{
			DstPrice:  inflatedDstPrice,
			HidePrice: inflatedHidePrice,
		}

		logging.GetLogger(ctx).Info(fmt.Sprintf("[CBSC] Calc price for CBSC | isMtskuToMpsku:%v | "+
			"query=%+v, sourcePrice=%v, exchangeRate=%v, profitRate=%v, hidePrice=%v, cbscPriceRate=%v, "+
			"mpskuRegion=%s, merchantCurrency=%s, pricePrecision=%d | "+
			"actualPriceResult=%v, inflatedPriceRes=%v, inflatedHidePrice=%v",
			isMtskuToMpsku, query, calcPrice, exchangeRate, profitRate, hidePrice, cbscPriceRate,
			query.MpskuRegion, merchantCurrency, pricePrecision,
			dstPrice, inflatedDstPrice, inflatedHidePrice))
	}

	return finalResult, nil
}
