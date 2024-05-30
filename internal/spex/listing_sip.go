package spex

import (
	"context"
	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	listing_sip "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/marketplace_listing_upload_crossupload_api.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
)

const (
	cmdGetAItemByPItemIds = "marketplace.listing.upload.crossupload.api.get_aitem_by_pitem_ids"
	cmdGetPItemByAItemIds = "marketplace.listing.upload.crossupload.api.get_pitem_by_aitem_ids"
)

type ListingSIP interface {
	GetAItemByPItemIds(ctx context.Context, pShopId, pItemId uint64, aShopIds []uint64) (*listing_sip.GetAitemByPitemIDsResponse, error)
	GetPItemByAItemIds(ctx context.Context, aShopId uint64, aItemIds []uint64) (*listing_sip.GetPitemByAitemIDsResponse, error)
}
type listingSIPProxy struct {}

func NewListingSIPProxy() ListingSIP {
    return &listingSIPProxy{}
}

func (l *listingSIPProxy) GetAItemByPItemIds(ctx context.Context, pShopId, pItemId uint64, aShopIds []uint64) (*listing_sip.GetAitemByPitemIDsResponse, error) {
	ctx, _ = cidutil.FillCtxWithNewCID(ctx, cidutil.GlobalCID)
	req := &listing_sip.GetAitemByPitemIDsRequest{
		PshopId: proto.Uint64(pShopId),
		PitemId: proto.Uint64(pItemId),
		AshopIds: aShopIds,
	}
	resp := &listing_sip.GetAitemByPitemIDsResponse{}
	err := callSPEX(ctx, cmdGetAItemByPItemIds, req, resp)
	if err != nil {
		return nil, cerr.New(err.Error(), uint32(pb.Constant_ERROR_EXTERNAL))
	}
	return resp, nil
}

func (l *listingSIPProxy) GetPItemByAItemIds(ctx context.Context, aShopId uint64, aItemIds []uint64) (*listing_sip.GetPitemByAitemIDsResponse, error) {
	ctx, _ = cidutil.FillCtxWithNewCID(ctx, cidutil.GlobalCID)
	req := &listing_sip.GetPitemByAitemIDsRequest{
		AshopId: proto.Uint64(aShopId),
		AitemIds: aItemIds,
	}
	resp := &listing_sip.GetPitemByAitemIDsResponse{}
	err := callSPEX(ctx, cmdGetPItemByAItemIds, req, resp)
	if err != nil {
		return nil, cerr.New(err.Error(), uint32(pb.Constant_ERROR_EXTERNAL))
	}
	return resp, nil
}
