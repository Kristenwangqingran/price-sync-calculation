package processor

import (
	"context"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

func (s *CalculationServiceImpl) SetAItemMargin(ctx context.Context, request *priceSyncPriceCalculationPb.SetAItemMarginRequest, response *priceSyncPriceCalculationPb.SetAItemMarginResponse) uint32 {
	p := &SetAItemMarginProcessor{
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

type SetAItemMarginProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.SetAItemMarginRequest
	response *priceSyncPriceCalculationPb.SetAItemMarginResponse

	commonSIPLogic logic.CommonSIPLogic
}

func (g *SetAItemMarginProcessor) process() error {
	err := g.commonSIPLogic.SetAItemMargin(g.ctx, g.request.GetAShopId(), g.request.GetAItemId(), g.request.GetAItemMargin())
	if err != nil {
		return err
	}
	return nil
}
