package spex

import (
	"context"

	ibsPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/item_business.pb"
)

const (
	cmdGetProductInfoForDisplay = "item.business.get_product_info_for_display"
	cmdGetProductInfoByItemIds  = "item.business.get_product_info_by_item_ids"
)

type ItemBusiness interface {
	GetProductInfoForDisplay(ctx context.Context, req *ibsPb.GetProductInfoForDisplayRequest) (*ibsPb.GetProductInfoForDisplayResponse, error)
	GetProductInfoByItemIds(ctx context.Context, req *ibsPb.GetProductInfoByItemIdsRequest) (*ibsPb.GetProductInfoByItemIdsResponse, error)
}

type itemBusinessProxy struct {
}

func NewItemBusiness() ItemBusiness {
	return &itemBusinessProxy{}
}

func (p *itemBusinessProxy) GetProductInfoForDisplay(ctx context.Context, req *ibsPb.GetProductInfoForDisplayRequest) (*ibsPb.GetProductInfoForDisplayResponse, error) {
	resp := &ibsPb.GetProductInfoForDisplayResponse{}
	err := callSPEX(ctx, cmdGetProductInfoForDisplay, req, resp)
	return resp, err
}

func (p *itemBusinessProxy) GetProductInfoByItemIds(ctx context.Context, req *ibsPb.GetProductInfoByItemIdsRequest) (*ibsPb.GetProductInfoByItemIdsResponse, error) {
	resp := &ibsPb.GetProductInfoByItemIdsResponse{}
	err := callSPEX(ctx, cmdGetProductInfoByItemIds, req, resp)
	return resp, err
}
