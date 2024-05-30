package service

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
)

type SipItemDataService interface {
	GetPrimaryItemDataBatch(ctx context.Context, primaryShopId uint64, primaryItemIds []uint64) (map[uint64]PrimaryItemData, error)
	GetPrimaryItemModelIdsByAffiItemModelIds(ctx context.Context,
		primaryShopId uint64, affiItemModelIds []model.ItemModelId) (map[model.ItemModelId]model.ItemModelId, error)
}

type PrimaryItemData struct {
	PrimaryItemId uint64

	Weight int64
}

type ItemMappingData struct {
	PrimaryItemId uint64
	AffiItemId    uint64

	ItemMargin     int64
	AffiRealWeight int64
	Ctime          int64
	Mtime          int64
}