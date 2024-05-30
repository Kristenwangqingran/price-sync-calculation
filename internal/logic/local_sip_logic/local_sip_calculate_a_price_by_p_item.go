package local_sip_logic

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/core-logic/clog"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/calculate"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/factors"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

func (l *LocalSipLogicImpl) CalculateAPriceByPItemForLocalSip(ctx context.Context, pShopId uint64, pItemId uint64, pRegion string, queries []model.LocalSipCalculateAPriceQuery, calculateForCreate bool) ([]model.LocalSipCalculateAPriceResult, error) {
	aRegions := model.PickUniqARegionFromLocalSipCalculateAPriceQueries(queries)
	localSipConfigMap, err := l.factors.GetLocalSipConfigByRegionBatch(ctx, pRegion, aRegions)
	if err != nil {
		return nil, err
	}

	aShopIds := model.PickUniqAShopIdFromLocalSipCalculateAPriceQueries(queries)
	shopMappings, err := l.factors.GetAShopDataForLocalSip(ctx, aShopIds)
	if err != nil {
		return nil, err
	}

	aItemDataByAShopIdMap, err := l.factors.GetAItemDataBatchForLocalSip(ctx, pShopId, pItemId, aShopIds)
	if err != nil {
		return nil, err
	}

	needQueryPItemWeight := false
	for _, shopId := range aShopIds {
		if v, ok := aItemDataByAShopIdMap[shopId]; !ok || v.GetAffiRealWeight() <= 0 {
			needQueryPItemWeight = true
			break
		}
	}

	var pItemData service.PrimaryItemData
	if needQueryPItemWeight {
		pItemData, err = l.factors.GetPItemDataForLocalSip(ctx, pShopId, pItemId)
		if err != nil {
			return nil, err
		}
	}

	// P item id  + A shop id is unique already for weight,
	// and why we don't use A item id here, since it is possible to be 0 (create A item case)
	weightMapByAShopId := make(map[uint64]int64)
	for _, query := range queries {
		aWeight := int64(0)
		if val, exist := aItemDataByAShopIdMap[query.AShopId]; exist {
			aWeight = int64(val.GetAffiRealWeight())
		}

		weightMapByAShopId[query.AShopId] = factors.GetCalcWeightByAItemAndPItemWeight(aWeight, pItemData.Weight)
	}

	shippingFeeQueries := make([]model.LocalSipShippingFeeQuery, 0)
	initHiddenFeeQueries := make([]model.LocalSipHiddenPriceQuery, 0)
	for _, query := range queries {
		weight := weightMapByAShopId[query.AShopId]

		shippingFeeQueries = append(shippingFeeQueries, model.LocalSipShippingFeeQuery{
			QueryId:              query.QueryId,
			CommonConfig:         localSipConfigMap[query.ARegion],
			AShopId:              query.AShopId,
			AItemId:              query.AItemId,
			AModelId:             query.AModelId,
			ARegion:              query.ARegion,
			LeafCategoryId:       query.LeafCategoryId,
			EnabledChannelIdList: query.EnabledChannelIdList,
			Weight:               weight,
		})

		initHiddenFeeQueries = append(initHiddenFeeQueries, model.LocalSipHiddenPriceQuery{
			QueryId:      query.QueryId,
			CommonConfig: localSipConfigMap[query.ARegion],
			AShopId:      query.AShopId,
			ARegion:      query.ARegion,
			Weight:       weight,
			PRegion:      pRegion,
		})
	}

	shippingFeeResults, err := l.factors.GetShippingFeeForLocalSip(ctx, pRegion, shippingFeeQueries, calculateForCreate)
	if err != nil {
		return nil, err
	}

	if len(shippingFeeResults) != len(shippingFeeQueries) {
		return nil, cerr.New(fmt.Sprintf("len(shippingFeeResults) != len(shippingFeeQueries), queries=%+v, results=%+v", shippingFeeQueries, shippingFeeResults), uint32(pb.Constant_ERROR_INTERNAL))
	}

	hiddenPriceResults, err := l.factors.GetInitialHiddenPriceForLocalSip(ctx, pItemId, pShopId, pRegion, initHiddenFeeQueries)
	if err != nil {
		return nil, err
	}

	if len(hiddenPriceResults) != len(initHiddenFeeQueries) {
		return nil, cerr.New(fmt.Sprintf("len(hiddenPriceResults) != len(initHiddenFeeQueries), queries=%+v, results=%+v", initHiddenFeeQueries, hiddenPriceResults), uint32(pb.Constant_ERROR_INTERNAL))
	}

	finalResults := make([]model.LocalSipCalculateAPriceResult, len(queries))
	for i, query := range queries {

		shippingFeeRes := shippingFeeResults[i]
		hiddenPriceRes := hiddenPriceResults[i]
		if shippingFeeRes.Err != nil {
			finalResults[i] = model.LocalSipCalculateAPriceResult{
				Err:      shippingFeeRes.Err,
				AShopId:  query.AShopId,
				ARegion:  query.ARegion,
				AItemId:  query.AItemId,
				AModelId: query.AModelId,
			}
			continue
		}
		if hiddenPriceRes.Err != nil {
			finalResults[i] = model.LocalSipCalculateAPriceResult{
				Err:      hiddenPriceRes.Err,
				AShopId:  query.AShopId,
				ARegion:  query.ARegion,
				AItemId:  query.AItemId,
				AModelId: query.AModelId,
			}
			continue
		}

		pNormalPrice := query.PNormalPrice

		weight := weightMapByAShopId[query.AShopId]
		realWeight := calcutil.DbWeightToGram(weight)

		var itemMargin int64
		if aItemDataByAShopIdMap[query.AShopId] == nil {
			// if item map not found, only continue process for create scenario
			if !calculateForCreate {
				return nil, cerr.New(fmt.Sprintf("cannot find item map for aItemId=%v, aShopId=%v", query.AItemId, query.AShopId), uint32(pb.Constant_ERROR_NOT_FOUND))
			}
		} else {
			itemMargin = int64(aItemDataByAShopIdMap[query.AShopId].GetItemMargin())
		}
		realItemMargin := calcutil.GetItemMargin(calcutil.ToRealRatio(itemMargin))

		if shopMappings[query.AShopId] == nil {
			return nil, cerr.New(fmt.Sprintf("cannot find shop map for aShopId=%v", query.AShopId), uint32(pb.Constant_ERROR_NOT_FOUND))
		}
		shopMargin := shopMappings[query.AShopId].GetShopMargin()
		realShopMargin := calcutil.GetShopMargin(calcutil.ToRealRatio(shopMargin))

		localPriceCfg := localSipConfigMap[query.ARegion]
		realHiddenPrice := hiddenPriceRes.HiddenPrice
		realShippingFee := shippingFeeRes.ShippingFee

		realCountryMargin := calcutil.GetLocalSipCountryMargin(localPriceCfg)
		realExchangeRate := calcutil.GetLocalSipExchangeRate(localPriceCfg)

		var resultNormalPrice int64
		if query.PNormalPrice > 0 {
			localSipANormalPrice := calcutil.CalculateAffiPriceForLocalSip(ctx, calcutil.ToRealPrice(pNormalPrice), realWeight, realItemMargin, realShopMargin, realHiddenPrice, realShippingFee, localPriceCfg, nil)
			resultNormalPrice = calcutil.ToDBPrice(calcutil.PriceRoundUpByCountry(query.ARegion, localSipANormalPrice))

			logging.GetLogger(ctx).Info(fmt.Sprintf("[Local SIP] Calc local sip normal price, "+
				"query=%+v, pNormalPrice=%v, weight=%v, itemMargin=%v, shopMargin=%v, localPriceConfig=%v, realHiddenPrice=%v, realShippingFee=%v | result=%v",
				query, pNormalPrice, weight, itemMargin, shopMargin, cutil.JSONEncode(localPriceCfg), realHiddenPrice, realShippingFee, localSipANormalPrice))
		}

		promotionPrices := make([]int64, 0)
		for _, pPromotionPrice := range query.PPromotionPrices {
			localSipAPromotionPrice := calcutil.CalculateAffiPriceForLocalSip(ctx, calcutil.ToRealPrice(pPromotionPrice), realWeight, realItemMargin, realShopMargin, realHiddenPrice, realShippingFee, localPriceCfg, nil)
			promotionPrices = append(promotionPrices, calcutil.ToDBPrice(calcutil.PriceRoundUpByCountry(query.ARegion, localSipAPromotionPrice)))

			logging.GetLogger(ctx).Info(fmt.Sprintf("[Local SIP] Calc local sip promotion price, "+
				"query=%+v, pPromotionPrice=%v, weight=%v, itemMargin=%v, shopMargin=%v, localpriceConfig=%v, realHiddenPrice=%v, realShippingFee=%v | result=%v",
				query, pPromotionPrice, weight, itemMargin, shopMargin, cutil.JSONEncode(localPriceCfg), realHiddenPrice, realShippingFee, localSipAPromotionPrice))
		}

		finalResults[i] = model.LocalSipCalculateAPriceResult{
			NormalPrice:     resultNormalPrice,
			PromotionPrices: promotionPrices,
			AShopId:         query.AShopId,
			ARegion:         query.ARegion,
			AItemId:         query.AItemId,
			AModelId:        query.AModelId,
			PriceCalSnap: &pb.LocalSipPriceFactorSnap{
				Weight:          proto.Float64(realWeight),
				ShopMargin:      proto.Float64(realShopMargin),
				ItemMargin:      proto.Float64(realItemMargin),
				ShippingFee:     proto.Float64(realShippingFee),
				CountryMargin:   proto.Float64(realCountryMargin),
				ExchangeRate:    proto.Float64(realExchangeRate),
				InitHiddenPrice: proto.Float64(realHiddenPrice),
			},
		}
	}

	return finalResults, nil
}

func (l *LocalSipLogicImpl) CalculateAItemOPL(ctx context.Context, pRegion string, pItemId uint64, aShopId uint64, aRegion string) (*pb.CustomizedOPL, error) {
	if config.GetOPLConfig().CustomizedOplRegionBlackList[aRegion] {
		clog.Infof(ctx, "A shop region in CustomizedOPLRegionBlacklist, aShopId=%d, aRegion=%s", aShopId, aRegion)
		return nil, nil
	}

	oplPrices, purchaseLimit, err := l.priceBusinessService.GetOPLPrices(ctx, pItemId, pRegion)
	if err != nil {
		clog.Errorf(ctx, "error when getting OPL prices, pItemId=%d, pRegion=%s, err=%v", pItemId, pRegion, err)
		return nil, err
	}
	if len(oplPrices) == 0 {
		return nil, nil
	}

	opl := calculate.GetAItemCustomizedOPLFromPItemOPLPrices(oplPrices, purchaseLimit)

	return opl, nil
}
