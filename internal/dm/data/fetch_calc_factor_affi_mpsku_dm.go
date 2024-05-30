package data

import (
	"context"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

type FetchCalcFactorForAffiMpskuDm interface {
	FetchCalcFactorDataForLocalSipOverseaDiscount(ctx context.Context, req *pb.CalcLocalSipOverseaDiscountPriceRequest) (*CalcFactorDataForAffiMpsku, error)
}
