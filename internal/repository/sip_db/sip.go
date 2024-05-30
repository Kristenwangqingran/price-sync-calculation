package sip_db

import (
	"context"

	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/orm"
)

type SipRepo interface {
	orm.DbSessionFactory
	GetHiddenPriceConfigRecordByWeight(ctx context.Context, session orm.DbSession, mstRegion, affiRegion string, weight int64) (*LocalHiddenPriceConfigRecord, error)
	GetLocalShippingFeeConfigRecordByWeight(ctx context.Context, session orm.DbSession, mstRegion, affiRegion string, weight int64) (*LocalShippingFeeConfigRecord, error)
	GetSystemConfigRecordByType(ctx context.Context, session orm.DbSession, configType int) (*SystemConfigRecord, error)
	GetHiddenPriceConfigList(ctx context.Context, session orm.DbSession, pRegion, aRegion string) ([]*LocalHiddenPriceConfigRecord, error)
	GetShippingFeeConfigList(ctx context.Context, session orm.DbSession, pRegion, aRegion string) ([]*LocalShippingFeeConfigRecord, error)
	GetExchangeRateByCurrency(ctx context.Context, session orm.DbSession, currencyPair string) (*ExchangeRate, error)
	GetHpfnConfigByHpfnKey(ctx context.Context, session orm.DbSession, hpfnKey string) (*HpfnConfig, error)
	GetAllHpfnConfig(ctx context.Context, session orm.DbSession) ([]*HpfnConfig, error)
	GetAllExchangeRate(ctx context.Context, session orm.DbSession) ([]*ExchangeRate, error)

	GetAShopDataByAffiShopId(ctx context.Context, session orm.DbSession, affiShopId uint64) (*internal.AShopData, error)
	GetAShopDataByAffiShopIdBatch(ctx context.Context, session orm.DbSession, affiShopIds []uint64) ([]*internal.AShopData, error)

	GetMstShopRecordByShopId(ctx context.Context, session orm.DbSession, pShopId uint64) (*MstShop, error)
	GetMstShopRecordByShopIdBatch(ctx context.Context, session orm.DbSession, pShopIds []uint64) ([]*MstShop, error)

	GetShopMapWithoutOffboardByAShopIdsAndPShopId(ctx context.Context, session orm.DbSession, pShopId uint64, aShopIds []uint64) ([]*internal.AShopData, error)

	GetAllEditItemPriceAllowList(ctx context.Context, session orm.DbSession) ([]*EditItemPriceAllowList, error)

	//TODO: remove pShopId after DB split is done
	SetAShopDataShopMargin(ctx context.Context, session orm.DbSession, aShopId, pShopId uint64, shopMargin int32) error
	SetAShopDataPromoId(ctx context.Context, session orm.DbSession, aShopId, pShopId uint64, promoId uint64) error
}
