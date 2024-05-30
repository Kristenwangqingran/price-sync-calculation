package db

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
)

type MSTShopDB interface {
	GetByPShopId(ctx context.Context, pShopId uint64) (*sip_db.MstShop, error)
	GetByShopIdBatch(ctx context.Context, pShopIds []uint64) ([]*sip_db.MstShop, error)
}

type mstShopDBImpl struct {
	sipRepo sip_db.SipRepo
}

func NewMSTShopDB(sipRepo sip_db.SipRepo) MSTShopDB {
	return &mstShopDBImpl{
		sipRepo: sipRepo,
	}
}

func (db *mstShopDBImpl) GetByPShopId(ctx context.Context, pShopId uint64) (*sip_db.MstShop, error) {
	session := db.sipRepo.DbSession()
	return db.sipRepo.GetMstShopRecordByShopId(ctx, session, pShopId)
}

func (db *mstShopDBImpl) GetByShopIdBatch(ctx context.Context, pShopIds []uint64) ([]*sip_db.MstShop, error) {
	session := db.sipRepo.DbSession()
	return db.sipRepo.GetMstShopRecordByShopIdBatch(ctx, session, pShopIds)
}
