package logic

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
)

type CbSipLogic interface {
	CalculateSipItemPriceForCbSip(ctx context.Context, req model.CbSipCalculateSipItemPriceRequest) ([]model.CbSipCalculateSipItemPriceResult, error)
	CalculateAPriceByPItemForCbSip(ctx context.Context, request model.CbSipCalculateAPriceByPItemRequest) ([]model.CbSipCalculateAPriceByPItemResult, error)
	CalculateAItemOPL(ctx context.Context, request *model.CbSipCalculateAOPLByPItemRequest) (*model.CbSipCalculateAOPLByPItemResult, error)
	GetCbSipAHiddenFeeConfig(ctx context.Context, req model.CbSipGetAHiddenPriceConfigRequest) (*model.CbSipGetAHiddenPriceConfigResult, error)
	GetCbSipRateConfig(ctx context.Context, infoType uint32) (model.CbSipRateConfigResult, error)
	GetCbSipShopLevelConfig(ctx context.Context, req model.CbSipGetShopLevelConfigRequest) (model.CbSipGetShopLevelConfigResult, error)
	GetCbSipRegionLevelConfig(ctx context.Context, req model.CbSipGetRegionLevelConfigRequest) (model.CbSipGetRegionLevelConfigResult, error)
}
