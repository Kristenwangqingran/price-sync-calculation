package spex

import (
	"context"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/shop_core.pb"
)

const (
	cmdBatchGetShopRegions        = "shop.core.batch_get_shop_regions"
	cmdGetShop                    = "shop.core.get_shop"
	cmdGetAShopInfo               = "shop.core.get_ashop_info"
	cmdGetAShopIdsByPShopId       = "shop.core.get_ashop_ids_by_pshop_id"
	cmdGetPShopIdByAShopId        = "shop.core.get_pshop_id_by_ashop_id"
	cmdBatchGetPShopIdsByAShopIds = "shop.core.batch_get_pshop_ids_by_ashop_ids"
	cmdIsSellerWarehouseShop      = "shop.core.is_seller_warehouse_shop"
	cmdGetShopWarehouseByShopId   = "shop.core.get_shop_warehouse_by_shop_id"
	cmdGetShopUseridList          = "shop.core.get_shop_userid_list"

	batchGetShopRegionsSizeLimit = 30 // from shop.core IDL comments
)

type ShopCore interface {
	BatchGetShopRegions(ctx context.Context, req *shop_core.BatchGetShopRegionsRequest) (*shop_core.BatchGetShopRegionsResponse, error)
	GetShop(ctx context.Context, req *shop_core.GetShopRequest) (*shop_core.GetShopResponseV2, error)
	GetAShopInfo(ctx context.Context, aShopId int64) (*shop_core.GetAshopInfoResponse, error)
	GetAShopIdsByPShopId(ctx context.Context, pShopId uint64) ([]int64, error)
	GetPShopIdByAShopId(ctx context.Context, aShopId uint64) (int64, error)
	GetPShopIdsByAShopIdsBatch(ctx context.Context, aShopIds []uint64) ([]*shop_core.PARelation, error)
	IsSellerWarehouseShop(ctx context.Context, req *shop_core.IsSellerWarehouseShopRequest) (*shop_core.IsSellerWarehouseShopResponse, error)
	GetShopWarehouseByShopId(ctx context.Context, req *shop_core.GetShopWarehouseByShopIdRequest) (*shop_core.GetShopWarehouseByShopIdResponse, error)
	GetUserIdByShopIdBatch(ctx context.Context, req *shop_core.GetShopUseridListRequest) (*shop_core.GetShopUseridListResponse, error)
}

type shopCoreProxy struct {
}

func NewShopCore() ShopCore {
	return &shopCoreProxy{}
}

func (p *shopCoreProxy) BatchGetShopRegions(ctx context.Context,
	req *shop_core.BatchGetShopRegionsRequest) (*shop_core.BatchGetShopRegionsResponse, error) {
	resp := &shop_core.BatchGetShopRegionsResponse{
		ShopRegionPairs: make([]*shop_core.ShopRegionPair, 0),
	}
	if len(req.GetShopIdList()) == 0 {
		return resp, nil
	}

	for start := 0; start < len(req.ShopIdList); start += batchGetShopRegionsSizeLimit {
		end := start + batchGetShopRegionsSizeLimit
		if end > len(req.ShopIdList) {
			end = len(req.ShopIdList)
		}

		subReq := &shop_core.BatchGetShopRegionsRequest{
			ShopIdList: req.ShopIdList[start:end],
		}

		subResp := &shop_core.BatchGetShopRegionsResponse{}

		if err := callSPEX(ctx, cmdBatchGetShopRegions, subReq, subResp); err != nil {
			return nil, err
		}

		resp.ShopRegionPairs = append(resp.ShopRegionPairs, subResp.ShopRegionPairs...)
	}

	if len(resp.ShopRegionPairs) != len(req.ShopIdList) {
		return nil, cerr.New("shop.core.batch_get_shop_regions resp number not match with req",
			uint32(pb.Constant_ERROR_EXTERNAL))
	}

	return resp, nil
}

func (p *shopCoreProxy) GetShop(ctx context.Context, req *shop_core.GetShopRequest) (*shop_core.GetShopResponseV2, error) {
	resp := &shop_core.GetShopResponseV2{}

	if err := callSPEX(ctx, cmdGetShop, req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (p *shopCoreProxy) GetAShopInfo(ctx context.Context, aShopId int64) (*shop_core.GetAshopInfoResponse, error) {
	req := &shop_core.GetAshopInfoRequest{
		AshopId: proto.Int64(aShopId),
	}
	resp := &shop_core.GetAshopInfoResponse{}

	err := callSPEX(ctx, cmdGetAShopInfo, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *shopCoreProxy) GetAShopIdsByPShopId(ctx context.Context, pShopId uint64) ([]int64, error) {
	req := &shop_core.GetAllAshopIdsByPshopIdRequest{
		PshopId: proto.Int64(int64(pShopId)),
	}
	resp := &shop_core.GetAshopIdsByPshopIdResponse{}
	err := callSPEX(ctx, cmdGetAShopIdsByPShopId, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetAshopIds(), nil
}

func (p *shopCoreProxy) GetPShopIdByAShopId(ctx context.Context, aShopId uint64) (int64, error) {
	req := &shop_core.GetPshopIdByAshopIdRequest{
		AshopId: proto.Int64(int64(aShopId)),
	}
	resp := &shop_core.GetPshopIdByAshopIdResponse{}
	err := callSPEX(ctx, cmdGetPShopIdByAShopId, req, resp)
	if err != nil {
		return 0, err
	}
	return resp.GetPshopId(), nil
}

func (p *shopCoreProxy) GetPShopIdsByAShopIdsBatch(ctx context.Context, aShopIds []uint64) ([]*shop_core.PARelation, error) {
	convertedShopIds := make([]int64, 0, len(aShopIds))
	for _, aShopId := range aShopIds {
		convertedShopIds = append(convertedShopIds, int64(aShopId))
	}
	req := &shop_core.BatchGetPshopIdsByAshopIdsRequest{
		AshopIds: convertedShopIds,
	}
	resp := &shop_core.BatchGetPshopIdsByAshopIdsResponse{}
	err := callSPEX(ctx, cmdBatchGetPShopIdsByAShopIds, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetPaRelations(), nil
}

func (p *shopCoreProxy) IsSellerWarehouseShop(ctx context.Context, req *shop_core.IsSellerWarehouseShopRequest) (*shop_core.IsSellerWarehouseShopResponse, error) {
	resp := &shop_core.IsSellerWarehouseShopResponse{}
	err := callSPEX(ctx, cmdIsSellerWarehouseShop, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *shopCoreProxy) GetShopWarehouseByShopId(ctx context.Context, req *shop_core.GetShopWarehouseByShopIdRequest) (*shop_core.GetShopWarehouseByShopIdResponse, error) {
	resp := &shop_core.GetShopWarehouseByShopIdResponse{}
	err := callSPEX(ctx, cmdGetShopWarehouseByShopId, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *shopCoreProxy) GetUserIdByShopIdBatch(ctx context.Context,
	req *shop_core.GetShopUseridListRequest) (*shop_core.GetShopUseridListResponse, error) {
	resp := &shop_core.GetShopUseridListResponse{}
	err := callSPEX(ctx, cmdGetShopUseridList, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
