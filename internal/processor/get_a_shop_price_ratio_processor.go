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

func (s *CalculationServiceImpl) GetAShopPriceRatio(ctx context.Context, request *priceSyncPriceCalculationPb.GetAShopPriceRatioRequest, response *priceSyncPriceCalculationPb.GetAShopPriceRatioResponse) uint32 {
	p := &GetAShopPriceRatioProcessor{
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

type GetAShopPriceRatioProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetAShopPriceRatioRequest
	response *priceSyncPriceCalculationPb.GetAShopPriceRatioResponse

	commonSIPLogic logic.CommonSIPLogic
}

func (g *GetAShopPriceRatioProcessor) process() error {
	aShopPriceRatios, err := g.commonSIPLogic.GetAShopPriceRatioBatch(g.ctx, g.request.GetShopIds())
	if err != nil {
		return err
	}
	for aShopId, priceRatio := range aShopPriceRatios {
		g.response.AShopPriceRatios = append(g.response.AShopPriceRatios, &priceSyncPriceCalculationPb.ShopPriceRatio{
			ShopId:     proto.Uint64(aShopId),
			PriceRatio: proto.Int64(priceRatio),
		})
	}
	return nil
}
