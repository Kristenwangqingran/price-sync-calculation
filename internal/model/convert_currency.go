package model

import (
	"fmt"

	"git.garena.com/shopee/core-server/core-logic/cutil"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

type ExchangeRateSource = uint32

const (
	ExchangeRateSourceSellerPlatform    = ExchangeRateSource(pb.Constant_SELLER_PLATFORM)
	ExchangeRateSourceCbSipExchangeRate = ExchangeRateSource(pb.Constant_CB_SIP_EXCHANGE_RATE)
	ExchangeRateSourceOrderMart         = ExchangeRateSource(pb.Constant_ORDER_MART_EXCHANGE_RATE)
)

type ConvertCurrencyRequest struct {
	SrcPriceList       []int64
	ExchangeRateSource ExchangeRateSource

	// for source cb sip && order mart
	SrcCurrency string
	DstCurrency string

	// for source seller platform
	MerchantId  uint64
	MpskuRegion string
}

type ConvertCurrencyResult struct {
	ExchangeRate float64
	DstPrices    []int64
}

type OrderMartExchangeRate struct {
	Currency     string  `json:"currency"`
	ExchangeRate float64 `json:"exchange_rate"`
	Region       string  `json:"grass_region"`
	Date         string  `json:"grass_date"`
}

func (r *OrderMartExchangeRate) Validate() error {
	if r == nil {
		return fmt.Errorf("row is nil")
	}

	if r.ExchangeRate == 0 || r.Currency == "" || r.Region == "" {
		return fmt.Errorf("invalid row data|row=%v", cutil.LazyJSONEncoder(r))
	}

	return nil
}
