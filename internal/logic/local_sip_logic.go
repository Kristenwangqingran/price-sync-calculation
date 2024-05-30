package logic

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

type LocalSipLogic interface {
	GetLocalSipPriceFactors(ctx context.Context, infoType model.LocalSipPriceFactorInfoType, queries []model.GetLocalSipPriceFactorQuery) ([]model.LocalSipPriceFactorInfo, error)
	CalculateAPriceByPItemForLocalSip(ctx context.Context, pShopId uint64, pItemId uint64, pRegion string, queries []model.LocalSipCalculateAPriceQuery, calculateForCreate bool) ([]model.LocalSipCalculateAPriceResult, error)
	CalculateAItemOPL(ctx context.Context, pRegion string, pItemId uint64, aShopId uint64, aRegion string) (*pb.CustomizedOPL, error)
}
