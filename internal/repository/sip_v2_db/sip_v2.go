package sip_v2_db

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/orm"
)

type SipV2Repo interface {
	orm.DbSessionFactory
	GetMstItemRecordBatch(ctx context.Context, session orm.DbSession, primaryShopId uint64, primaryItemIds []uint64) ([]*MstItemRecord, error)
	GetAItemData(ctx context.Context, session orm.DbSession, primaryShopId uint64, affiItemId uint64) ([]*internal.AItemData, error)
	GetAItemDataBatch(ctx context.Context, session orm.DbSession, primaryShopId uint64, affiItemIds []uint64) ([]*internal.AItemData, error)
	GetMtskuMapByAffiMskuIds(ctx context.Context, session orm.DbSession, mstShopId uint64, affiMskuIds []model.ItemModelId) ([]*MskuMapRecord, error)
	UpdateAItemMargin(ctx context.Context, session orm.DbSession, primaryShopId uint64, affiItemId uint64, aItemMargin int32) error
	UpdateAItemRealWeight(ctx context.Context, session orm.DbSession, primaryShopId uint64, affiItemId uint64, aItemRealWeight int32) error
}
