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

func (s *CalculationServiceImpl) CalculateAPriceByPItemForLocalSip(ctx context.Context, request *priceSyncPriceCalculationPb.CalculateAPriceByPItemForLocalSIPRequest, response *priceSyncPriceCalculationPb.CalculateAPriceByPItemForLocalSIPResponse) uint32 {
	p := &calculateAPriceByPItemForLocalSipProcessor{
		ctx:           ctx,
		request:       request,
		response:      response,
		localSipLogic: s.localSipLogic,
	}

	err := p.process()
	if err != nil {
		response.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type calculateAPriceByPItemForLocalSipProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.CalculateAPriceByPItemForLocalSIPRequest
	response *priceSyncPriceCalculationPb.CalculateAPriceByPItemForLocalSIPResponse

	localSipLogic logic.LocalSipLogic
}

func (c *calculateAPriceByPItemForLocalSipProcessor) process() error {
	if err := c.validateRequest(); err != nil {
		return err
	}
	queries := make([]model.LocalSipCalculateAPriceQuery, 0, len(c.request.GetQueries()))
	for i, query := range c.request.GetQueries() {
		queries = append(queries, model.LocalSipCalculateAPriceQuery{
			QueryId:              i,
			AShopId:              query.GetAShopId(),
			ARegion:              query.GetARegion(),
			AItemId:              query.GetAItemId(),
			AModelId:             query.GetAModelId(),
			PNormalPrice:         query.GetPNormalPrice(),
			PPromotionPrices:     query.GetPPromotionPrices(),
			LeafCategoryId:       query.GetLeafCategoryId(),
			EnabledChannelIdList: query.GetEnabledChannelIdList(),
		})
	}
	results, err := c.localSipLogic.CalculateAPriceByPItemForLocalSip(c.ctx, c.request.GetPShopId(), c.request.GetPItemId(), c.request.GetPRegion(), queries, c.request.GetCalculateForCreate())
	if err != nil {
		return err
	}
	respResults := make([]*priceSyncPriceCalculationPb.LocalSipAPriceInfo, 0, len(results))
	for i, result := range results {
		if result.Err != nil {
			respResults = append(respResults, &priceSyncPriceCalculationPb.LocalSipAPriceInfo{
				ErrCode: proto.Uint32(cerr.Code(result.Err)),
				ErrMsg:  proto.String(result.Err.Error()),
			})
		} else {
			var resNormalPrice *int64
			if c.request.GetQueries()[i].PNormalPrice != nil {
				resNormalPrice = proto.Int64(result.NormalPrice)
			}
			var aItemId, aModelId *uint64
			if c.request.GetQueries()[i].AItemId != nil {
				aItemId = proto.Uint64(result.AItemId)
			}

			if c.request.GetQueries()[i].AModelId != nil {
				aModelId = proto.Uint64(result.AModelId)
			}

			respResults = append(respResults, &priceSyncPriceCalculationPb.LocalSipAPriceInfo{
				NormalPrice:     resNormalPrice,
				PromotionPrices: result.PromotionPrices,
				AShopId:         proto.Uint64(result.AShopId),
				ARegion:         proto.String(result.ARegion),
				AItemId:         aItemId,
				AModelId:        aModelId,
				Snap:            result.PriceCalSnap,
			})
		}
	}

	aShopItemIdRegionMap := make(map[uint64]map[uint64]string)
	for _, q := range c.request.GetQueries() {
		if aShopItemIdRegionMap[q.GetAShopId()] == nil {
			aShopItemIdRegionMap[q.GetAShopId()] = make(map[uint64]string)
		}
		aShopItemIdRegionMap[q.GetAShopId()][q.GetAItemId()] = q.GetARegion()
	}

	aShopItemIdOPLList := make([]*priceSyncPriceCalculationPb.ShopItemCustomizedOPL, 0)
	for aShopId, aItemIdRegionMap := range aShopItemIdRegionMap {
		for aItemId, aRegion := range aItemIdRegionMap {
			opl, err := c.localSipLogic.CalculateAItemOPL(c.ctx, c.request.GetPRegion(), c.request.GetPItemId(), aItemId, aRegion)
			if err != nil {
				return err
			}
			aShopItemIdOPLList = append(aShopItemIdOPLList, &priceSyncPriceCalculationPb.ShopItemCustomizedOPL{
				ShopId:        proto.Uint64(aShopId),
				ItemId:        proto.Uint64(aItemId),
				CustomizedOpl: opl,
			})
		}
	}

	c.response.Results = respResults
	c.response.AShopItemCustomizedOpls = aShopItemIdOPLList
	return nil
}

func (c *calculateAPriceByPItemForLocalSipProcessor) validateRequest() error {
	req := c.request
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

	if len(req.GetQueries()) == 0 {
		return cerr.New("empty queries", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	for _, query := range c.request.Queries {
		if query.GetAShopId() == 0 {
			return cerr.New("invalid AShopId", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
		if len(query.GetARegion()) == 0 {
			return cerr.New("invalid ARegion", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if !req.GetCalculateForCreate() && query.AItemId == nil {
			return cerr.New("a_item_id is required for non-creation scenario", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if !cutil.IsValidCountry(query.GetARegion()) {
			return cerr.New(fmt.Sprintf("region %v is invalid", query.GetARegion()), uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
		if query.PNormalPrice == nil && len(query.GetPPromotionPrices()) == 0 {
			return cerr.New("PNormalPrice and PPromotionPrices should be provided at least one", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if query.PNormalPrice != nil && query.GetPNormalPrice() <= 0 {
			return cerr.New("invalid PNormalPrice", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		for _, price := range query.GetPPromotionPrices() {
			if price <= 0 {
				return cerr.New("invalid p promotion price", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
			}
		}
	}
	return nil
}
