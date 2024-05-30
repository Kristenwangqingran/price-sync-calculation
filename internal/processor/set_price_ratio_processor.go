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

func (s *CalculationServiceImpl) SetPriceRatio(ctx context.Context, request *priceSyncPriceCalculationPb.SetPriceRatioRequest, response *priceSyncPriceCalculationPb.SetPriceRatioResponse) uint32 {
	p := &SetPriceRatioProcessor{
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

type SetPriceRatioProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.SetPriceRatioRequest
	response *priceSyncPriceCalculationPb.SetPriceRatioResponse

	commonSIPLogic logic.CommonSIPLogic
}

func (g *SetPriceRatioProcessor) process() error {
	err := g.validateRequest()
	if err != nil {
		return err
	}
	//TODO
	//aShopPriceRatios, err := g.commonSIPLogic.Set(g.ctx, g.request.GetShopIds())
	//if err != nil {
	//	return err
	//}
	//for aShopId, priceRatio := range aShopPriceRatios {
	//	g.response.AShopPriceRatios = append(g.response.AShopPriceRatios, &priceSyncPriceCalculationPb.ShopPriceRatio{
	//		ShopId:     proto.Uint64(aShopId),
	//		PriceRatio: proto.Int64(priceRatio),
	//	})
	//}
	return nil
}

func (g *SetPriceRatioProcessor) validateRequest() error {
	req := g.request
	if len(req.GetAShopPriceRatioSettings()) == 0 {
		return cerr.New("given a_shop_price_ratio_settings list is empty", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	return nil
}
