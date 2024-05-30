package processor

import (
	"context"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

func (s *CalculationServiceImpl) GetCbscPriceFactor(ctx context.Context, request *priceSyncPriceCalculationPb.GetCbscPriceFactorRequest, response *priceSyncPriceCalculationPb.GetCbscPriceFactorResponse) uint32 {
	p := &getCbscPriceFactorProcessor{
		ctx:       ctx,
		request:   request,
		response:  response,
		cbscLogic: s.cbscLogic,
	}

	err := p.process()
	if err != nil {
		response.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type getCbscPriceFactorProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetCbscPriceFactorRequest
	response *priceSyncPriceCalculationPb.GetCbscPriceFactorResponse

	cbscLogic logic.CbscLogic
}

func (g *getCbscPriceFactorProcessor) process() error {
	if err := g.validateRequest(); err != nil {
		return err
	}

	factors, err := g.cbscLogic.GetCbscPriceFactor(g.ctx, g.request)
	if err != nil {
		return err
	}

	g.response.Results = factors
	return nil
}

func (g *getCbscPriceFactorProcessor) validateRequest() error {
	req := g.request

	if _, ok := priceSyncPriceCalculationPb.Constant_CbscPriceFactorInfoType_name[int32(req.GetInfoType())]; !ok {
		return cerr.New("invalid info type", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if req.MerchantId == nil {
		return cerr.New("invalid MerchantId", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	return nil
}
