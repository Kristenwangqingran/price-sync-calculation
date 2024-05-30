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

func (s *CalculationServiceImpl) GetAItemMargin(ctx context.Context, request *priceSyncPriceCalculationPb.GetAItemMarginRequest, response *priceSyncPriceCalculationPb.GetAItemMarginResponse) uint32 {
	p := &GetAItemMarginProcessor{
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

type GetAItemMarginProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetAItemMarginRequest
	response *priceSyncPriceCalculationPb.GetAItemMarginResponse

	commonSIPLogic logic.CommonSIPLogic
}

func (g *GetAItemMarginProcessor) process() error {
	//TODO: implement after listing provides
	aShopIdToItemIdMap := make(map[uint64][]uint64)
	for _, aShopIdToAItemIds := range g.request.GetShopIdToItemIdsList() {
		aShopId := aShopIdToAItemIds.GetShopId()
		itemIds := aShopIdToAItemIds.GetItemIds()
		if len(itemIds) == 0 {
			ulog.DefaultLoggerFromContext(g.ctx).Info("aShopId in request ignored since no aItemId was provided", ulog.Uint64("aShopId", aShopId))
			continue
		}
		aShopIdToItemIdMap[aShopId] = itemIds
	}
	aItemMarginMap, err := g.commonSIPLogic.GetAItemMarginBatch(g.ctx, aShopIdToItemIdMap)
	if err != nil {
		return err
	}
	aItemMargins := make([]*priceSyncPriceCalculationPb.ItemMargin, 0, len(aItemMarginMap))
	for aItemId, margin := range aItemMarginMap {
		aItemMargins = append(aItemMargins, &priceSyncPriceCalculationPb.ItemMargin{
			ItemId:     proto.Uint64(aItemId),
			ItemMargin: proto.Int64(int64(margin)),
		})
	}
	g.response.AItemMargins = aItemMargins
	return nil

}
