package service

import (
	"context"

	shopMerchantPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/shop_merchant.pb"
)

// ShopMerchantService acts as a abstraction layer over shop merchant Spex Service
type ShopMerchantService interface {
	// GetMerchantRegionInfoMap get merchant region map.
	// If failed to get for one merchant id, then record error in the response and continue to handle remaining part.
	GetMerchantRegionInfoMap(ctx context.Context, merchantIdList []uint64) map[uint64]*MerchantRegionInfo

	// GetMerchantInfoMap get merchant info map.
	// If failed to get for one merchant id, then only record error log and continue to handle remaining part.
	GetMerchantInfoMap(ctx context.Context, merchantIdList []int64) map[int64]*shopMerchantPb.Merchant

	GetMerchantRegion(ctx context.Context, merchantId uint64) (string, error)

	GetMerchantInfoByShopId(ctx context.Context, shopId uint64) (*shopMerchantPb.Merchant, error)

	CheckMerchantShopCbsc(ctx context.Context, shopId uint64) (bool, error)
}

type MerchantRegionInfo struct {
	Err            error
	MerchantRegion string
}
