package service

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_business.pb"
)

type PriceBusinessService interface {
	GetOPLPrices(ctx context.Context, itemId uint64, region string) ([]*price_business.ItemLevelPrice, uint32, error)
}
