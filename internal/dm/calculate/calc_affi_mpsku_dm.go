package calculate

import (
	"context"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/data"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

type CalculateAffiMpskuDm interface {
	CalcLocalSipOverseaDiscount(ctx context.Context, req *pb.CalcLocalSipOverseaDiscountPriceRequest, calcData *data.CalcFactorDataForAffiMpsku) ([]*pb.LocalSIPAffiPriceResult, error)
}
