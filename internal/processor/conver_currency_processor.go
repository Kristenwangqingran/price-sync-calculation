package processor

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

func (s *CalculationServiceImpl) ConvertCurrency(ctx context.Context, request *priceSyncPriceCalculationPb.ConvertCurrencyRequest, response *priceSyncPriceCalculationPb.ConvertCurrencyResponse) uint32 {
	p := &convertCurrencyProcessor{
		ctx:           ctx,
		request:       request,
		response:      response,
		currencyLogic: s.currencyConvertLogic,
	}

	err := p.process()
	if err != nil {
		response.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type convertCurrencyProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.ConvertCurrencyRequest
	response *priceSyncPriceCalculationPb.ConvertCurrencyResponse

	currencyLogic logic.CurrencyConvertLogic
}

func (c *convertCurrencyProcessor) process() error {
	if err := c.validateRequest(); err != nil {
		return err
	}

	result, err := c.currencyLogic.ConvertCurrency(c.ctx, model.ConvertCurrencyRequest{
		SrcPriceList:       c.request.GetSrcPriceList(),
		ExchangeRateSource: c.request.GetExchangeRateSource(),
		SrcCurrency:        c.request.GetSrcCurrency(),
		DstCurrency:        c.request.GetDstCurrency(),
		MerchantId:         c.request.GetMerchantId(),
		MpskuRegion:        c.request.GetMpskuRegion(),
	})

	if err != nil {
		return err
	}
	c.response.DstPrices = result.DstPrices
	c.response.ExchangeRate = proto.Float64(result.ExchangeRate)
	return nil
}

func (c *convertCurrencyProcessor) validateRequest() error {
	req := c.request
	if len(req.GetSrcPriceList()) == 0 {
		return cerr.New("invalid SrcPriceList", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if req.ExchangeRateSource == nil {
		return cerr.New("invalid ExchangeRateSource", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if _, ok := priceSyncPriceCalculationPb.Constant_ExchangeRateSource_name[int32(req.GetExchangeRateSource())]; !ok {
		return cerr.New("invalid ExchangeRateSource", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if req.GetExchangeRateSource() == uint32(priceSyncPriceCalculationPb.Constant_SELLER_PLATFORM) {
		if len(req.GetMpskuRegion()) == 0 || req.MerchantId == nil || !cutil.IsValidCountry(req.GetMpskuRegion()) {
			return cerr.New(fmt.Sprintf("for exchangeRateSource=sellerPlatform, correct mpskuRegion and merchantId should be provided"), uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
	}

	return nil
}
