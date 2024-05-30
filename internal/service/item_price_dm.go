package service

import (
	"context"
	"fmt"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_basic.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_common_definition.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
	"github.com/golang/protobuf/proto"
)

type itemPriceDM struct {
	priceBasicService spex.PriceBasic
}

func NewItemPriceService(priceBasicService spex.PriceBasic) ItemPriceService {
	return &itemPriceDM{
		priceBasicService: priceBasicService,
	}
}

func (dm *itemPriceDM) GetOriginPriceBatch(ctx context.Context,
	region string, itemModelIds []model.ItemModelId) (map[model.ItemModelId]int64, error) {
	itemIds := make([]uint64, 0)
	for _, itemModelId := range itemModelIds {
		itemIds = append(itemIds, itemModelId.ItemId)
	}

	ctx, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return nil, cerr.Wrap(err, fmt.Sprintf("add region(%s) to context failed", region), uint32(pb.Constant_ERROR_INTERNAL))
	}

	req := &price_basic.GetItemPriceBatchRequest{
		IncludeFuture: proto.Bool(false),
		ItemIdList:    itemIds,
	}

	resp, err := dm.priceBasicService.GetItemPriceBatch(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make(map[model.ItemModelId]int64)
	for _, itemPrice := range resp.ItemModelPrices {
		for _, modelPrice := range itemPrice.ModelPrices {
			itemModelId := model.ItemModelId{
				ItemId:  itemPrice.GetItemId(),
				ModelId: modelPrice.GetModelId(),
			}
			for _, price := range modelPrice.Prices {
				if price.GetPromotionType() == uint32(price_common_definition.Constant_PROMOTION_TYPE_NORMAL) &&
					price.GetPromotionId() == 0 {
					result[itemModelId] = price.GetPrice()
				}
			}
		}
	}
	return result, nil
}
