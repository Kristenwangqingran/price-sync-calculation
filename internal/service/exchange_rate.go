package service

import (
	"context"

	internalExchangeRatePb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_exchange_rate.pb"
)

type ExchangeRateService interface {
	// GetMerchantExchangeRateMap fetch exchange rate map by merchant ids.
	// If failed to get for one merchant id, then only record error and continue to handle remaining merchant ids.
	GetMerchantExchangeRateMap(ctx context.Context, merchantIdList []uint64) map[uint64]*MerchantExchangeRateInfo

	// GetMerchantExchangeRate get merchant exchange rate by one merchant id.
	// If failed, then return error.
	GetMerchantExchangeRate(ctx context.Context, merchantId uint64) (string, map[string]float64, error)

	GetMerchantExchangeRateInfo(ctx context.Context, merchantId uint64) (*internalExchangeRatePb.ExchangeRateInfo, error)
}

type MerchantExchangeRateInfo struct {
	Err                     error
	MerchantCurrency        string
	MerchantExchangeRateMap map[string]float64
}
