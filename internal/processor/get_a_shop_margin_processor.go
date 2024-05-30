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

func (s *CalculationServiceImpl) GetAShopMargin(ctx context.Context, request *priceSyncPriceCalculationPb.GetAShopMarginRequest, response *priceSyncPriceCalculationPb.GetAShopMarginResponse) uint32 {
	p := &GetAShopMarginProcessor{
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

type GetAShopMarginProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetAShopMarginRequest
	response *priceSyncPriceCalculationPb.GetAShopMarginResponse

	commonSIPLogic logic.CommonSIPLogic
}

func (g *GetAShopMarginProcessor) process() error {
	shopMargins, err := g.commonSIPLogic.GetAShopMarginBatch(g.ctx, g.request.GetShopIds())
	if err != nil {
		return err
	}
	for aShopId, margin := range shopMargins {
		g.response.AShopMargins = append(g.response.AShopMargins, &priceSyncPriceCalculationPb.ShopMargin{
			ShopId: proto.Uint64(aShopId),
			Margin: proto.Int64(margin),
		})
	}
	return nil
}
