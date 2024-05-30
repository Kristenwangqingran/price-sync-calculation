package currency_convert_logic

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/wire"
	"github.com/shopspring/decimal"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/factors"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
)

type CurrencyConvertLogicImpl struct {
	factorsRepo factors.CalculationFactorsRepo
}

func NewCurrencyConverterLogic(factorsRepo factors.CalculationFactorsRepo) *CurrencyConvertLogicImpl {
	return &CurrencyConvertLogicImpl{factorsRepo: factorsRepo}
}

func (c *CurrencyConvertLogicImpl) ConvertCurrency(ctx context.Context, req model.ConvertCurrencyRequest) (model.ConvertCurrencyResult, error) {
	switch req.ExchangeRateSource {
	case model.ExchangeRateSourceCbSipExchangeRate:
		return c.convertCurrencyForCbSip(ctx, req)
	case model.ExchangeRateSourceSellerPlatform:
		return c.convertCurrencyBySellerPlatform(ctx, req)
	case model.ExchangeRateSourceOrderMart:
		return c.convertCurrencyByOrderMart(ctx, req)
	default:
		return model.ConvertCurrencyResult{}, cerr.New(fmt.Sprintf("invalid exchange rate source: %v", req.ExchangeRateSource), uint32(pb.Constant_ERROR_PARAMS))
	}
}

func (c *CurrencyConvertLogicImpl) convertCurrencyBySellerPlatform(ctx context.Context, req model.ConvertCurrencyRequest) (model.ConvertCurrencyResult, error) {
	info, err := c.factorsRepo.GetMerchantExchangeRateInfo(ctx, req.MerchantId)
	if err != nil {
		return model.ConvertCurrencyResult{}, err
	}

	var exchangeRate float64
	foundExchangeRate := false
	for _, data := range info.GetExchangeRateList() {
		if req.MpskuRegion == data.GetRegion() {
			exchangeRate = data.GetExchangeRate()
			foundExchangeRate = true
			break
		}
	}
	if !foundExchangeRate {
		return model.ConvertCurrencyResult{}, cerr.New(fmt.Sprintf("cannot find exchnage rate for merchantId=%v and mpskuRegion=%v", req.MerchantId, req.MpskuRegion), uint32(pb.Constant_ERROR_NOT_FOUND))
	}

	res := c.convertByExchangeRate(req.SrcPriceList, exchangeRate, false, nil)
	return model.ConvertCurrencyResult{
		ExchangeRate: exchangeRate,
		DstPrices:    res,
	}, nil
}

func (c *CurrencyConvertLogicImpl) convertCurrencyForCbSip(ctx context.Context, req model.ConvertCurrencyRequest) (model.ConvertCurrencyResult, error) {
	exchangeRateMap, err := c.factorsRepo.GetAllExchangeRateMapForCbSip(ctx)
	if err != nil {
		return model.ConvertCurrencyResult{}, err
	}
	exchangeRateStr, ok := exchangeRateMap[req.SrcCurrency][req.DstCurrency]
	if !ok {
		return model.ConvertCurrencyResult{}, cerr.New(fmt.Sprintf("failed to get exchange rate for srcCurrency=%v and dstCurrency=%v", req.SrcCurrency, req.DstCurrency), uint32(pb.Constant_ERROR_NOT_FOUND))
	}
	exchangeRate, err := strconv.ParseFloat(exchangeRateStr, 64)
	if err != nil {
		return model.ConvertCurrencyResult{}, cerr.New(fmt.Sprintf("invalid exchange rate: %v, err=%v", exchangeRateStr, err), uint32(pb.Constant_ERROR_INTERNAL))
	}

	res := c.convertByExchangeRate(req.SrcPriceList, exchangeRate, false, nil)

	return model.ConvertCurrencyResult{
		ExchangeRate: exchangeRate,
		DstPrices:    res,
	}, nil
}

func (c *CurrencyConvertLogicImpl) convertCurrencyByOrderMart(ctx context.Context, req model.ConvertCurrencyRequest) (model.ConvertCurrencyResult, error) {
	exchangeRate, err := c.factorsRepo.GetOrderMartExchangeRate(ctx, req.SrcCurrency, req.DstCurrency)
	if err != nil {
		return model.ConvertCurrencyResult{}, err
	}

	pricePrecision := config.GetPricePrecision(req.DstCurrency)
	res := c.convertByExchangeRate(req.SrcPriceList, exchangeRate, true, &pricePrecision)
	return model.ConvertCurrencyResult{
		ExchangeRate: exchangeRate,
		DstPrices:    res,
	}, nil
}

func (c *CurrencyConvertLogicImpl) convertByExchangeRate(srcPriceList []int64, exchangeRate float64, needRegionPrecision bool, roundPlace *int32) []int64 {
	res := make([]int64, 0)
	decimalExchangeRate := decimal.NewFromFloat(exchangeRate)
	for _, srcPrice := range srcPriceList {
		decimalSrcPrice := decimal.NewFromFloat(calcutil.ToRealPrice(srcPrice))
		decimalDstValue := decimalSrcPrice.Mul(decimalExchangeRate)

		dstPrice, _ := decimalDstValue.Float64()

		var dstDbPrice int64
		if needRegionPrecision && roundPlace != nil {
			dstDbPrice = calcutil.RoundFloatToInt(dstPrice, int32(constant.PricePrecision), *roundPlace)
		} else {
			dstDbPrice = calcutil.ToDBPrice(dstPrice)
		}
		res = append(res, dstDbPrice)
	}
	return res
}

var ProviderSet = wire.NewSet(
	NewCurrencyConverterLogic,
)
