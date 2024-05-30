package spex

import (
	"context"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	lupb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/marketplace_listing_upload_product.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
)

const (
	cmdGetProductInfo = "marketplace.listing.upload.product.get_product_info"
)

type ListingUploadService interface {
	GetProductInfo(ctx context.Context, region string, shopId, itemId uint64) (*lupb.GetProductInfoResponse, error)
}

type ListingUploadServiceImpl struct {
}

func NewListinguploadServiceImpl() ListingUploadService {
	return &ListingUploadServiceImpl{}
}

func (l *ListingUploadServiceImpl) GetProductInfo(ctx context.Context, region string, shopId, itemId uint64) (*lupb.GetProductInfoResponse, error) {
	ctx, _ = cidutil.FillCtxWithNewCID(ctx, region)
	req := &lupb.GetProductInfoRequest{
		ShopId:          proto.Uint64(shopId),
		Region:          proto.String(region),
		MpskuItemIdList: []uint64{itemId},
		FieldList:       []uint32{uint32(lupb.Constant_PRICE), uint32(lupb.Constant_SIP_INFO)},
	}
	resp := &lupb.GetProductInfoResponse{}
	err := callSPEX(ctx, cmdGetProductInfo, req, resp)
	if err != nil {
		return nil, cerr.New(err.Error(), uint32(pb.Constant_ERROR_EXTERNAL))
	}
	return resp, nil
}
