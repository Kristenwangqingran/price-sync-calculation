package calculate

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/data"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

type CalculateMtskuAndMpskuDm interface {
	CalcGlobalDiscount(ctx context.Context, queries []*priceSyncPriceCalculationPb.GlobalDiscountQueryId, calcFactorData *data.CalcFactorDataForMtskuAndMpsku) ([]*priceSyncPriceCalculationPb.GlobalDiscountInfo, error)
}

type QueryDataForMtskuAndMpsku struct {
	MerchantId         uint64
	MpskuShopId        uint64
	MpskuItemId        uint64
	MpskuModelId       uint64
	MpskuRegion        string
	MtskuOriginalPrice int64
	QueryPriceData     int64 // if QueryMtskuToMpsku = true, store mtsku price, else store mpsku price
	QueryMtskuToMpsku  bool
}
