package service

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/clog"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_business.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_common_definition.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"

	"github.com/golang/protobuf/proto"
)

type PriceBusinessServiceDm struct {
	priceBusinessSpexService spex.PriceBusiness
}

func NewPriceBusinessService(spex spex.PriceBusiness) *PriceBusinessServiceDm {
	return &PriceBusinessServiceDm{
		priceBusinessSpexService: spex,
	}
}

func (dm *PriceBusinessServiceDm) GetOPLPrices(ctx context.Context, itemId uint64, region string) ([]*price_business.ItemLevelPrice, uint32, error) {
	req := &price_business.GetPurchaseInfoForDisplayRequest{
		ItemModelPurchaseQuantity: []*price_business.ItemModelPurchaseQuantity{
			{
				ItemId: proto.Uint64(itemId),
			},
		},
		InfoTypes: []uint32{uint32(price_business.Constant_PRICE)},
	}
	ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return nil, 0, cerr.New(fmt.Sprintf("failed to fill CID, err=%s", err.Error()),
			uint32(pb.Constant_ERROR_INTERNAL))
	}

	resp, err := dm.priceBusinessSpexService.GetPurchaseInfoForDisplay(ctxWithCID, req, region)
	if len(resp.GetInfo()) == 0 {
		clog.Infof(ctx, "GetPurchaseInfo returned empty info list, itemId=%d", itemId)
		return nil, 0, nil
	}

	// info list length should be at most one, with item_id equals to requested item_id, the check is just for extra safety
	if len(resp.GetInfo()) != 1 {
		err := fmt.Errorf("GetPurchaseInfo returned info list length expected to be 1 but got %d, itemId=%d", len(resp.GetInfo()), itemId)
		clog.Errorf(ctx, err.Error())
		return nil, 0, cerr.Wrap(err, "", uint32(pb.Constant_ERROR_EXTERNAL))
	}

	if resp.GetInfo()[0].GetItemId() != itemId {
		err := fmt.Errorf("GetPurchaseInfo returned info of wrong itemId, returned=%d, expected=%d", resp.GetInfo()[0].GetItemId(), itemId)
		clog.Errorf(ctx, err.Error())
		return nil, 0, cerr.Wrap(err, "", uint32(pb.Constant_ERROR_EXTERNAL))
	}

	info := resp.GetInfo()[0]

	oplPriceList := make([]*price_business.ItemLevelPrice, 0)
	if info.GetItemPriceInfo() == nil {
		clog.Infof(ctx, "GetPurchaseInfo returned info does not contain ItemPriceInfo, itemId=%d", itemId)
		return nil, 0, nil
	}

	itemPriceInfo := info.GetItemPriceInfo()
	if itemPriceInfo.GetPurchaseLimit() == 0 {
		return nil, 0, nil
	}

	for _, price := range append(itemPriceInfo.GetOngoingPrices(), itemPriceInfo.GetFuturePrices()...) {
		if price.GetRuleType() == uint32(price_common_definition.Constant_RULE_TYPE_OVERALL_PURCHASE_LIMIT) {
			oplPriceList = append(oplPriceList, price)
		}
	}
	clog.Infof(ctx, "GetPurchaseInfo contains OPL price of length=%d, itemId=%d", len(oplPriceList), itemId)
	return oplPriceList, itemPriceInfo.GetPurchaseLimit(), nil
}
