package service

import (
	"context"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
)

type ItemPriceService interface {
	GetOriginPriceBatch(ctx context.Context, region string, itemModelIds []model.ItemModelId) (map[model.ItemModelId]int64, error)
}
