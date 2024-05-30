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

func (s *CalculationServiceImpl) CreateCbSipAShopSellerDiscountPromotion(ctx context.Context, request *priceSyncPriceCalculationPb.CreateCBSIPAShopSellerDiscountPromotionRequest, response *priceSyncPriceCalculationPb.CreateCBSIPAShopSellerDiscountPromotionResponse) uint32 {
	p := &CreateCBSIPAShopSellerDiscountPromotionProcessor{
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

type CreateCBSIPAShopSellerDiscountPromotionProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.CreateCBSIPAShopSellerDiscountPromotionRequest
	response *priceSyncPriceCalculationPb.CreateCBSIPAShopSellerDiscountPromotionResponse

	commonSIPLogic logic.CommonSIPLogic
}

func (c *CreateCBSIPAShopSellerDiscountPromotionProcessor) process() error {
	err := c.commonSIPLogic.CreateCBSIPAShopSellerDiscountPromotion(c.ctx, c.request.GetAShopId())
	if err != nil {
		return err
	}
	return nil
}
