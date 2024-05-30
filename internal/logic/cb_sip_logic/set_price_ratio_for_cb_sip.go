package cb_sip_logic

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

// TODO
func (c *CbSipLogicImpl) SetPriceRatio(ctx context.Context, request *pb.SetPriceRatioRequest) (*pb.SetPriceRatioResponse, error) {
	primaryShopRegion, err := c.shopCoreService.GetShopRegionByShopId(ctx, request.GetPShopId())
	if err != nil {
		return nil, err
	}

	primaryShopDetail, err := c.shopCoreService.GetShopDetail(ctx, request.GetPShopId(), primaryShopRegion)
	if err != nil {
		return nil, err
	}

	if !primaryShopDetail.IsSipCb {
		return nil, cerr.Wrap(fmt.Errorf("set_price_ratio only allowed when P shop is CB, given shop is not CB"), "", uint32(pb.Constant_ERROR_PARAMS))
	}

	if request.GetIsCreate() {
		//affiRegionList := make([]string, 0, len(request.GetAShopPriceRatioSettings()))
		//defaultPriceRatioList, err := c.systemConfigDM.GetDefaultCBSIPPriceRatioBatch(ctx, primaryShopRegion, affiRegionList)
		//if err != nil {
		//	return nil, err
		//}
		//for _, priceRatioSetting := range
	}
	return nil, nil
}
