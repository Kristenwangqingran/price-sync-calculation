package processor

import (
	"context"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

func (s *CalculationServiceImpl) GetCbSipShopLevelConfig(ctx context.Context, request *priceSyncPriceCalculationPb.GetCbSipShopLevelConfigRequest, response *priceSyncPriceCalculationPb.GetCbSipShopLevelConfigResponse) uint32 {
	p := &getCbSipShopLevelConfigProcessor{
		ctx:        ctx,
		request:    request,
		response:   response,
		cbsipLogic: s.cbsipLogic,
	}

	err := p.process()
	if err != nil {
		response.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type getCbSipShopLevelConfigProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetCbSipShopLevelConfigRequest
	response *priceSyncPriceCalculationPb.GetCbSipShopLevelConfigResponse

	cbsipLogic logic.CbSipLogic
}

func (g *getCbSipShopLevelConfigProcessor) process() error {
	if err := g.validateRequest(); err != nil {
		return err
	}

	config, err := g.cbsipLogic.GetCbSipShopLevelConfig(g.ctx, model.CbSipGetShopLevelConfigRequest{
		PShopId: g.request.GetPShopId(),
	})

	if err != nil {
		return err
	}

	respResults := make([]*priceSyncPriceCalculationPb.CbSipAffiShopInfo, 0)
	for _, info := range config.AShopConfigList {
		respResults = append(respResults, &priceSyncPriceCalculationPb.CbSipAffiShopInfo{
			AShopId: proto.Uint64(info.AShopId),
			SipRate: proto.Float64(info.SipRate),
		})
	}
	g.response.AShopConfigList = respResults
	return nil
}

func (g *getCbSipShopLevelConfigProcessor) validateRequest() error {
	req := g.request
	if req.GetPShopId() == 0 {
		return cerr.New("invalid PShopId", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	return nil
}
