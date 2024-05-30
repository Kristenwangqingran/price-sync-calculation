package logic

import (
	"context"

	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

type CommonSIPLogic interface {
	GetAShopMarginBatch(ctx context.Context, affiShopIds []uint64) (map[uint64]int64, error)
	GetAShopPriceRatioBatch(ctx context.Context, affiShopIds []uint64) (map[uint64]int64, error)
	GetAItemMarginBatch(ctx context.Context, affiShopIDToItemIdsMap map[uint64][]uint64) (map[uint64]int32, error)
	GetAItemRealWeight(ctx context.Context, affiShopId, affiItemId uint64) (int32, error)
	SetAShopMargin(ctx context.Context, affiShopId uint64, margin int64) error
	GetPShopOpsPriceRatioSettingBatch(ctx context.Context, pShopId []uint64) ([]*pb.PShopOpsPriceRatioSetting, error)
	SetAItemMargin(ctx context.Context, affiShopId, affiItemId uint64, aItemMargin int64) error
	SetAItemRealWeight(ctx context.Context, affiShopId, affiItemId uint64, aItemRealWeight int64) error
	CreateCBSIPAShopSellerDiscountPromotion(ctx context.Context, affiShopId uint64) error
	GetCBSIPAShopSellerDiscountPromotion(ctx context.Context, affiShopId uint64) (uint64, error)
}
