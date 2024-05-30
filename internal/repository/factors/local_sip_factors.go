package factors

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_lps.pb"
	internal_sip "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	ibsPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/item_business.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/shop_core.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/convutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

const (
	splitForLocation = ","
)

func (c *CalculationFactorsRepoImpl) GetAllLocalSipPriceConfig(ctx context.Context) (map[string]map[string]*model.CommonPriceConfig, error) {
	return c.localSipSystemConfigService.GetAllLocalPriceConfig(ctx)
}

func (c *CalculationFactorsRepoImpl) GetPItemDataForLocalSip(ctx context.Context, pShopId uint64, pItemId uint64) (service.PrimaryItemData, error) {
	res, err := c.sipItemDataService.GetPrimaryItemDataBatch(ctx, pShopId, []uint64{pItemId})
	if err != nil {
		return service.PrimaryItemData{}, err
	}
	if _, ok := res[pItemId]; !ok {
		return service.PrimaryItemData{}, cerr.New(fmt.Sprintf("failed to get pItemData for itemid=%v, pShopId=%v", pItemId, pShopId), uint32(pb.Constant_ERROR_GET_ITEM_INFO))
	}

	return res[pItemId], nil
}

// exchange rate, buffer (region margin)
func (c *CalculationFactorsRepoImpl) GetLocalSipConfigByRegionBatch(ctx context.Context, pRegion string, aRegions []string) (map[string]*model.CommonPriceConfig, error) {
	res := make(map[string]*model.CommonPriceConfig)
	for _, aRegion := range aRegions {
		cfg, err := c.localSipSystemConfigService.GetLocalPriceConfigByRegion(ctx, pRegion, aRegion)
		if err != nil {
			return nil, err
		}
		res[aRegion] = cfg
	}
	return res, nil
}

func (c *CalculationFactorsRepoImpl) GetAShopDataForLocalSip(ctx context.Context, aShopIds []uint64) (map[uint64]*internal_sip.AShopData, error) {
	return  c.aShopDataDM.GetAShopDataBatch(ctx, aShopIds)
}

func (c *CalculationFactorsRepoImpl) GetAItemDataBatchForLocalSip(ctx context.Context, pShopId uint64, pItemId uint64, aShopIds []uint64) (map[uint64]*internal_sip.AItemData, error) {
	aShopItemIds, err := c.listingSIPService.GetAffiItemIdsByPrimaryItemId(ctx, pShopId, pItemId, aShopIds)
    if err != nil {
    	return nil, err
	}
	aItemIds := make([]uint64, 0, len(aShopItemIds))
	for _, aShopItemId := range aShopItemIds {
		aItemIds = append(aItemIds, aShopItemId.GetAitemId())
	}

	aItemIdToAItemDataMap, err := c.aItemDataDM.GetAItemDataBatch(ctx, pShopId, aItemIds)
	if err != nil {
		return nil, err
	}

	aShopIdToAItemDataMap := make(map[uint64]*internal_sip.AItemData)
	for _, aShopItemId := range aShopItemIds {
		aShopIdToAItemDataMap[aShopItemId.GetAshopId()] = aItemIdToAItemDataMap[aShopItemId.GetAitemId()]
	}

	return aShopIdToAItemDataMap, nil
}

// GetShippingFeeForLocalSip gets shipping fee from transit warehouse to buyer address.
func (c *CalculationFactorsRepoImpl) GetShippingFeeForLocalSip(ctx context.Context, pRegion string, queries []model.LocalSipShippingFeeQuery, calcForCreate bool) ([]model.LocalSipShippingFeeResult, error) {
	slsQueries, dbQueries := model.GroupLocalSipShippingFeeQueryBySlsToggle(queries)

	var err error
	var dbResults []model.LocalSipShippingFeeResult
	if len(dbQueries) > 0 {
		dbResults, err = c.getShippingFeeForLocalSipFromDb(ctx, pRegion, dbQueries)
		if err != nil {
			return nil, err
		}
	}

	var slsResults []model.LocalSipShippingFeeResult
	if len(slsQueries) > 0 {
		if calcForCreate {
			slsResults, err = c.getShippingFeeForLocalSipCreateFromSls(ctx, slsQueries)
		} else {
			slsResults, err = c.getShippingFeeForLocalSipFromSls(ctx, slsQueries)
		}
		if err != nil {
			return nil, err
		}
	}

	finalResults := make([]model.LocalSipShippingFeeResult, 0)
	finalResults = append(finalResults, dbResults...)
	finalResults = append(finalResults, slsResults...)
	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].QueryId < finalResults[j].QueryId
	})

	return finalResults, nil
}

func (c *CalculationFactorsRepoImpl) getShippingFeeForLocalSipFromDb(ctx context.Context, pRegion string, dbQueries []model.LocalSipShippingFeeQuery) ([]model.LocalSipShippingFeeResult, error) {
	results := make([]model.LocalSipShippingFeeResult, len(dbQueries))
	session := c.sipRepo.DbSession()
	for i, query := range dbQueries {
		record, err := c.sipRepo.GetLocalShippingFeeConfigRecordByWeight(ctx, session, pRegion, query.ARegion, query.Weight)

		if err != nil {
			dbErr := cerr.New(fmt.Sprintf("failed to get shipping fee, err=%v", err), uint32(pb.Constant_ERROR_DATABASE))
			results[i] = model.LocalSipShippingFeeResult{
				QueryId:     query.QueryId,
				Err:         dbErr,
				ShippingFee: 0,
			}
			continue
		}
		if record == nil {
			dbErr := cerr.New(fmt.Sprintf("cannot find shipping fee record for query=%+v", query), uint32(pb.Constant_ERROR_NOT_FOUND))
			results[i] = model.LocalSipShippingFeeResult{
				QueryId:     query.QueryId,
				Err:         dbErr,
				ShippingFee: 0,
			}
			continue
		}

		results[i] = model.LocalSipShippingFeeResult{
			QueryId:     query.QueryId,
			ShippingFee: calcutil.ToRealPrice(record.ShippingFeePrice),
		}
	}
	return results, nil
}

func (c *CalculationFactorsRepoImpl) getShippingFeeForLocalSipCreateFromSls(ctx context.Context, queries []model.LocalSipShippingFeeQuery) ([]model.LocalSipShippingFeeResult, error) {
	result := make([]model.LocalSipShippingFeeResult, len(queries))

	for i, query := range queries {
		// validate before call sls api
		if len(query.EnabledChannelIdList) == 0 {
			result[i] = model.LocalSipShippingFeeResult{
				QueryId: query.QueryId,
				Err:     cerr.New("EnabledChannelIdList is empty for getShippingFeeForLocalSipCreateFromSls", uint32(pb.Constant_ERROR_PARAMS)),
			}
			continue
		}
		if query.LeafCategoryId == 0 {
			result[i] = model.LocalSipShippingFeeResult{
				QueryId: query.QueryId,
				Err:     cerr.New("leafCategoryID not found for getShippingFeeForLocalSipCreateFromSls", uint32(pb.Constant_ERROR_PARAMS)),
			}
			continue
		}

		weightInGram := calcutil.DbWeightToGram(query.Weight)

		shippingFee, err := c.logisticService.CalcShippingFeeForLocalSIP(ctx, query.AShopId, query.ARegion, query.AItemId, query.LeafCategoryId, weightInGram, convutil.Int64sToUint32s(query.EnabledChannelIdList))

		if err != nil {
			result[i] = model.LocalSipShippingFeeResult{
				QueryId: query.QueryId,
				Err:     err,
			}
			continue
		}

		result[i] = model.LocalSipShippingFeeResult{
			QueryId:     query.QueryId,
			ShippingFee: shippingFee,
		}
	}

	return result, nil
}

func (c *CalculationFactorsRepoImpl) getShippingFeeForLocalSipFromSls(ctx context.Context, queries []model.LocalSipShippingFeeQuery) ([]model.LocalSipShippingFeeResult, error) {

	aItemProductInfoMap, err := c.getAItemProductInfos(ctx, queries)
	if err != nil {
		return nil, err
	}

	aItemEnabledChannelsMap, err := c.getAItemEnabledChannelsMap(ctx, aItemProductInfoMap, queries)
	if err != nil {
		return nil, err
	}
	aItemLeafCatIdMap, err := c.getAItemLeafCatIdMap(ctx, aItemProductInfoMap, queries)
	if err != nil {
		return nil, err
	}

	result := make([]model.LocalSipShippingFeeResult, len(queries))

	for i, query := range queries {
		enabledChannelInfo, ok := aItemEnabledChannelsMap[query.AItemId]
		if !ok {
			logging.GetLogger(ctx).Warn("calcShippingFeeFromSLS: enabledChannelIDList not found", ulog.Uint64("affiItemId", query.AItemId))
			err := cerr.New("enabledChannelIDList not found", uint32(pb.Constant_ERROR_GET_ENABLED_CHANNELS))
			result[i] = model.LocalSipShippingFeeResult{
				QueryId: query.QueryId,
				Err:     err,
			}
			continue
		}
		if enabledChannelInfo.Err != nil {
			result[i] = model.LocalSipShippingFeeResult{
				QueryId: query.QueryId,
				Err:     enabledChannelInfo.Err,
			}
			continue
		}

		if len(enabledChannelInfo.EnabledChannelIds) < 1 {
			logging.GetLogger(ctx).Warn("calcShippingFeeFromSLS: EnabledChannelIds is empty", ulog.Uint64("affiItemId", query.AItemId))
			err := cerr.New("EnabledChannelIds is empty", uint32(pb.Constant_ERROR_EMPTY_ENABLED_CHANNELS))
			result[i] = model.LocalSipShippingFeeResult{
				QueryId: query.QueryId,
				Err:     err,
			}
			continue
		}

		leafCategoryID, ok := aItemLeafCatIdMap[query.AItemId]
		if !ok {
			logging.GetLogger(ctx).Warn("calcShippingFeeFromSLS: leafCategoryID not found", ulog.Uint64("affiItemId", query.AItemId))
			err := cerr.New("leafCategoryID not found", uint32(pb.Constant_ERROR_CALCULATE_HIDDEN_FEE))
			result[i] = model.LocalSipShippingFeeResult{
				QueryId: query.QueryId,
				Err:     err,
			}
			continue
		}
		if leafCategoryID.Err != nil {
			result[i] = model.LocalSipShippingFeeResult{
				QueryId: query.QueryId,
				Err:     leafCategoryID.Err,
			}
			continue
		}

		weightInGram := calcutil.DbWeightToGram(query.Weight)

		shippingFee, err := c.logisticService.CalcShippingFeeForLocalSIP(ctx,
			uint64(query.AShopId), query.ARegion, query.AItemId,
			uint64(leafCategoryID.LeafCatId), weightInGram, enabledChannelInfo.EnabledChannelIds)

		if err != nil {
			result[i] = model.LocalSipShippingFeeResult{
				QueryId: query.QueryId,
				Err:     err,
			}
			continue
		}

		result[i] = model.LocalSipShippingFeeResult{
			QueryId:     query.QueryId,
			ShippingFee: shippingFee,
		}
	}

	return result, nil
}

func (c *CalculationFactorsRepoImpl) getAItemProductInfos(ctx context.Context, queries []model.LocalSipShippingFeeQuery) (map[uint64]*ibsPb.ProductInfo, error) {
	aItemIdsByARegion := make(map[string][]uint64)
	for _, query := range queries {
		aItemIdsByARegion[query.ARegion] = append(aItemIdsByARegion[query.ARegion], query.AItemId)
	}

	result := c.itemService.GetProductInfoMapForMixedRegion(ctx, aItemIdsByARegion)
	return result, nil
}

func (c *CalculationFactorsRepoImpl) getAItemEnabledChannelsMap(ctx context.Context, aItemProductInfos map[uint64]*ibsPb.ProductInfo, queries []model.LocalSipShippingFeeQuery) (map[uint64]*service.EnabledChannelIdsInfo, error) {
	aItemIds := model.PickAItemIdsFromLocalShippingShippingFeeQuery(queries)
	enabledChannlesMap := c.itemService.GetItemEnableChannelIdsMap(ctx, aItemIds, aItemProductInfos)
	return enabledChannlesMap, nil
}

func (c *CalculationFactorsRepoImpl) getAItemLeafCatIdMap(ctx context.Context, aItemProductInfos map[uint64]*ibsPb.ProductInfo, queries []model.LocalSipShippingFeeQuery) (map[uint64]*service.ItemLeafCategoryInfo, error) {
	aItemIds := model.PickAItemIdsFromLocalShippingShippingFeeQuery(queries)
	res := c.itemService.GetItemLeafCatIdMap(ctx, aItemIds, aItemProductInfos)
	return res, nil
}

// GetInitialHiddenPriceForLocalSip gets hidden price for local sip.
// The hidden price is the shipping cost from seller pickup address to transit warehouse delivery address.
// The related docs are:
// - hidden price calculation by SLS https://confluence.shopee.io/display/SCPM/%5BSPML-15978%5DLocal+SIP+Pricing+formula+update
// - multiple warehouse https://confluence.shopee.io/display/SCPM/%5BSPML-17283%5D%5Blocal+SC%5D%5BMW%5DMulti-Warehouse+Phase+2
// - multiple warehouse & use location id for SLS api https://confluence.shopee.io/pages/viewpage.action?pageId=1795010579
func (c *CalculationFactorsRepoImpl) GetInitialHiddenPriceForLocalSip(ctx context.Context, pItemId uint64, pShopId uint64, pRegion string, queries []model.LocalSipHiddenPriceQuery) ([]model.LocalSipHiddenPriceResult, error) {
	slsQueries, dbQueries := model.GroupLocalSipHiddenPriceQueryBySlsToggleStatus(queries)

	var err error
	var dbHiddenPrices []model.LocalSipHiddenPriceResult
	if len(dbQueries) > 0 {
		dbHiddenPrices, err = c.getInitialHiddenPriceForLocalSipFromDb(ctx, pRegion, dbQueries)
		if err != nil {
			return nil, err
		}
	}

	var slsHiddenPrices []model.LocalSipHiddenPriceResult
	if len(slsQueries) > 0 {
		slsHiddenPrices, err = c.getInitialHiddenPriceForLocalSipFromSls(ctx, pItemId, pShopId, pRegion, slsQueries)
		if err != nil {
			return nil, err
		}
	}

	result := make([]model.LocalSipHiddenPriceResult, 0)
	result = append(result, dbHiddenPrices...)
	result = append(result, slsHiddenPrices...)
	sort.Slice(result, func(i, j int) bool {
		return result[i].QueryId < result[j].QueryId
	})

	return result, nil
}

// getInitialHiddenPriceForLocalSipFromSls calculates hidden fee by SLS api.
// (1) for multiple warehouse seller, there are multiple seller locations, so we will:
//
//	a) get hidden fee for each seller location, which means get max hidden fee among all channel and buyer location.
//	b) get mean value among all seller locations.
//
// (2) for non-multiple warehouse seller and use SLS calculation logic for the A region,
// we will use seller pickup location to calculate,
// and choose max hidden fee among all channel and buyer location directly.

// Note: one location has multiple location ids,
// and follows the format [state location id, city location id, district location id, street location id]
func (c *CalculationFactorsRepoImpl) getInitialHiddenPriceForLocalSipFromSls(ctx context.Context, pItemId uint64, pShopId uint64, pRegion string, localSipHiddenPriceAItemQueries []model.LocalSipHiddenPriceQuery) ([]model.LocalSipHiddenPriceResult, error) {
	// get seller user id
	pShopInfo, err := c.shopCoreService.GetShopDetail(ctx, pShopId, pRegion)
	if err != nil {
		return nil, err
	}
	sellerUserId := proto.Uint64(uint64(pShopInfo.UserId))

	// get seller pickup address ids
	availSellerAddressIds := make([]uint64, 0)
	isSellerMultiWarehouseShop, _, err := c.shopCoreService.IsSellerWarehouseShop(ctx, pShopId, pRegion)
	if err != nil {
		return nil, err
	}
	if isSellerMultiWarehouseShop { // if is multiple warehouse, then has multiple address ids which are from shop side
		shopWarehouses, err := c.shopCoreService.GetShopWarehouseByShopId(ctx, pShopId, pRegion)
		if err != nil {
			return nil, err
		}

		availSellerAddressIdsMap := make(map[uint64]bool, 0)
		for _, sw := range shopWarehouses {
			if sw != nil && sw.GetHolidayMode() == int32(shop_core.Constant_WAREHOUSE_MODE_OFF) {
				availSellerAddressIdsMap[uint64(sw.GetAddressId())] = true
			}
		}

		for s := range availSellerAddressIdsMap {
			availSellerAddressIds = append(availSellerAddressIds, s)
		}
	} else { // if is non-multiple warehouse, then use pickup address id for calculation, and only has 1 address id
		availSellerAddressIds = append(availSellerAddressIds, uint64(pShopInfo.PickupAddressId))
	}
	if len(availSellerAddressIds) == 0 {
		return nil, cerr.New("failed to calculate hidden fee, due to empty availSellerAddressIds",
			uint32(pb.Constant_ERROR_CALCULATE_HIDDEN_FEE))
	}

	// get transit warehouse buyer user id
	dummyBuyerId, err := c.logisticService.GetDummyBuyerUserID(ctx, pRegion)
	if err != nil {
		return nil, err
	}

	// get transit warehouse delivery address ids
	pItemProductInfoMap := c.itemService.GetProductInfoMapForSameRegion(ctx, []uint64{pItemId}, pRegion)
	if pItemProductInfoMap[pItemId] == nil {
		return nil, cerr.New(fmt.Sprintf("failed to get pItemInfo for pItemId=%v", pItemId),
			uint32(pb.Constant_ERROR_GET_ITEM_INFO))
	}
	pItemSlsChannelList, err := c.getPItemChannelList(ctx, pItemId, pItemProductInfoMap)
	if err != nil {
		return nil, err
	}
	transitWarehouseDeliveryAddressIdByChannelIdMap := getTransitWarehouseDeliveryAddressIds(pItemSlsChannelList, pRegion)
	if len(transitWarehouseDeliveryAddressIdByChannelIdMap) == 0 {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to calculate hidden fee, due to empty transitWarehouseDeliveryAddressIds. "+
				"pItemSlsChannelList=%v, pRegion=%s",
			pItemSlsChannelList, pRegion))

		return nil, cerr.New("failed to calculate hidden fee, due to empty transitWarehouseDeliveryAddressIds",
			uint32(pb.Constant_ERROR_CALCULATE_HIDDEN_FEE))
	}

	// get seller pickup locations queries (one location has multiple location id for detail address)
	sellerAndBuyerAddressQueries := make([]*internal.AddressQuery, 0)
	uniqueId := 0
	for _, addressId := range availSellerAddressIds {
		sellerAndBuyerAddressQueries = append(sellerAndBuyerAddressQueries, &internal.AddressQuery{
			UniqueId:        proto.String(fmt.Sprintf("%d", uniqueId)),
			SellerId:        sellerUserId,
			SellerAddressId: proto.Uint64(addressId),
		})
		uniqueId++
	}

	// get transit warehouse delivery locations queries (one location has multiple location id for detail address)
	for channelId, deliveryAddressId := range transitWarehouseDeliveryAddressIdByChannelIdMap {
		sellerAndBuyerAddressQueries = append(sellerAndBuyerAddressQueries, &internal.AddressQuery{
			UniqueId:       proto.String(fmt.Sprintf("%d", uniqueId)),
			BuyerId:        proto.Uint64(uint64(dummyBuyerId)),
			ChannelId:      proto.Uint32(channelId),
			BuyerAddressId: proto.Uint64(deliveryAddressId),
		})
		uniqueId++
	}

	// get seller & buyer location via address id
	availSellerLocationByUniqIdMap, availBuyerLocationIdsByChannelIdMap, err := c.getLocationIdByAddressId(ctx, pRegion, sellerAndBuyerAddressQueries)
	if err != nil {
		return nil, err
	}
	if len(availSellerLocationByUniqIdMap) == 0 {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to calculate hidden fee, due to availSellerLocation is empty. "+
				"isSellerMultiWarehouseShop=%v, sellerUserId=%v, sellerAddressIds=%v",
			isSellerMultiWarehouseShop, pShopInfo.UserId, availSellerAddressIds,
		))
		return nil, cerr.New("failed to calculate hidden fee, due to availSellerLocation is empty",
			uint32(pb.Constant_ERROR_CALCULATE_HIDDEN_FEE))
	}
	if len(availBuyerLocationIdsByChannelIdMap) == 0 {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to calculate hidden fee, due to availBuyerLocation is empty. "+
				"isSellerMultiWarehouseShop=%v, dummyBuyerUserId=%v, buyerAddressIdsByChannelId=%+v",
			isSellerMultiWarehouseShop, dummyBuyerId, transitWarehouseDeliveryAddressIdByChannelIdMap,
		))
		return nil, cerr.New("failed to calculate hidden fee, due to availBuyerLocation is empty",
			uint32(pb.Constant_ERROR_CALCULATE_HIDDEN_FEE))
	}

	// group channel ids by buyer location,
	// since we wanna query sls calculation api together with those channels which are under same buyer location.
	channelIdsByBuyerLocStrMap := make(map[string][]uint32)
	for channelId, buyerLoc := range availBuyerLocationIdsByChannelIdMap {
		buyerLocStr := convLocFromIntsToStr(buyerLoc)
		if _, exist := channelIdsByBuyerLocStrMap[buyerLocStr]; !exist {
			channelIdsByBuyerLocStrMap[buyerLocStr] = make([]uint32, 0)
		}
		channelIdsByBuyerLocStrMap[buyerLocStr] = append(channelIdsByBuyerLocStrMap[buyerLocStr], channelId)
	}

	// get p item leaf category id for SLS calculation
	pItemLeafCatId, err := c.getPItemLeafCatId(ctx, pItemId, pItemProductInfoMap)
	if err != nil {
		return nil, err
	}

	// hiddenPriceTotalForAItemQueries and hiddenPriceCntForAItemQueries store hidden price data for different A region,
	// the length and order is same like localSipHiddenPriceAItemQueries.
	hiddenPriceTotalForAItemQueries := make([]float64, len(localSipHiddenPriceAItemQueries))
	hiddenPriceCntForAItemQueries := make([]int, len(localSipHiddenPriceAItemQueries))
	for _, sellerLoc := range availSellerLocationByUniqIdMap { // fetch hidden fee for each seller location
		slsQueries := make([]*model.SlsHiddenPriceQuery, 0)                             // sls hidden fee calculation queries for all a region and buyer locations
		hiddenPriceQueryIdxForSlsQueryMap := make(map[model.SlsHiddenPriceQueryKey]int) // hiddenPriceQueryIdxForSlsQueryMap stores sls query -> query idx / result idx

		for _, aQuery := range localSipHiddenPriceAItemQueries {
			for buyerLocStr, channelIds := range channelIdsByBuyerLocStrMap {
				buyerLoc, err := convLocFromStrToInts(buyerLocStr)
				if err != nil {
					return nil, err
				}

				slsQueryKey := model.SlsHiddenPriceQueryKey{
					WeightInDB:          aQuery.Weight,
					DeliveryLocationStr: buyerLocStr,
				}
				slsQuery := &model.SlsHiddenPriceQuery{
					PItemEnabledChannelIds: channelIds,
					DeliveryLocationIds:    buyerLoc,
					WeightInGram:           calcutil.DbWeightToGram(aQuery.Weight),
				}

				// since for a item query, the only possible difference is weight for sls query.
				// so here we use hiddenPriceQueryIdxForSlsQueryMap to make sls queries unique and no same queries.
				if _, exist := hiddenPriceQueryIdxForSlsQueryMap[slsQueryKey]; exist { // if alr exist, then skip
					continue
				}

				// if not exist, then append one sls query and record query index
				slsQueries = append(slsQueries, slsQuery)
				hiddenPriceQueryIdxForSlsQueryMap[slsQueryKey] = len(slsQueries) - 1
			}
		}

		slsResList, err := c.logisticService.CalcHiddenFeeForLocalSip(
			ctx, pShopId, pRegion, pItemId, pItemLeafCatId, sellerLoc, slsQueries)
		if err != nil { // if has error, then ignore and continue to calculate for next seller location
			logging.GetLogger(ctx).Warn(fmt.Sprintf("failed to calculate hidden fee for this seller location id=%v, err=%s",
				sellerLoc, err.Error()))
			continue
		}

		if len(slsResList) != len(slsQueries) {
			logging.GetLogger(ctx).Warn(fmt.Sprintf("failed to calculate hidden fee for this seller location id=%v, "+
				"due to response length from sls is unexpected, request length=%d, response length=%d",
				sellerLoc, len(slsQueries), len(slsResList)))
			continue
		}

		for idx, aQuery := range localSipHiddenPriceAItemQueries { // for each A region, handle separately
			var iMaxEsf float64 // fetch max hidden fee among all channels and buyer locations, for one A region
			var success bool

			for buyerLocStr, channelIds := range channelIdsByBuyerLocStrMap { // for each buyer location and its channels
				buyerLoc, err := convLocFromStrToInts(buyerLocStr)
				if err != nil {
					return nil, err
				}

				slsQueryKey := model.SlsHiddenPriceQueryKey{
					WeightInDB:          aQuery.Weight,
					DeliveryLocationStr: buyerLocStr,
				}
				slsQuery := &model.SlsHiddenPriceQuery{
					PItemEnabledChannelIds: channelIds,
					DeliveryLocationIds:    buyerLoc,
					WeightInGram:           calcutil.DbWeightToGram(aQuery.Weight),
				}

				if _, exist := hiddenPriceQueryIdxForSlsQueryMap[slsQueryKey]; !exist { // usually won't happen
					logging.GetLogger(ctx).Warn(fmt.Sprintf(
						"cannot find this slsQuery index, slsQuery=%v",
						cutil.JSONEncode(slsQuery)))
					continue
				}

				resIdx := hiddenPriceQueryIdxForSlsQueryMap[slsQueryKey]
				if resIdx >= len(slsResList) || resIdx < 0 { // usually won't happen, since we checked length of queries and results alr
					logging.GetLogger(ctx).Warn(fmt.Sprintf(
						"this slsQuery index is not in sls result, slsQuery=%v, resIdx=%d, slsResList length=%d",
						cutil.JSONEncode(slsQuery), resIdx, len(slsResList)))
					continue
				}

				for _, channelRes := range slsResList[resIdx] { // for each buyer location and its channels
					if channelRes.RetCode != 0 {
						logging.GetLogger(ctx).Warn(fmt.Sprintf("failed to calculate hidden fee, "+
							"sellerLocationId=%v, channelId=%v, code=%v, msg=%v",
							sellerLoc, channelRes.ProductID, channelRes.RetCode, channelRes.Message))
						continue
					}

					if channelRes.ESF > iMaxEsf {
						iMaxEsf = channelRes.ESF
						success = true
					}
				}
			}

			if success { // only add when success, and if failed for this seller location, ignore and continue for next one
				hiddenPriceTotalForAItemQueries[idx] += iMaxEsf
				hiddenPriceCntForAItemQueries[idx]++

				logging.GetLogger(ctx).Info(fmt.Sprintf("calculate hidden fee successfully, "+
					"query=%v, sellerLoc=%v, currentMaxHiddenFee(before divide exchange rate)=%v",
					cutil.JSONEncode(localSipHiddenPriceAItemQueries[idx]), sellerLoc, iMaxEsf))
			}
		}
	}

	result := make([]model.LocalSipHiddenPriceResult, len(localSipHiddenPriceAItemQueries))
	for idx, query := range localSipHiddenPriceAItemQueries {
		if hiddenPriceCntForAItemQueries[idx] > 0 {
			exchangeRate := *query.CommonConfig.ExchangeRate
			meanMaxHiddenPrice := hiddenPriceTotalForAItemQueries[idx] / float64(hiddenPriceCntForAItemQueries[idx]) / exchangeRate
			result[idx] = model.LocalSipHiddenPriceResult{
				QueryId:     query.QueryId,
				HiddenPrice: meanMaxHiddenPrice,
			}

			logging.GetLogger(ctx).Info(fmt.Sprintf("calculate hidden fee successfully. "+
				"query=%v, hidden fee(%v) = hiddenPriceTotal(%v) / hiddenPriceCnt(%v) / exchangeRate(%v)",
				cutil.JSONEncode(query), meanMaxHiddenPrice, hiddenPriceTotalForAItemQueries[idx], hiddenPriceCntForAItemQueries[idx], exchangeRate))
		} else { // if all failed, then return error, otherwise use successful results to calculate
			result[idx] = model.LocalSipHiddenPriceResult{
				QueryId: query.QueryId,
				Err: cerr.New(fmt.Sprintf("failed to calculate hidden fee among all channels, queryId=%v, aRegion=%s",
					query.QueryId, query.ARegion),
					uint32(priceSyncPriceCalculationPb.Constant_ERROR_CALCULATE_HIDDEN_FEE)),
			}
		}
	}

	return result, nil
}

func convLocFromIntsToStr(locIds []uint64) string {
	s := ""
	for idx, locId := range locIds {
		if idx != 0 {
			s += splitForLocation
		}
		s += fmt.Sprintf("%d", locId)
	}

	return s
}

func convLocFromStrToInts(locIdStr string) ([]uint64, error) {
	a := make([]uint64, 0)
	locIdStrs := strings.Split(locIdStr, splitForLocation)

	for _, locIdS := range locIdStrs {
		locId, err := strconv.Atoi(locIdS)
		if err != nil {
			return nil, cerr.New("failed to conv location from string to int list",
				uint32(pb.Constant_ERROR_INTERNAL))
		}
		a = append(a, uint64(locId))
	}

	return a, nil
}

// getTransitWarehouseDeliveryAddressIds gets transit warehouse address ids,
// 1 p channel id + 1 p region -> 1 delivery address id.
// now 1 channel should only have 1 p region, so safe to use channel id as result map key.
func getTransitWarehouseDeliveryAddressIds(pItemSlsChannelList []uint32, pRegion string) map[uint32]uint64 {
	transitWarehouseDeliveryAddressIdByChannelIdMap := make(map[uint32]uint64)

	for _, channelId := range pItemSlsChannelList {
		addressId := config.GetSlsTransitWarehouseDeliveryAddressId(channelId, pRegion)
		if addressId != nil {
			transitWarehouseDeliveryAddressIdByChannelIdMap[channelId] = *addressId
		}
	}

	return transitWarehouseDeliveryAddressIdByChannelIdMap
}

// getLocationIdByAddressId gets location id via address id
func (c *CalculationFactorsRepoImpl) getLocationIdByAddressId(ctx context.Context, region string, addressQueries []*internal.AddressQuery) (map[string][]uint64, map[uint32][]uint64, error) {
	locationInfoMapByUniqIdMap, err := c.logisticService.GetSlsLocationInfoByAddressInfoBatch(ctx, region, addressQueries)
	if err != nil {
		return nil, nil, err
	}

	availSellerLocationIdsByUniqIdMap := make(map[string][]uint64, 0)
	availBuyerLocationIdsByChannelIdMap := make(map[uint32][]uint64)
	for _, locationInfo := range locationInfoMapByUniqIdMap {
		sellerRes := locationInfo.GetSellerResult()
		if sellerRes != nil {
			if sellerRes.GetRetcode() == 0 {
				sellerLocationIds := sellerRes.GetLocationIds()
				if len(sellerLocationIds) != 0 {
					availSellerLocationIdsByUniqIdMap[locationInfo.GetUniqueId()] = sellerLocationIds
				}
			} else {
				logging.GetLogger(ctx).Warn(
					fmt.Sprintf("failed to getLocationIdByAddressId for this seller query, uniqId=%s, retCode=%d, message=%s",
						locationInfo.GetUniqueId(), sellerRes.GetRetcode(), sellerRes.GetMessage()))
			}
		}

		buyerRes := locationInfo.GetBuyerResult()
		if buyerRes != nil {
			if buyerRes.GetRetcode() == 0 {
				buyerLocationIds := buyerRes.GetLocationIds()
				if len(buyerLocationIds) != 0 {
					availBuyerLocationIdsByChannelIdMap[locationInfo.GetChannelId()] = buyerLocationIds
				}
			} else {
				logging.GetLogger(ctx).Warn(
					fmt.Sprintf("failed to getLocationIdByAddressId for this buyer query, uniq=%s, retCode=%d, message=%s",
						locationInfo.GetUniqueId(), buyerRes.GetRetcode(), buyerRes.GetMessage()))
			}

		}
	}

	return availSellerLocationIdsByUniqIdMap, availBuyerLocationIdsByChannelIdMap, nil
}

func (c *CalculationFactorsRepoImpl) getPItemLeafCatId(ctx context.Context, pItemId uint64, pItemProductInfoMap map[uint64]*ibsPb.ProductInfo) (uint64, error) {
	pItemLeafCatIdMap := c.itemService.GetItemLeafCatIdMap(ctx, []uint64{pItemId}, pItemProductInfoMap)
	pItemLeafCatIdRes, ok := pItemLeafCatIdMap[pItemId]
	if !ok || pItemLeafCatIdRes == nil || pItemLeafCatIdRes.Err != nil {
		return 0, cerr.New(fmt.Sprintf("failed to get pItemLeafCatId for pItemId=%v", pItemId), uint32(pb.Constant_ERROR_GET_ITEM_INFO))
	}
	pItemLeafCatId := pItemLeafCatIdRes.LeafCatId
	return uint64(pItemLeafCatId), nil
}

func (c *CalculationFactorsRepoImpl) getPItemChannelList(ctx context.Context, pItemId uint64, pItemProductInfoMap map[uint64]*ibsPb.ProductInfo) ([]uint32, error) {
	channelWhiteList, err := c.localSipSystemConfigService.GetChannelWhitelist(ctx)
	if err != nil {
		return nil, err
	}
	whileListChannelMap := make(map[int64]struct{})
	for _, channel := range channelWhiteList {
		whileListChannelMap[channel] = struct{}{}
	}

	pItemEnabledChannelIdsMap := c.itemService.GetItemEnableChannelIdsMap(ctx, []uint64{pItemId}, pItemProductInfoMap)
	if pItemEnabledChannelIdsMap[pItemId] == nil || pItemEnabledChannelIdsMap[pItemId].Err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to get pItemEnabledChannelIds for pItemId=%v", pItemId), uint32(pb.Constant_ERROR_GET_ENABLED_CHANNELS))
	}
	slsChannelList := make([]uint32, 0)
	for _, channel := range pItemEnabledChannelIdsMap[pItemId].EnabledChannelIds {
		if _, ok := whileListChannelMap[int64(channel)]; ok {
			slsChannelList = append(slsChannelList, channel)
		}
	}
	if len(slsChannelList) < 1 {
		logging.GetLogger(ctx).Warn("calcHiddenPriceFromSLS: slsChannelList is empty", ulog.Uint64("primaryItemId", pItemId))
		return nil, cerr.New("slsChannelList is empty", uint32(pb.Constant_ERROR_EMPTY_ENABLED_CHANNELS))
	}

	return slsChannelList, nil
}

func (c *CalculationFactorsRepoImpl) getInitialHiddenPriceForLocalSipFromDb(ctx context.Context, pRegion string, dbQueries []model.LocalSipHiddenPriceQuery) ([]model.LocalSipHiddenPriceResult, error) {
	res := make([]model.LocalSipHiddenPriceResult, len(dbQueries))
	dbSession := c.sipRepo.DbSession()
	for i, query := range dbQueries {
		record, err := c.sipRepo.GetHiddenPriceConfigRecordByWeight(ctx, dbSession, pRegion, query.ARegion, query.Weight)
		if err != nil {
			res[i] = model.LocalSipHiddenPriceResult{
				QueryId: query.QueryId,
				Err:     err,
			}
			continue
		}
		if record == nil {
			res[i] = model.LocalSipHiddenPriceResult{
				Err: cerr.New(fmt.Sprintf("hidden fee record for weight not found|P-region=%s|A-region=%s|weight=%d", pRegion, query.ARegion, query.Weight), uint32(pb.Constant_ERROR_NOT_FOUND)),
			}
			continue
		}
		res[i] = model.LocalSipHiddenPriceResult{
			QueryId:     query.QueryId,
			HiddenPrice: calcutil.ToRealPrice(record.HiddenPrice),
		}
	}

	return res, nil
}
