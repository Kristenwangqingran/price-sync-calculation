package a_item

import (
	"context"
	"fmt"

	"git.garena.com/shopee/common/ulog"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/db"
)

// SipItemDataService will be replaced by AItemDataDM in future
type AItemDataDM interface {
	GetAItemDataBatch(ctx context.Context, primaryShopId uint64, affiItemIds []uint64) (map[uint64]*internal.AItemData, error)
	GetAItemData(ctx context.Context, primaryShopId uint64, affiItemId uint64) (*internal.AItemData, error)
	GetAItemMarginBatch(ctx context.Context, primaryShopId uint64, affiItemIds []uint64) (map[uint64]int32, error)
	GetAItemRealWeight(ctx context.Context, primaryShopId, affiItemId uint64) (int32, error)
	SetAItemMargin(ctx context.Context, primaryShopId uint64, affiItemId uint64, aItemMargin int32) error
	SetAItemRealWeight(ctx context.Context, primaryShopId uint64, affiItemId uint64, aItemRealWeight int32) error
}

type aItemDataDMImpl struct {
	aItemDataDB db.AItemDataDB
}

func NewAItemDataDM(aItemDataDB db.AItemDataDB) AItemDataDM {
	return &aItemDataDMImpl{
		aItemDataDB: aItemDataDB,
	}
}

func (dm *aItemDataDMImpl) GetAItemDataBatch(ctx context.Context, primaryShopId uint64, affiItemIds []uint64) (map[uint64]*internal.AItemData, error) {
	records, err := dm.aItemDataDB.GetAItemDataBatch(ctx, primaryShopId, affiItemIds)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		ulog.DefaultLoggerFromContext(ctx).Warn("no a_item data retrieved from DB", ulog.Reflect("affi_item_ids", affiItemIds))
		return nil, nil
	}

	// Known dirty data issue that one affi_item_id can have multiple records in item_map_tab
	// at this stage, we always return the one with latest ctime.
	recordsMap := make(map[uint64]*internal.AItemData)
	for _, record := range records {
		currLatest, ok := recordsMap[record.GetAffiItemid()]
		if !ok || record.GetCtime() > currLatest.GetCtime() {
			recordsMap[record.GetAffiItemid()] = record
		}
	}

	return recordsMap, nil
}

func (dm *aItemDataDMImpl) GetAItemData(ctx context.Context, primaryShopId uint64, affiItemId uint64) (*internal.AItemData, error) {
	records, err := dm.aItemDataDB.GetAItemData(ctx, primaryShopId, affiItemId)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		ulog.DefaultLoggerFromContext(ctx).Warn("no a_item data retrieved from DB", ulog.Reflect("affi_item_id", affiItemId))
		return nil, nil
	}
	return getRecordWithLatestCTime(records), nil
}

func (dm *aItemDataDMImpl) GetAItemMarginBatch(ctx context.Context, primaryShopId uint64, affiItemIds []uint64) (map[uint64]int32, error) {
	recordsMap, err := dm.GetAItemDataBatch(ctx, primaryShopId, affiItemIds)
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]int32)
	for affiItemId, aItemData := range recordsMap {
		result[affiItemId] = aItemData.GetItemMargin()
	}
	return result, nil
}

func (dm *aItemDataDMImpl) GetAItemRealWeight(ctx context.Context, primaryShopId, affiItemId uint64) (int32, error) {
	record, err := dm.GetAItemData(ctx, primaryShopId, affiItemId)
	if err != nil {
		return 0, err
	}
	if record == nil {
		return 0, fmt.Errorf("no A item data was returned from DB, aItemId=%d", affiItemId)
	}

	return record.GetAffiRealWeight(), nil
}

func (dm *aItemDataDMImpl) SetAItemMargin(ctx context.Context, primaryShopId uint64, affiItemId uint64, aItemMargin int32) error {
	err := dm.aItemDataDB.SetAItemMargin(ctx, primaryShopId, affiItemId, aItemMargin)
	if err != nil {
		return err
	}
	return nil
}

func (dm *aItemDataDMImpl) SetAItemRealWeight(ctx context.Context, primaryShopId uint64, affiItemId uint64, aItemRealWeight int32) error {
	err := dm.aItemDataDB.SetAItemRealWeight(ctx, primaryShopId, affiItemId, aItemRealWeight)
	if err != nil {
		return err
	}
	return nil
}

func getRecordWithLatestCTime(aItemDataList []*internal.AItemData) *internal.AItemData {
	if len(aItemDataList) == 0 {
		return nil
	}
	currLatest := aItemDataList[0]
	for _, aItemData := range aItemDataList {
		if aItemData.GetCtime() > currLatest.GetCtime() {
			currLatest = aItemData
		}
	}
	return currLatest
}
