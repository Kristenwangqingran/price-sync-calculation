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

func (s *CalculationServiceImpl) GetCbSipRateConfig(ctx context.Context, request *priceSyncPriceCalculationPb.GetCbSipRateConfigRequest, response *priceSyncPriceCalculationPb.GetCbSipRateConfigResponse) uint32 {
	p := &getCbSipRateConfigProcessor{
		ctx:        ctx,
		request:    request,
		response:   response,
		cbsipLogic: s.cbsipLogic,
	}

	err := p.process()
	if err != nil {
		response.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type getCbSipRateConfigProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetCbSipRateConfigRequest
	response *priceSyncPriceCalculationPb.GetCbSipRateConfigResponse

	cbsipLogic logic.CbSipLogic
}

func (g *getCbSipRateConfigProcessor) process() error {
	if err := g.validateRequest(); err != nil {
		return err
	}

	config, err := g.cbsipLogic.GetCbSipRateConfig(g.ctx, g.request.GetCbSipRateInfoType())
	if err != nil {
		return err
	}

	g.response.SipRateLimitStr = proto.String(config.SipRateLimitStr)
	g.response.DefaultSipRateStr = proto.String(config.DefaultSipRateStr)
	return nil
}

func (g *getCbSipRateConfigProcessor) validateRequest() error {
	req := g.request

	if req.CbSipRateInfoType == nil {
		return cerr.New("infoType is nil", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if _, ok := priceSyncPriceCalculationPb.Constant_CbSipRateInfoType_name[int32(req.GetCbSipRateInfoType())]; !ok {
		return cerr.New("invalid infoType", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	return nil
}
