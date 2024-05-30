package processor

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

func (s *CalculationServiceImpl) SetAItemRealWeight(ctx context.Context, request *priceSyncPriceCalculationPb.SetAItemRealWeightRequest, response *priceSyncPriceCalculationPb.SetAItemRealWeightResponse) uint32 {
	p := &SetAItemRealWeightProcessor{
		ctx:            ctx,
		request:        request,
		response:       response,
		commonSIPLogic: s.commonSipLogic,
	}

	err := p.process()
	if err != nil {
		response.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type SetAItemRealWeightProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.SetAItemRealWeightRequest
	response *priceSyncPriceCalculationPb.SetAItemRealWeightResponse

	commonSIPLogic logic.CommonSIPLogic
}

func (g *SetAItemRealWeightProcessor) process() error {
	err := g.validateRequest()
	if err != nil {
		return err
	}
	err = g.commonSIPLogic.SetAItemRealWeight(g.ctx, g.request.GetAShopId(), g.request.GetAItemId(), g.request.GetAItemRealWeight())
	if err != nil {
		return err
	}
	return nil
}

func (g *SetAItemRealWeightProcessor) validateRequest() error {
	if int32(g.request.GetAItemRealWeight()) < 0 {
		return cerr.Wrap(fmt.Errorf("a item real weight has to be >= 0, given value converted to int32 is %d", g.request.GetAItemRealWeight()), "", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	return nil
}
