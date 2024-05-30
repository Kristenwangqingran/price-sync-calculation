package processor

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

func (s *CalculationServiceImpl) CalculateAPriceByPItemForCbSip(ctx context.Context, request *priceSyncPriceCalculationPb.CalculateAPriceByPItemForCBSIPRequest, response *priceSyncPriceCalculationPb.CalculateAPriceByPItemForCBSIPResponse) uint32 {
	p := &calculateAPriceByPItemForCbSipProcessor{
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

type calculateAPriceByPItemForCbSipProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.CalculateAPriceByPItemForCBSIPRequest
	response *priceSyncPriceCalculationPb.CalculateAPriceByPItemForCBSIPResponse

	cbsipLogic logic.CbSipLogic
}

func (c *calculateAPriceByPItemForCbSipProcessor) process() error {
	if err := c.validateRequest(); err != nil {
		return err
	}

	queries := make([]model.AItemCbSipQueryId, 0)
	for _, q := range c.request.GetQueries() {
		queries = append(queries, model.AItemCbSipQueryId{
			AModelId:        q.GetAModelId(),
			PItemPrice:      q.GetPItemPrice(),
			PNormalPrice:    q.GetPNormalPrice(),
			PPromotionPrice: q.PPromotionPrice,
		})
	}
	priceResults, err := c.cbsipLogic.CalculateAPriceByPItemForCbSip(c.ctx, model.CbSipCalculateAPriceByPItemRequest{
		MerchantId:         c.request.GetMerchantId(),
		MerchantRegion:     c.request.GetMerchantRegion(),
		PShopId:            c.request.GetPShopId(),
		PRegion:            c.request.GetPRegion(),
		PItemId:            c.request.GetPItemId(),
		AShopId:            c.request.GetAShopId(),
		ARegion:            c.request.GetARegion(),
		AItemId:            c.request.GetAItemId(),
		Queries:            queries,
		CalculateForCreate: c.request.GetCalculateForCreate(),
	})
	if err != nil {
		return err
	}

	respResults := make([]*priceSyncPriceCalculationPb.AItemPriceResultInfo, 0, len(priceResults))
	for _, result := range priceResults {
		respResults = append(respResults, &priceSyncPriceCalculationPb.AItemPriceResultInfo{
			NormalPrice:             proto.Int64(result.ANormalPrice),
			SettlementPrice:         proto.Int64(result.ASettlementPrice),
			SettlementPriceCurrency: proto.String(result.ASettlementPriceCurrency),
			PromotionPrice:          proto.Int64(result.APromotionPrice),
			Snap:                    result.Snap,
		})
	}

	oplResult, err := c.cbsipLogic.CalculateAItemOPL(c.ctx, &model.CbSipCalculateAOPLByPItemRequest{
		PRegion: c.request.GetPRegion(),
		PItemId: c.request.GetPItemId(),
		AShopId: c.request.GetAShopId(),
		ARegion: c.request.GetARegion(),
	})
	if err != nil {
		return err
	}

	c.response.Results = respResults
	c.response.CustomizedOpl = oplResult.Opl

	return nil
}

func (c *calculateAPriceByPItemForCbSipProcessor) validateRequest() error {
	req := c.request
	if req.GetMerchantId() == 0 {
		return cerr.New("invalid MerchantId", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	if len(req.GetMerchantRegion()) == 0 {
		return cerr.New("invalid MerchantRegion", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	if req.GetPShopId() == 0 {
		return cerr.New("invalid PShopId", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	if len(req.GetPRegion()) == 0 {
		return cerr.New("invalid PRegion", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	if !cutil.IsValidCountry(req.GetPRegion()) {
		return cerr.New(fmt.Sprintf("region %v is invalid", req.GetPRegion()), uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if req.GetPItemId() == 0 {
		return cerr.New("invalid PItemId", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	if req.GetAShopId() == 0 {
		return cerr.New("invalid AShopId", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	if len(req.GetARegion()) == 0 {
		return cerr.New("invalid ARegion", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	if !cutil.IsValidCountry(req.GetARegion()) {
		return cerr.New(fmt.Sprintf("region %v is invalid", req.GetARegion()), uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if len(req.GetQueries()) == 0 {
		return cerr.New("empty queries", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	for _, query := range req.GetQueries() {
		if query.GetPItemPrice() <= 0 {
			return cerr.New("invalid PItemPrice", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
		if query.GetPNormalPrice() <= 0 {
			return cerr.New("invalid PNormalPrice", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
		if query.GetPPromotionPrice() < 0 || (query.PPromotionPrice != nil && query.GetPPromotionPrice() == 0) {
			return cerr.New("invalid PPromotionPrice", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
	}

	return nil
}
