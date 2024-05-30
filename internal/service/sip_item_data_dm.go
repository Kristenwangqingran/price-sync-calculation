package service

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/db"
)

type sipItemDataService struct {
	itemMapDB db.ItemMapDB
	mskuDB    db.MskuDB
}

func NewSipItemDataService(itemMapDB db.ItemMapDB, mskuDB db.MskuDB) SipItemDataService {
	return &sipItemDataService{
		itemMapDB: itemMapDB,
		mskuDB:    mskuDB,
	}
}

func (dm *sipItemDataService) GetPrimaryItemDataBatch(ctx context.Context, primaryShopId uint64, primaryItemIds []uint64) (map[uint64]PrimaryItemData, error) {
	records, err := dm.itemMapDB.GetMstItemRecordBatch(ctx, primaryShopId, primaryItemIds)
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]PrimaryItemData)
	for _, record := range records {
		result[record.ItemId] = PrimaryItemData{
			PrimaryItemId: record.ItemId,
			Weight:        record.Weight,
		}
	}
	return result, nil
}

func (dm *sipItemDataService) GetPrimaryItemModelIdsByAffiItemModelIds(ctx context.Context,
	primaryShopId uint64, affiItemModelIds []model.ItemModelId) (map[model.ItemModelId]model.ItemModelId, error) {

	records, err := dm.mskuDB.GetByAffiMskuIds(ctx, primaryShopId, affiItemModelIds)
	if err != nil {
		return nil, err
	}

	result := make(map[model.ItemModelId]model.ItemModelId)
	for _, record := range records {
		affiMskuId := model.ItemModelId{
			ItemId:  record.AffiItemId,
			ModelId: record.AffiModelId,
		}

		primaryMskuId := model.ItemModelId{
			ItemId:  record.MstItemId,
			ModelId: record.MstModelId,
		}

		result[affiMskuId] = primaryMskuId
	}

	return result, nil
}

