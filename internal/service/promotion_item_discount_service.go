package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/promotion_item_discount.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/reqidutil"
)

type ItemDiscountService interface {
	GetShopSellerDiscountByShopIdPromoId(ctx context.Context, shopId uint64, promotionId uint64, region string) (*promotion_item_discount.SellerDiscountInfo, error)
	AddShopSellerDiscount(ctx context.Context, shopId, userId uint64, region, title string, startTime, endTime int64) (uint64, error)
	IsSellerDiscountPromotionValid(sellerDiscount *promotion_item_discount.SellerDiscountInfo) bool
}

type itemDiscountServiceImpl struct {
	spexProxy spex.ItemDiscount
}

func NewItemDiscountService(spexProxy spex.ItemDiscount) ItemDiscountService {
	return &itemDiscountServiceImpl{
		spexProxy: spexProxy,
	}
}

func (i *itemDiscountServiceImpl) GetShopSellerDiscountByShopIdPromoId(ctx context.Context, shopId uint64, promotionId uint64, region string) (*promotion_item_discount.SellerDiscountInfo, error) {
	ctx, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return nil,
			cerr.Wrap(err, fmt.Sprintf("add region(%s) to context failed in GetShopSellerDiscountByShopIdPromoId", region),
				uint32(pb.Constant_ERROR_INTERNAL))
	}
	req := &promotion_item_discount.GetSellerDiscountListRequest{
		RequestId:    proto.String(reqidutil.GetOrNewRequestId(ctx)),
		ShopId:       proto.Uint64(shopId),
		Region:       proto.String(region),
		PromotionIds: []uint64{promotionId},
		Limit:        proto.Int32(1),
	}

	resp, err := i.spexProxy.GetSellerDiscountList(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.GetSellerDiscountList()) > 0 {
		return resp.GetSellerDiscountList()[0], nil
	}
	return nil, nil
}

func (i *itemDiscountServiceImpl) AddShopSellerDiscount(ctx context.Context, shopId, userId uint64, region, title string, startTime, endTime int64) (uint64, error) {
	ctx, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return 0,
			cerr.Wrap(err, fmt.Sprintf("add region(%s) to context failed in GetShopSellerDiscountByShopIdPromoId", region),
				uint32(pb.Constant_ERROR_INTERNAL))
	}
	req := &promotion_item_discount.SetSellerDiscountRequest{
		RequestId:   proto.String(reqidutil.GetOrNewRequestId(ctx)),
		PromotionId: proto.Uint64(0),
		ShopId:      proto.Uint64(shopId),
		UserId:      proto.Uint64(userId),
		Country:     proto.String(region),
		Title:       proto.String(title),
		Status:      proto.Uint32(uint32(promotion_item_discount.Constant_SELLER_PROMOTION_STATUS_NORMAL)),
		StartTime:   proto.Uint32(uint32(startTime)),
		EndTime:     proto.Uint32(uint32(endTime)),
	}
	resp, err := i.spexProxy.SetSellerDiscount(ctx, req)
	if err != nil {
		return 0, err
	}
	return resp.GetPromotionId(), nil
}

func (i *itemDiscountServiceImpl) IsSellerDiscountPromotionValid(sellerDiscount *promotion_item_discount.SellerDiscountInfo) bool {
	return sellerDiscount != nil && sellerDiscount.GetStatus() == 1 && sellerDiscount.GetEndTime() > uint32(time.Now().Unix())
}
