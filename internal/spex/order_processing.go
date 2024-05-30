package spex

import (
	"context"
	"strings"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/marketplace_order_processing_cb_collection_api.pb"
	"git.garena.com/shopee/platform/golang_splib/sps"
)

const (
	cmdGetDummyBuyerId = "marketplace.order_processing.cb_collection.api.get_dummy_buyer_id"
)

type OrderProcessing interface {
	GetDummyBuyerId(ctx context.Context, req *marketplace_order_processing_cb_collection_api.GetDummyBuyerIdRequest, region string) (*marketplace_order_processing_cb_collection_api.GetDummyBuyerIdResponse, error)
}

type orderProcessingProxy struct {
}

func GetOrderProcessing() OrderProcessing {
	return &orderProcessingProxy{}
}

func (p *orderProcessingProxy) GetDummyBuyerId(ctx context.Context,
	req *marketplace_order_processing_cb_collection_api.GetDummyBuyerIdRequest, region string) (*marketplace_order_processing_cb_collection_api.GetDummyBuyerIdResponse, error) {
	resp := &marketplace_order_processing_cb_collection_api.GetDummyBuyerIdResponse{}

	if err := callSPEX(ctx, cmdGetDummyBuyerId, req, resp, sps.WithRequestParam(strings.ToLower(region))); err != nil {
		return nil, err
	}

	return resp, nil
}
