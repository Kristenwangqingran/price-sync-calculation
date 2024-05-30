package account_service

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
)

type AccountServiceRepo interface {
	GetAllShopList(ctx context.Context, accountId *uint64, merchantId uint64) ([]*model.CNSCShop, error)
	GetUserStatusMap(ctx context.Context, shopIdRegions []model.ShopIdRegion) (map[uint64]int32, error)
}
