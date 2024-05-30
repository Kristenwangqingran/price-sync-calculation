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

func (s *CalculationServiceImpl) GetAItemRealWeight(ctx context.Context, request *priceSyncPriceCalculationPb.GetAItemRealWeightRequest, response *priceSyncPriceCalculationPb.GetAItemRealWeightResponse) uint32 {
	p := &GetAItemRealWeightProcessor{
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

type GetAItemRealWeightProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetAItemRealWeightRequest
	response *priceSyncPriceCalculationPb.GetAItemRealWeightResponse

	commonSIPLogic logic.CommonSIPLogic
}

func (g *GetAItemRealWeightProcessor) process() error {
	aItemRealWeight, err := g.commonSIPLogic.GetAItemRealWeight(g.ctx, g.request.GetShopId(), g.request.GetItemId())
	if err != nil {
		return err
	}
	g.response.AItemRealWeight = proto.Int64(int64(aItemRealWeight))
	return nil

}
