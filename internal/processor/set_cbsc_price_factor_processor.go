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

func (s *CalculationServiceImpl) SetCbscPriceFactor(ctx context.Context,
	req *priceSyncPriceCalculationPb.SetCbscPriceFactorRequest, resp *priceSyncPriceCalculationPb.SetCbscPriceFactorResponse) uint32 {
	p := &setCbscPriceFactorProcessor{
		ctx:       ctx,
		request:   req,
		response:  resp,
		cbscLogic: s.cbscLogic,
	}

	err := p.process()
	if err != nil {
		resp.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type setCbscPriceFactorProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.SetCbscPriceFactorRequest
	response *priceSyncPriceCalculationPb.SetCbscPriceFactorResponse

	cbscLogic logic.CbscLogic
}

func (g *setCbscPriceFactorProcessor) process() error {
	if err := g.validateRequest(); err != nil {
		return cerr.Wrap(err, "", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	shopSettings := make([]model.ShopCbscPriceFactorSetting, 0)
	for _, factor := range g.request.GetShopCbscPriceFactors() {
		shopSettings = append(shopSettings, model.ShopCbscPriceFactorSetting{
			ShopId:         uint64(factor.GetShopId()),
			Region:         factor.GetRegion(),
			ProfitRate:     factor.ProfitRate,
			ServiceFeeRate: factor.ServiceFeeRate,
		})
	}

	query := model.SetCbscPriceFactorQuery{
		MerchantId:   g.request.GetMerchantId(),
		ShopSettings: shopSettings,
	}

	err := g.cbscLogic.SetCbscPriceFactor(g.ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (g *setCbscPriceFactorProcessor) validateRequest() error {
	if g.request.GetMerchantId() == 0 || len(g.request.GetShopCbscPriceFactors()) == 0 {
		return cerr.New(fmt.Sprintf("merchant_id = 0 or empty factor setting"),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	serviceFeeLimit, profitRateLimitMap, err := g.cbscLogic.GetCbscPriceFactorLimit(g.ctx, g.request.GetMerchantId())
	if err != nil {
		return err
	}

	for _, factor := range g.request.GetShopCbscPriceFactors() {
		if factor.GetProfitRate() < 0 || factor.GetServiceFeeRate() < 0 {
			return cerr.New(fmt.Sprintf("profit_rate < 0 or service_fee_rate < 0"),
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if len(factor.GetRegion()) == 0 || factor.GetShopId() <= 0 {
			return cerr.New(fmt.Sprintf("region is empty or invalid shop_id"),
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		// check service fee limit if exists
		if serviceFeeLimit != nil && factor.ServiceFeeRate != nil &&
			(serviceFeeLimit.GetMinServiceFeeRate() > int64(factor.GetServiceFeeRate()) || serviceFeeLimit.GetMaxServiceFeeRate() < int64(factor.GetServiceFeeRate())) {
			return cerr.New(fmt.Sprintf("service_fee_rate not meet limitation|limit=%v", cutil.LazyJSONEncoder(serviceFeeLimit)),
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		// check profit rate limit if exists
		profitRateLimit := profitRateLimitMap[factor.GetRegion()]
		if profitRateLimit != nil && factor.ProfitRate != nil &&
			(profitRateLimit.GetMinProfitRate() > int64(factor.GetProfitRate()) || profitRateLimit.GetMaxProfitRate() < int64(factor.GetProfitRate())) {
			return cerr.New(fmt.Sprintf("profit_rate not meet limitation|limit=%v", cutil.LazyJSONEncoder(profitRateLimit)),
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
	}

	return nil
}
