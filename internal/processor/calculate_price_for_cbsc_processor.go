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

func (s *CalculationServiceImpl) CalculatePriceForCbsc(ctx context.Context, request *priceSyncPriceCalculationPb.CalculatePriceForCbscRequest, response *priceSyncPriceCalculationPb.CalculatePriceForCbscResponse) uint32 {
	p := &calculatePriceForCbscProcessor{
		ctx:       ctx,
		request:   request,
		response:  response,
		cbscLogic: s.cbscLogic,
	}

	err := p.process()
	if err != nil {
		response.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type calculatePriceForCbscProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.CalculatePriceForCbscRequest
	response *priceSyncPriceCalculationPb.CalculatePriceForCbscResponse

	cbscLogic logic.CbscLogic
}

func (c *calculatePriceForCbscProcessor) process() error {
	if err := c.validateRequest(); err != nil {
		return err
	}

	queries := make([]model.MtskuMpskuPriceQuery, 0, len(c.request.GetQueries()))
	for _, query := range c.request.GetQueries() {
		queries = append(queries, model.MtskuMpskuPriceQuery{
			SourcePrice:       query.GetSrcPrice(),
			MpskuShopId:       query.GetMpskuShopId(),
			MpskuRegion:       query.GetMpskuRegion(),
			MpskuItemId:       query.GetMpskuItemId(),
			Weight:            query.GetWeight(),
			LeafCategoryId:    query.GetLeafCategoryId(),
			EnabledChannelIds: query.GetEnabledChannelIdList(),
		})

	}
	results, err := c.cbscLogic.CalculatePriceForCbsc(c.ctx, c.request.GetMerchantId(), c.request.GetIsMtskuToMpsku(), queries)
	if err != nil {
		return err
	}
	respResults := make([]*priceSyncPriceCalculationPb.MtskuMpskuPriceQueryInfo, 0, len(results))
	for _, result := range results {
		if result.Err != nil {
			respResults = append(respResults, &priceSyncPriceCalculationPb.MtskuMpskuPriceQueryInfo{
				ErrCode:        proto.Uint32(cerr.Code(result.Err)),
				ErrMsg:         proto.String(result.Err.Error()),
				HidePriceError: proto.Int32(result.HidePriceErrorCode),
			})
		} else {
			respResults = append(respResults, &priceSyncPriceCalculationPb.MtskuMpskuPriceQueryInfo{
				DstPrice:       proto.Int64(result.DstPrice),
				HidePrice:      proto.Int64(result.HidePrice),
				HidePriceError: proto.Int32(result.HidePriceErrorCode),
			})
		}
	}

	c.response.Results = respResults
	return nil
}

func (c *calculatePriceForCbscProcessor) validateRequest() error {
	req := c.request
	if req.GetMerchantId() == 0 {
		return cerr.New("invalid MerchantId", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if len(req.GetQueries()) == 0 {
		return cerr.New("empty queries", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	for _, query := range req.GetQueries() {
		if query.GetSrcPrice() <= 0 {
			return cerr.New("invalid SrcPrice", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
		if query.GetMpskuShopId() == 0 {
			return cerr.New("invalid MpskuShopId", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
		if len(query.GetMpskuRegion()) == 0 {
			return cerr.New("invalid MpskuRegion", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
		if !cutil.IsValidCountry(query.GetMpskuRegion()) {
			return cerr.New(fmt.Sprintf("region %v is invalid", query.GetMpskuRegion()), uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
	}
	return nil
}
