package a_shop

import (
	"context"

	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/db"
)

// ShopService will be replaced by ShopCoreService and AShopDataDM
// ShopCoreService gets data from shop.core service, while AShopDataDM gets data from shop_map_tab, which will later be migrated to shop_data_tab under price_sync_db
type AShopDataDM interface {
	GetAShopData(ctx context.Context, aShopId uint64) (*internal.AShopData, error)
	GetAShopDataBatch(ctx context.Context, aShopIds []uint64) (map[uint64]*internal.AShopData, error)
	GetAShopMarginBatch(ctx context.Context, aShopIds []uint64) (map[uint64]int64, error)
	GetAShopPriceRatioBatch(ctx context.Context, affiShopIds []uint64) (map[uint64]int64, error)
	SetAShopDataShopMargin(ctx context.Context, aShopId, pShopId uint64, shopMargin int32) error
	SetAShopDataPromoId(ctx context.Context, aShopId, pShopId uint64, promoId uint64) error
	GetAShopPromoId(ctx context.Context, aShopId uint64) (uint64, error)
}

type aShopDataDMImpl struct {
	aShopDataDB db.AShopDataDB
}

func NewAShopDataDM(aShopDataDB db.AShopDataDB) AShopDataDM {
	return &aShopDataDMImpl{
		aShopDataDB: aShopDataDB,
	}
}

func (dm *aShopDataDMImpl) GetAShopData(ctx context.Context, aShopId uint64) (*internal.AShopData, error) {
	return dm.aShopDataDB.GetByAffiShopId(ctx, aShopId)
}

func (dm *aShopDataDMImpl) GetAShopDataBatch(ctx context.Context, aShopIds []uint64) (map[uint64]*internal.AShopData, error) {
	aShopDataList, err := dm.aShopDataDB.GetByAffiShopIds(ctx, aShopIds)
	if err != nil {
		return nil, err
	}

	aShopDataMap := make(map[uint64]*internal.AShopData)
	for _, aShopData := range aShopDataList {
		aShopDataMap[aShopData.GetAffiShopid()] = aShopData
	}
	return aShopDataMap, nil
}

func (dm *aShopDataDMImpl) GetAShopMarginBatch(ctx context.Context, aShopIds []uint64) (map[uint64]int64, error) {
	shopData, err := dm.aShopDataDB.GetByAffiShopIds(ctx, aShopIds)
	if err != nil {
		return nil, err
	}
	shopMarginMap := make(map[uint64]int64)
	for _, shopDatum := range shopData {
		shopMarginMap[shopDatum.GetAffiShopid()] = shopDatum.GetShopMargin()
	}
	return shopMarginMap, nil
}

func (dm *aShopDataDMImpl) GetAShopPriceRatioBatch(ctx context.Context, aShopIds []uint64) (map[uint64]int64, error) {
	shopData, err := dm.aShopDataDB.GetByAffiShopIds(ctx, aShopIds)
	if err != nil {
		return nil, err
	}
	shopPriceRatioMap := make(map[uint64]int64)
	for _, shopDatum := range shopData {
		shopPriceRatioMap[shopDatum.GetAffiShopid()] = int64(shopDatum.GetPriceRatio())
	}
	return shopPriceRatioMap, nil
}

func (dm *aShopDataDMImpl) GetAShopPromoId(ctx context.Context, aShopId uint64) (uint64, error) {
	shopData, err := dm.aShopDataDB.GetByAffiShopId(ctx, aShopId)
	if err != nil {
		return 0, err
	}
	return shopData.GetPromotionId(), nil
}

func (dm *aShopDataDMImpl) SetAShopDataShopMargin(ctx context.Context, aShopId, pShopId uint64, shopMargin int32) error {
	return dm.aShopDataDB.SetAShopDataShopMargin(ctx, aShopId, pShopId, shopMargin)
}

func (dm *aShopDataDMImpl) SetAShopDataPromoId(ctx context.Context, aShopId, pShopId uint64, promoId uint64) error {
	return dm.aShopDataDB.SetAShopDataPromoId(ctx, aShopId, pShopId, promoId)
}
