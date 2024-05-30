package spex

import (
	"context"
	"fmt"
	"strings"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	marketplaceLogisticsShopChannelsPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/marketplace_logistics_shop_channels.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
	"git.garena.com/shopee/platform/golang_splib/sps"
)

const (
	cmdGetChannels = "marketplace.logistics.shop_channels.get_channels"
)

type LogisticsShopChannels interface {
	GetChannelsRequest(ctx context.Context, region string) ([]*marketplaceLogisticsShopChannelsPb.Channel, error)
}

type logisticsShopChannelsProxy struct {
}

func NewLogisticsShopChannels() LogisticsShopChannels {
	return &logisticsShopChannelsProxy{}
}

func (p *logisticsShopChannelsProxy) GetChannelsRequest(ctx context.Context, region string) ([]*marketplaceLogisticsShopChannelsPb.Channel, error) {
	req := &marketplaceLogisticsShopChannelsPb.GetChannelsRequest{}
	resp := &marketplaceLogisticsShopChannelsPb.GetChannelsResponse{}

	ctx, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return nil, cerr.Wrap(err, fmt.Sprintf("failed to add region into context, region=%s", region),
			uint32(pb.Constant_ERROR_INTERNAL))
	}

	err = callSPEX(ctx, cmdGetChannels, req, resp, sps.WithRequestParam(strings.ToLower(region)))
	if err != nil {
		return nil, err
	}
	return resp.GetChannelList(), nil
}
