package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	internalFulfillmentChannelPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_fulfillment_channel.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"

	ibsPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/item_business.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
)

type ItemServiceDm struct {
	ItemBusinessSpexService spex.ItemBusiness
}

func NewItemService(ibsSpexService spex.ItemBusiness) ItemService {
	return &ItemServiceDm{
		ItemBusinessSpexService: ibsSpexService,
	}
}

func (dm *ItemServiceDm) GetProductInfoMapForMixedRegion(ctx context.Context, regionItemIdsMap map[string][]uint64) map[uint64]*ibsPb.ProductInfo {
	itemInfoMap := make(map[uint64]*ibsPb.ProductInfo)

	for region, itemIds := range regionItemIdsMap {
		regionItemInfoMap := dm.GetProductInfoMapForSameRegion(ctx, itemIds, region)

		for itemId, itemInfo := range regionItemInfoMap {
			if itemInfo != nil {
				itemInfoMap[itemId] = itemInfo
			}
		}
	}

	return itemInfoMap
}

func (dm *ItemServiceDm) GetProductInfoMapForSameRegion(ctx context.Context, itemIds []uint64, region string) map[uint64]*ibsPb.ProductInfo {
	if len(itemIds) == 0 {
		return nil
	}

	itemIdsReq := make([]*ibsPb.ItemId, len(itemIds))
	for idx, itemId := range itemIds {
		itemIdsReq[idx] = &ibsPb.ItemId{
			ItemId: proto.Uint64(itemId),
		}
	}

	batchSize := config.GetBatchConfig().MaxBatchSizeForIBSGetProductInfoByItemIds
	productInfoMap := make(map[uint64]*ibsPb.ProductInfo)
	for start := 0; start < len(itemIds); {
		end := start + int(batchSize)
		if end > len(itemIds) {
			end = len(itemIds)
		}

		req := &ibsPb.GetProductInfoByItemIdsRequest{
			ItemIds: itemIdsReq[start:end],
			InfoTypes: []uint32{
				uint32(ibsPb.Constant_CATEGORY),
				uint32(ibsPb.Constant_LOGISTICS),
			},
			NeedDeleted: proto.Bool(true),
		}

		ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, region)
		resp, err := dm.ItemBusinessSpexService.GetProductInfoByItemIds(ctxWithCID, req)
		if err != nil {
			logging.GetLogger(ctx).Error(fmt.Sprintf("failed to GetProductInfoByItemIds, err=%s", err.Error()))
		}

		for _, info := range resp.GetInfo() {
			productInfoMap[info.GetItemId()] = info
		}

		start += int(batchSize)
	}

	return productInfoMap
}

func (dm *ItemServiceDm) GetItemLeafCatIdMap(ctx context.Context, itemIdList []uint64, itemInfoMap map[uint64]*ibsPb.ProductInfo) map[uint64]*ItemLeafCategoryInfo {
	itemLeafCatIdInfoMap := make(map[uint64]*ItemLeafCategoryInfo)

	for _, itemId := range itemIdList {
		itemInfo := itemInfoMap[itemId]
		itemCats := itemInfo.GetCat().GetGlobalCat().GetCatIds()

		if len(itemCats) == 0 {
			errMsg := fmt.Sprintf("failed to get category, itemId=%d", itemId)
			logging.GetLogger(ctx).Error(errMsg)

			itemLeafCatIdInfoMap[itemId] = &ItemLeafCategoryInfo{
				Err: cerr.New(errMsg,
					uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_ITEM_INFO)),
			}

			continue
		}

		itemLeafCatIdInfoMap[itemId] = &ItemLeafCategoryInfo{
			LeafCatId: itemCats[len(itemCats)-1],
		}
	}

	return itemLeafCatIdInfoMap
}

func (dm *ItemServiceDm) GetItemWeightMap(ctx context.Context, itemIdList []uint64, itemInfoMap map[uint64]*ibsPb.ProductInfo) map[uint64]*ItemWeightInfo {
	itemWeightInfoMap := make(map[uint64]*ItemWeightInfo)

	for _, itemId := range itemIdList {
		itemInfo := itemInfoMap[itemId]
		itemWeight := itemInfo.GetLogistics().GetWeight()

		if itemInfo.GetLogistics() == nil || itemInfo.GetLogistics().Weight == nil {
			errMsg := fmt.Sprintf("failed to get item weight, itemId=%d", itemId)
			logging.GetLogger(ctx).Error(errMsg)

			itemWeightInfoMap[itemId] = &ItemWeightInfo{
				Err: cerr.New(errMsg,
					uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_ITEM_INFO)),
			}

			continue
		}

		itemWeightInfoMap[itemId] = &ItemWeightInfo{
			ItemWeight: itemWeight,
		}
	}

	return itemWeightInfoMap
}

func (dm *ItemServiceDm) GetItemEnableChannelIdsMap(ctx context.Context, itemIdList []uint64, itemInfoMap map[uint64]*ibsPb.ProductInfo) map[uint64]*EnabledChannelIdsInfo {
	itemEnabledChannelIdsMap := make(map[uint64]*EnabledChannelIdsInfo)

	for _, itemId := range itemIdList {
		itemInfo := itemInfoMap[itemId]
		if itemInfo == nil {

			errMsg := fmt.Sprintf("failed to get item info, itemId=%d", itemId)
			logging.GetLogger(ctx).Error(errMsg)

			itemEnabledChannelIdsMap[itemId] = &EnabledChannelIdsInfo{
				Err: cerr.New(errMsg,
					uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_ITEM_INFO)),
			}

			continue
		}

		itemEnabledChannelList, err := GetItemEnabledChannelList(itemInfo.GetLogistics(), ctx)
		if err != nil {
			logging.GetLogger(ctx).Error(err.Error())

			itemEnabledChannelIdsMap[itemId] = &EnabledChannelIdsInfo{
				Err: cerr.New(err.Error(),
					uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_ITEM_INFO)),
			}

			continue
		}

		itemEnabledChannelIdsMap[itemInfo.GetItemId()] = &EnabledChannelIdsInfo{
			EnabledChannelIds: itemEnabledChannelList,
		}
	}

	return itemEnabledChannelIdsMap
}

func GetItemEnabledChannelList(logistics *ibsPb.Logistics, ctx context.Context) ([]uint32, error) {
	if logistics == nil || len(logistics.GetLogisticsInfo()) == 0 {
		return nil, nil
	}

	var enabledChannelIDList []uint32
	itemLogisticsInfoMap := make(map[string]*internalFulfillmentChannelPb.ItemLogisticsInfo)
	err := json.Unmarshal(logistics.GetLogisticsInfo(), &itemLogisticsInfoMap)
	if err != nil {
		errMsg := fmt.Sprintf("failed to unmarshal item logistics info, err=%s, logistics=%s",
			err, logistics.GetLogisticsInfo())
		logging.GetLogger(ctx).Error(errMsg)
		return nil, cerr.New(errMsg, uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	for chanIDStr, logisticsInfo := range itemLogisticsInfoMap {
		if !logisticsInfo.GetEnabled() {
			continue
		}
		chanID, err := strconv.ParseUint(chanIDStr, 10, 64)
		if err != nil {
			logging.GetLogger(ctx).Error(fmt.Sprintf("failed to parse channel id string=%s, and will ignore", chanIDStr))
			continue
		}
		enabledChannelIDList = append(enabledChannelIDList, uint32(chanID))
	}
	return enabledChannelIDList, nil
}
