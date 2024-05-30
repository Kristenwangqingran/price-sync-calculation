package spex

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/promotion_item_discount.pb"
)

const (
	cmdGetSellerDiscountList = "promotion.item_discount.get_seller_discount_list"
	cmdSetSellerDiscount     = "promotion.item_discount.set_seller_discount"
)

type ItemDiscount interface {
	GetSellerDiscountList(ctx context.Context, req *promotion_item_discount.GetSellerDiscountListRequest) (*promotion_item_discount.GetSellerDiscountListResponse, error)
	SetSellerDiscount(ctx context.Context, req *promotion_item_discount.SetSellerDiscountRequest) (*promotion_item_discount.SetSellerDiscountResponse, error)
}

type itemDiscountProxy struct {
}

func NewItemDiscount() ItemDiscount {
	return &itemDiscountProxy{}
}

func (i *itemDiscountProxy) GetSellerDiscountList(ctx context.Context, req *promotion_item_discount.GetSellerDiscountListRequest) (*promotion_item_discount.GetSellerDiscountListResponse, error) {
	resp := &promotion_item_discount.GetSellerDiscountListResponse{}
	err := callSPEX(ctx, cmdGetSellerDiscountList, req, resp)
	return resp, err
}

func (i *itemDiscountProxy) SetSellerDiscount(ctx context.Context, req *promotion_item_discount.SetSellerDiscountRequest) (*promotion_item_discount.SetSellerDiscountResponse, error) {
	resp := &promotion_item_discount.SetSellerDiscountResponse{}
	err := callSPEX(ctx, cmdSetSellerDiscount, req, resp)
	return resp, err
}
