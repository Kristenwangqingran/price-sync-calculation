package data

import (
	"context"

	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

type FetchCalcFactorForMtskuAndMpskuDm interface {
	FetchCalcFactorDataForGlobalDiscount(ctx context.Context, queries []*priceSyncPriceCalculationPb.GlobalDiscountQueryId) *CalcFactorDataForMtskuAndMpsku
}
