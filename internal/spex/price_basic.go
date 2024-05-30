package spex

import (
	"context"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_basic.pb"
)

const (
	cmdGetItemPriceBatch = "price.basic.get_item_price_batch"

	getItemPriceBatchLimit = 500
)

type PriceBasic interface {
	GetItemPriceBatch(ctx context.Context, req *price_basic.GetItemPriceBatchRequest) (*price_basic.GetItemPriceBatchResponse, error)
}

type priceBasicProxy struct {
}

func GetPriceBasic() PriceBasic {
	return &priceBasicProxy{}
}

func (p *priceBasicProxy) GetItemPriceBatch(ctx context.Context, req *price_basic.GetItemPriceBatchRequest) (*price_basic.GetItemPriceBatchResponse, error) {
	resp := &price_basic.GetItemPriceBatchResponse{
		ItemModelPrices: make([]*price_basic.ItemPrices, 0),
	}

	for start := 0; start < len(req.ItemIdList); start += getItemPriceBatchLimit {
		end := start + getItemPriceBatchLimit
		if end > len(req.ItemIdList) {
			end = len(req.ItemIdList)
		}

		subReq := &price_basic.GetItemPriceBatchRequest{
			ItemIdList: req.ItemIdList[start:end],
		}

		subResp := &price_basic.GetItemPriceBatchResponse{}

		if err := callSPEX(ctx, cmdGetItemPriceBatch, subReq, subResp); err != nil {
			return nil, err
		}

		resp.ItemModelPrices = append(resp.ItemModelPrices, subResp.ItemModelPrices...)
	}

	return resp, nil
}
