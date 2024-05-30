package data

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	ibsPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/item_business.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
)

type FetchItemShopEnabledChannelIdsQuery struct {
	ItemId uint64
	ShopId uint64
	Region string
}

func FetchItemShopEnabledChannelIds(ctx context.Context, queries []*FetchItemShopEnabledChannelIdsQuery, itemInfoMap map[uint64]*ibsPb.ProductInfo, itemService service.ItemService, logisticService service.LogisticService) map[uint64]*service.EnabledChannelIdsInfo {
	itemShopEnabledChannelIdsMap := make(map[uint64]*service.EnabledChannelIdsInfo)

	itemIdList := make([]uint64, len(queries))
	for idx, q := range queries {
		itemIdList[idx] = q.ItemId
	}
	itemEnabledChannelIdsInfoMap := itemService.GetItemEnableChannelIdsMap(ctx, itemIdList, itemInfoMap)

	for _, query := range queries {
		shopChannelDetailMap, err := logisticService.GetShopChannelDetailMap(ctx, query.ShopId, query.Region)
		if err != nil {
			itemShopEnabledChannelIdsMap[query.ItemId] = &service.EnabledChannelIdsInfo{
				Err: cerr.New(fmt.Sprintf(
					"failed to get shop level channel detail, shopId=%d, region=%s, itemId=%d",
					query.ShopId, query.Region, query.ItemId),
					uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_ENABLED_CHANNELS)),
			}
		}

		itemEnabledChannelIdsInfo := itemEnabledChannelIdsInfoMap[query.ItemId]
		if itemEnabledChannelIdsInfo == nil || itemEnabledChannelIdsInfo.Err != nil {
			errMsg := fmt.Sprintf("failed to get item enabled channel ids, itemId=%d", query.ItemId)
			if itemEnabledChannelIdsInfo.Err != nil {
				errMsg = fmt.Sprintf("%s, err=%s", errMsg, itemEnabledChannelIdsInfo.Err.Error())
			}

			itemShopEnabledChannelIdsMap[query.ItemId] = &service.EnabledChannelIdsInfo{
				Err: cerr.New(errMsg,
					uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_ENABLED_CHANNELS)),
			}

			continue
		}

		itemShopEnabledChannelList := GetItemShopEnabledChannelList(shopChannelDetailMap, itemEnabledChannelIdsInfo.EnabledChannelIds)
		if len(itemShopEnabledChannelList) == 0 {
			itemShopEnabledChannelIdsMap[query.ItemId] = &service.EnabledChannelIdsInfo{
				Err: cerr.New(fmt.Sprintf("no available channel, itemId=%d", query.ItemId),
					uint32(priceSyncPriceCalculationPb.Constant_ERROR_EMPTY_ENABLED_CHANNELS)),
			}

			continue
		}

		itemShopEnabledChannelIdsMap[query.ItemId] = &service.EnabledChannelIdsInfo{
			EnabledChannelIds: itemShopEnabledChannelList,
		}
	}

	return itemShopEnabledChannelIdsMap
}
