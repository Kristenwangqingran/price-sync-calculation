package service

import (
	"context"

	ibsPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/item_business.pb"
)

// ItemService acts as a abstraction layer over item Spex Service
type ItemService interface {
	// GetProductInfoMapForMixedRegion get product map by item ids for multiple regions.
	// Since we need fill correct cid to call item side api.
	GetProductInfoMapForMixedRegion(ctx context.Context, regionItemIdsMap map[string][]uint64) map[uint64]*ibsPb.ProductInfo

	// GetProductInfoMapForSameRegion get product map by item ids under same region.
	// If happened error, then will log error, and continue to handle remaining part.
	GetProductInfoMapForSameRegion(ctx context.Context, itemIds []uint64, region string) map[uint64]*ibsPb.ProductInfo

	// GetItemLeafCatIdMap get leaf category id map.
	// If failed to get for one, then record error in the response and continue to handle remaining part.
	GetItemLeafCatIdMap(ctx context.Context, itemIdList []uint64, itemInfoMap map[uint64]*ibsPb.ProductInfo) map[uint64]*ItemLeafCategoryInfo

	// GetItemWeightMap get item weight map.
	// If failed to get for one, then record error in the response and continue to handle remaining part.
	GetItemWeightMap(ctx context.Context, itemIdList []uint64, itemInfoMap map[uint64]*ibsPb.ProductInfo) map[uint64]*ItemWeightInfo

	// GetItemEnableChannelIdsMap get enabled ids map on item level.
	// If failed to get for one, then record error in the response and continue to handle remaining part.
	GetItemEnableChannelIdsMap(ctx context.Context, itemIdList []uint64, itemInfoMap map[uint64]*ibsPb.ProductInfo) map[uint64]*EnabledChannelIdsInfo
}

type ItemWeightInfo struct {
	Err        error
	ItemWeight uint64
}

type ItemLeafCategoryInfo struct {
	Err       error
	LeafCatId uint32
}

type EnabledChannelIdsInfo struct {
	Err               error
	EnabledChannelIds []uint32
}
