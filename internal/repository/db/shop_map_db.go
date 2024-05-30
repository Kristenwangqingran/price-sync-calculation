package db

import (
	"context"

	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
)

type AShopDataDB interface {
	GetByAffiShopId(ctx context.Context, affiShopId uint64) (*internal.AShopData, error)
	GetByAffiShopIds(ctx context.Context, affiShopIds []uint64) ([]*internal.AShopData, error)
	SetAShopDataShopMargin(ctx context.Context, aShopId, pShopId uint64, shopMargin int32) error
	SetAShopDataPromoId(ctx context.Context, aShopId, pShopId uint64, promoId uint64) error
}

type aShopDataDBImpl struct {
	sipRepo sip_db.SipRepo
}

func NewShopMapDB(sipRepo sip_db.SipRepo) AShopDataDB {
	return &aShopDataDBImpl{
		sipRepo: sipRepo,
	}
}

func (db *aShopDataDBImpl) GetByAffiShopIds(ctx context.Context, affiShopIds []uint64) ([]*internal.AShopData, error) {
	session := db.sipRepo.DbSession()
	return db.sipRepo.GetAShopDataByAffiShopIdBatch(ctx, session, affiShopIds)
}

func (db *aShopDataDBImpl) GetByAffiShopId(ctx context.Context, affiShopId uint64) (*internal.AShopData, error) {
	session := db.sipRepo.DbSession()
	return db.sipRepo.GetAShopDataByAffiShopId(ctx, session, affiShopId)
}

func (db *aShopDataDBImpl) SetAShopDataShopMargin(ctx context.Context, aShopId, pShopId uint64, shopMargin int32) error {
	session := db.sipRepo.DbSession()
	return db.sipRepo.SetAShopDataShopMargin(ctx, session, aShopId, pShopId, shopMargin)
}

func (db *aShopDataDBImpl) SetAShopDataPromoId(ctx context.Context, aShopId, pShopId uint64, promoId uint64) error {
	session := db.sipRepo.DbSession()
	return db.sipRepo.SetAShopDataPromoId(ctx, session, aShopId, pShopId, promoId)
}
