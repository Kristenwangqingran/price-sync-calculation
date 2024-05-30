package data

import (
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	internalFulfillmentChannelPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_fulfillment_channel.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

// getUniqueMerchantIdListFromGlobalDiscountQueryIds get merchant id map and list from the request,
// already remove duplicate element after handled.
func getUniqueMerchantIdListFromGlobalDiscountQueryIds(queries []*priceSyncPriceCalculationPb.GlobalDiscountQueryId) []uint64 {
	merchantIdMap := make(map[uint64]bool)
	for _, query := range queries {
		_, exist := merchantIdMap[query.GetMerchantId()]
		if !exist {
			merchantIdMap[query.GetMerchantId()] = true
		}
	}

	merchantIdList := make([]uint64, 0, len(merchantIdMap))
	for m := range merchantIdMap {
		merchantIdList = append(merchantIdList, m)
	}

	return merchantIdList
}

// getUniqueMpskuShopIdListFromGlobalDiscountQueryIds get mpsku shop id list from the request,
// already remove duplicate element after handled.
func getUniqueMpskuShopIdListFromGlobalDiscountQueryIds(queries []*priceSyncPriceCalculationPb.GlobalDiscountQueryId) []*model.ShopIdRegion {
	mpskuQueryIdMap := make(map[uint64]*priceSyncPriceCalculationPb.GlobalDiscountQueryId)
	for _, query := range queries {
		_, exist := mpskuQueryIdMap[query.GetMpskuShopId()]
		if !exist {
			mpskuQueryIdMap[query.GetMpskuShopId()] = query
		}
	}

	mpskuShopIdRegionList := make([]*model.ShopIdRegion, 0, len(mpskuQueryIdMap))
	for _, q := range mpskuQueryIdMap {
		mpskuShopIdRegionList = append(mpskuShopIdRegionList, &model.ShopIdRegion{
			ShopId: q.GetMpskuShopId(),
			Region: q.GetMpskuRegion(),
		})
	}

	return mpskuShopIdRegionList
}

// getUniqueMpskuItemIdListFromGlobalDiscountQueryIds get mpsku item id list from the request,
// already remove duplicate element after handled.
func getUniqueMpskuItemIdListFromGlobalDiscountQueryIds(queries []*priceSyncPriceCalculationPb.GlobalDiscountQueryId) ([]uint64, map[string][]uint64) {
	mpskuItemIdMap := make(map[uint64]bool)
	mpskuRegionItemIdMap := make(map[string][]uint64)
	for _, query := range queries {
		_, exist := mpskuItemIdMap[query.GetMpskuItemId()]
		if !exist {
			mpskuItemIdMap[query.GetMpskuItemId()] = true

			if mpskuRegionItemIdMap[query.GetMpskuRegion()] == nil {
				mpskuRegionItemIdMap[query.GetMpskuRegion()] = make([]uint64, 0)
			}
			mpskuRegionItemIdMap[query.GetMpskuRegion()] = append(mpskuRegionItemIdMap[query.GetMpskuRegion()], query.GetMpskuItemId())
		}
	}

	mpskuItemIdList := make([]uint64, 0, len(mpskuItemIdMap))
	for m := range mpskuItemIdMap {
		mpskuItemIdList = append(mpskuItemIdList, m)
	}

	return mpskuItemIdList, mpskuRegionItemIdMap
}

func GetItemShopEnabledChannelList(shopChannelDetailMap map[uint32]*internalFulfillmentChannelPb.FulfillmentChannelDetail, itemEnabledChannelIDList []uint32) []uint32 {
	if len(shopChannelDetailMap) == 0 || len(itemEnabledChannelIDList) == 0 {
		return nil
	}

	var enabledChannelIDList []uint32

	for _, chanID := range itemEnabledChannelIDList {
		shopChan, ok := shopChannelDetailMap[chanID]
		if !ok || !shopChan.GetEnabled() {
			continue
		}
		enabledChannelIDList = append(enabledChannelIDList, chanID)
	}

	return enabledChannelIDList
}
