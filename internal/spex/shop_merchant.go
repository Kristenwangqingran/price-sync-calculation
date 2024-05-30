package spex

import (
	"context"

	sm "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/shop_merchant.pb"
)

const (
	cmdGetShopIDListByMerchant    = "shop.merchant.get_shop_id_list_by_merchant"
	cmdCheckMerchantShopCbsc      = "shop.merchant.check_merchant_shop_cbsc"
	cmdGetMerchant                = "shop.merchant.get_merchant"
	cmdGetMerchantList            = "shop.merchant.get_merchant_list"
	cmdBatchCheckMerchantShopCbsc = "shop.merchant.batch_check_merchant_shop_cbsc"
)

type ShopMerchant interface {
	GetShopIDListByMerchant(ctx context.Context, req *sm.GetShopIdListByMerchantRequest) (*sm.GetShopIdListByMerchantResponse, error)
	CheckMerchantShopCbsc(ctx context.Context, req *sm.CheckMerchantShopCbscRequest) (*sm.CheckMerchantShopCbscResponse, error)
	GetMerchant(ctx context.Context, req *sm.GetMerchantRequest) (*sm.GetMerchantResponse, error)
	GetMerchantList(ctx context.Context, req *sm.GetMerchantListRequest) (*sm.GetMerchantListResponse, error)
	BatchCheckMerchantShopCbsc(ctx context.Context, req *sm.BatchCheckMerchantShopCbscRequest) (*sm.BatchCheckMerchantShopCbscResponse, error)
}

type shopMerchantProxy struct {
}

func NewShopMerchant() ShopMerchant {
	return &shopMerchantProxy{}
}

func (p *shopMerchantProxy) GetShopIDListByMerchant(ctx context.Context, req *sm.GetShopIdListByMerchantRequest) (*sm.GetShopIdListByMerchantResponse, error) {
	resp := &sm.GetShopIdListByMerchantResponse{}
	err := callSPEX(ctx, cmdGetShopIDListByMerchant, req, resp)
	return resp, err
}

func (p *shopMerchantProxy) CheckMerchantShopCbsc(ctx context.Context, req *sm.CheckMerchantShopCbscRequest) (*sm.CheckMerchantShopCbscResponse, error) {
	resp := &sm.CheckMerchantShopCbscResponse{}
	err := callSPEX(ctx, cmdCheckMerchantShopCbsc, req, resp)
	return resp, err
}

func (p *shopMerchantProxy) GetMerchant(ctx context.Context, req *sm.GetMerchantRequest) (*sm.GetMerchantResponse, error) {
	resp := &sm.GetMerchantResponse{}
	err := callSPEX(ctx, cmdGetMerchant, req, resp)
	return resp, err
}

func (p *shopMerchantProxy) GetMerchantList(ctx context.Context, req *sm.GetMerchantListRequest) (*sm.GetMerchantListResponse, error) {
	resp := &sm.GetMerchantListResponse{}
	err := callSPEX(ctx, cmdGetMerchantList, req, resp)
	return resp, err
}

func (p *shopMerchantProxy) BatchCheckMerchantShopCbsc(ctx context.Context, req *sm.BatchCheckMerchantShopCbscRequest) (*sm.BatchCheckMerchantShopCbscResponse, error) {
	resp := &sm.BatchCheckMerchantShopCbscResponse{}
	err := callSPEX(ctx, cmdBatchCheckMerchantShopCbsc, req, resp)
	return resp, err
}
