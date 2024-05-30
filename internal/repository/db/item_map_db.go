package db

import (
	"context"

	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_v2_db"
)

// TODO: change name
type ItemMapDB interface {
	GetMstItemRecordBatch(ctx context.Context, primaryShopId uint64, primaryItemIds []uint64) ([]*sip_v2_db.MstItemRecord, error)
}

type itemMapDB struct {
	sipV2Repo sip_v2_db.SipV2Repo
}

func NewItemMapDB(sipV2Repo sip_v2_db.SipV2Repo) ItemMapDB {
	return &itemMapDB{
		sipV2Repo: sipV2Repo,
	}
}

func (db *itemMapDB) GetMstItemRecordBatch(ctx context.Context, primaryShopId uint64, primaryItemIds []uint64) ([]*sip_v2_db.MstItemRecord, error) {
	session := db.sipV2Repo.DbSession()
	return db.sipV2Repo.GetMstItemRecordBatch(ctx, session, primaryShopId, primaryItemIds)
}

type AItemDataDB interface {
	//TODO: remove primaryShopId after moved to new DB
	GetAItemDataBatch(ctx context.Context, primaryShopId uint64, affiItemIds []uint64) ([]*internal.AItemData, error)
	//TODO: return single result for GetAItemData after fixing dirty data in DB
	GetAItemData(ctx context.Context, primaryShopId uint64, affiItemId uint64) ([]*internal.AItemData, error)
	SetAItemMargin(ctx context.Context, primaryShopId uint64, affiItemId uint64, aItemMargin int32) error
	SetAItemRealWeight(ctx context.Context, primaryShopId uint64, affiItemId uint64, aItemRealWeight int32) error
}

type aItemDBImpl struct {
	sipV2Repo sip_v2_db.SipV2Repo
}

func NewAItemDB(sipV2Repo sip_v2_db.SipV2Repo) AItemDataDB {
	return &aItemDBImpl{sipV2Repo: sipV2Repo}
}

func (db *aItemDBImpl) GetAItemDataBatch(ctx context.Context, primaryShopId uint64, affiItemIds []uint64) ([]*internal.AItemData, error) {
	session := db.sipV2Repo.DbSession()
	return db.sipV2Repo.GetAItemDataBatch(ctx, session, primaryShopId, affiItemIds)
}

func (db *aItemDBImpl) GetAItemData(ctx context.Context, primaryShopId uint64, affiItemId uint64) ([]*internal.AItemData, error) {
	session := db.sipV2Repo.DbSession()
	return db.sipV2Repo.GetAItemData(ctx, session, primaryShopId, affiItemId)
}

func (db *aItemDBImpl) SetAItemMargin(ctx context.Context, primaryShopId uint64, affiItemId uint64, aItemMargin int32) error {
	session := db.sipV2Repo.DbSession()
	return db.sipV2Repo.UpdateAItemMargin(ctx, session, primaryShopId, affiItemId, aItemMargin)
}

func (db *aItemDBImpl) SetAItemRealWeight(ctx context.Context, primaryShopId uint64, affiItemId uint64, aItemRealWeight int32) error {
	session := db.sipV2Repo.DbSession()
	return db.sipV2Repo.UpdateAItemRealWeight(ctx, session, primaryShopId, affiItemId, aItemRealWeight)
}
