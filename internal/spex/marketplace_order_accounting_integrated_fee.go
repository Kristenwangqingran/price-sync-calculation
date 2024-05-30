package spex

import (
	"context"
	"fmt"
	"strings"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	moaif "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/marketplace_order_accounting_integrated_fee.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
)

const (
	cmdGetAppliableRulesByShop = "marketplace.order_accounting.integrated_fee.get_appliable_rules_by_shop"
)

type MarketplaceOrderAccountingIntegratedFee interface {
	GetAppliableRulesByShop(ctx context.Context, req *moaif.GetAppliableRulesByShopRequest) (*moaif.GetAppliableRulesByShopResponse, error)
}

type marketplaceOrderAccountingIntegratedFeeProxy struct {
}

func NewMarketplaceOrderAccountingIntegratedFee() MarketplaceOrderAccountingIntegratedFee {
	return &marketplaceOrderAccountingIntegratedFeeProxy{}
}

func (p *marketplaceOrderAccountingIntegratedFeeProxy) GetAppliableRulesByShop(ctx context.Context, req *moaif.GetAppliableRulesByShopRequest) (*moaif.GetAppliableRulesByShopResponse, error) {
	cid := cidutil.FetchCIDFromCtx(ctx)
	if config.IsRegionDeprecated(strings.ToUpper(cid)) {
		return nil, cerr.New(fmt.Sprintf("unexpected cid=%s", cid), uint32(pb.Constant_ERROR_PARAMS))
	}

	resp := &moaif.GetAppliableRulesByShopResponse{}
	err := callSPEX(ctx, cmdGetAppliableRulesByShop, req, resp)
	return resp, err
}
