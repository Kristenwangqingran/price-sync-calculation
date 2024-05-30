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

func (s *CalculationServiceImpl) GetCbSipAShopSellerDiscountPromotion(ctx context.Context, request *priceSyncPriceCalculationPb.GetCBSIPAShopSellerDiscountPromotionRequest, response *priceSyncPriceCalculationPb.GetCBSIPAShopSellerDiscountPromotionResponse) uint32 {
	p := &GetCBSIPAShopSellerDiscountPromotionProcessor{
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

type GetCBSIPAShopSellerDiscountPromotionProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetCBSIPAShopSellerDiscountPromotionRequest
	response *priceSyncPriceCalculationPb.GetCBSIPAShopSellerDiscountPromotionResponse

	commonSIPLogic logic.CommonSIPLogic
}

func (c *GetCBSIPAShopSellerDiscountPromotionProcessor) process() error {
	promoId, err := c.commonSIPLogic.GetCBSIPAShopSellerDiscountPromotion(c.ctx, c.request.GetAShopId())
	if err != nil {
		return err
	}
	c.response.PromotionId = proto.Uint64(promoId)
	return nil
}
