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
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/convutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

func (s *CalculationServiceImpl) CalculateSipItemPriceForCbSip(ctx context.Context, request *priceSyncPriceCalculationPb.CalculateSipItemPriceForCbSipRequest, response *priceSyncPriceCalculationPb.CalculateSipItemPriceForCbSipResponse) uint32 {
	p := &calculateSipItemPriceForCbSipProcessor{
		ctx:        ctx,
		request:    request,
		response:   response,
		cbSipLogic: s.cbsipLogic,
	}

	err := p.process()
	if err != nil {
		response.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type calculateSipItemPriceForCbSipProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.CalculateSipItemPriceForCbSipRequest
	response *priceSyncPriceCalculationPb.CalculateSipItemPriceForCbSipResponse

	cbSipLogic logic.CbSipLogic
}

func (c *calculateSipItemPriceForCbSipProcessor) process() error {
	if err := c.validateRequest(); err != nil {
		return err
	}

	queries := make([]model.CbSipCalculateSipItemPriceSingleQuery, 0, len(c.request.GetQueries()))
	for _, query := range c.request.GetQueries() {
		queries = append(queries, model.CbSipCalculateSipItemPriceSingleQuery{
			ModelId: query.GetModelId(),
			Price:   query.GetPrice(),
		})
	}
	results, err := c.cbSipLogic.CalculateSipItemPriceForCbSip(c.ctx, model.CbSipCalculateSipItemPriceRequest{
		ShopId:         c.request.GetShopId(),
		Region:         c.request.GetRegion(),
		ItemId:         c.request.GetItemId(),
		ChannelIdList:  convutil.Uint64sToUint32s(c.request.GetChannelIdList()),
		LeafCategoryId: c.request.GetLeafCategoryId(),
		Weight:         c.request.GetWeight(),
		Queries:        queries,
	})
	if err != nil {
		return err
	}
	respResults := make([]*priceSyncPriceCalculationPb.CbSipItemPriceInfo, 0, len(results))
	for _, result := range results {
		respResults = append(respResults, &priceSyncPriceCalculationPb.CbSipItemPriceInfo{
			ModelId:        proto.Uint64(result.ModelId),
			CbSipItemPrice: proto.Int64(result.CbSipItemPrice),
			Currency:       proto.String(result.Currency),
		})
	}

	c.response.CbSipItemPriceInfoList = respResults
	return nil
}

func (c *calculateSipItemPriceForCbSipProcessor) validateRequest() error {
	req := c.request
	if req.GetShopId() == 0 {
		return cerr.New("invalid ShopId", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	if len(req.GetRegion()) == 0 {
		return cerr.New("invalid Region", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	if !cutil.IsValidCountry(req.GetRegion()) {
		return cerr.New(fmt.Sprintf("region %v is invalid", req.GetRegion()), uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if req.GetLeafCategoryId() == 0 {
		return cerr.New("invalid LeafCategoryId", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	if req.GetWeight() == 0 {
		return cerr.New("invalid Weight", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	if len(req.GetQueries()) == 0 {
		return cerr.New("empty queries", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}
	for _, query := range req.GetQueries() {
		if query.GetPrice() <= 0 {
			return cerr.New("invalid Price", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
	}

	return nil
}
