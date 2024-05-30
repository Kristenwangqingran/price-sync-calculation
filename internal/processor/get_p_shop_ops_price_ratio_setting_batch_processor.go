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

func (s *CalculationServiceImpl) GetPShopOpsPriceRatioSettingBatch(ctx context.Context, request *priceSyncPriceCalculationPb.GetPShopOpsPriceRatioSettingBatchRequest, response *priceSyncPriceCalculationPb.GetPShopOpsPriceRatioSettingBatchResponse) uint32 {
	p := &GetPShopOpsPriceRatioSettingBatchProcessor{
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

type GetPShopOpsPriceRatioSettingBatchProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetPShopOpsPriceRatioSettingBatchRequest
	response *priceSyncPriceCalculationPb.GetPShopOpsPriceRatioSettingBatchResponse

	commonSIPLogic logic.CommonSIPLogic
}

func (g *GetPShopOpsPriceRatioSettingBatchProcessor) process() error {
	pShopOpsPriceRatioSettings, err := g.commonSIPLogic.GetPShopOpsPriceRatioSettingBatch(g.ctx, g.request.GetPShopIds())
	if err != nil {
		return err
	}
	g.response.OpsPriceRatioSetting = pShopOpsPriceRatioSettings
	return nil
}
