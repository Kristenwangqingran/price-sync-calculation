package spex

import (
	"context"
	"strings"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_business.pb"
	"git.garena.com/shopee/platform/golang_splib/sps"
)

const (
	cmdGetPurchaseInfoForDisplay = "price.business.get_purchase_info_for_display"
)

type PriceBusiness interface {
	GetPurchaseInfoForDisplay(ctx context.Context, req *price_business.GetPurchaseInfoForDisplayRequest, region string) (*price_business.GetPurchaseInfoForDisplayResponse, error)
}

type priceBusinessProxy struct {
}

func NewPriceBusiness() PriceBusiness {
	return &priceBusinessProxy{}
}

func (p *priceBusinessProxy) GetPurchaseInfoForDisplay(ctx context.Context, req *price_business.GetPurchaseInfoForDisplayRequest, region string) (*price_business.GetPurchaseInfoForDisplayResponse, error) {
	resp := &price_business.GetPurchaseInfoForDisplayResponse{}
	if err := callSPEX(ctx, cmdGetPurchaseInfoForDisplay, req, resp, sps.WithRequestParam(strings.ToLower(region))); err != nil {
		return nil, err
	}
	return resp, nil
}
