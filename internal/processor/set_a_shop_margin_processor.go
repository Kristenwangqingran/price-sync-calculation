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

func (s *CalculationServiceImpl) SetAShopMargin(ctx context.Context, request *priceSyncPriceCalculationPb.SetAShopMarginRequest, response *priceSyncPriceCalculationPb.SetAShopMarginResponse) uint32 {
	p := &SetAShopMarginProcessor{
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

type SetAShopMarginProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.SetAShopMarginRequest
	response *priceSyncPriceCalculationPb.SetAShopMarginResponse

	commonSIPLogic logic.CommonSIPLogic
}

func (g *SetAShopMarginProcessor) process() error {
	err := g.commonSIPLogic.SetAShopMargin(g.ctx, g.request.GetShopId(), g.request.GetMargin())
	if err != nil {
		return err
	}
	return nil
}
