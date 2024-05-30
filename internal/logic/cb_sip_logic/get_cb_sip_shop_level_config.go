package cb_sip_logic

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
)

func (c *CbSipLogicImpl) GetCbSipShopLevelConfig(ctx context.Context, req model.CbSipGetShopLevelConfigRequest) (model.CbSipGetShopLevelConfigResult, error) {
	res, err := c.factorsRepo.GetShopSipRateConfigByPShopIdForCbSip(ctx, req.PShopId)
	if err != nil {
		return model.CbSipGetShopLevelConfigResult{}, err
	}
	return model.CbSipGetShopLevelConfigResult{
		AShopConfigList: res,
	}, nil
}
