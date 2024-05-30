package db

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
)

type LocalShippingFeeConfigDB interface {
	GetLocalShippingFeeConfigRecordByWeight(ctx context.Context, mstRegion, affiRegion string, weight int64) (*sip_db.LocalShippingFeeConfigRecord, error)
}

type localShippingFeeConfigDB struct {
	sipRepo sip_db.SipRepo
}

func NewLocalShippingFeeConfigDB(sipRepo sip_db.SipRepo) LocalShippingFeeConfigDB {
	return &localShippingFeeConfigDB{
		sipRepo: sipRepo,
	}
}

func (db *localShippingFeeConfigDB) GetLocalShippingFeeConfigRecordByWeight(ctx context.Context, mstRegion, affiRegion string, weight int64) (*sip_db.LocalShippingFeeConfigRecord, error) {
	session := db.sipRepo.DbSession()
	return db.sipRepo.GetLocalShippingFeeConfigRecordByWeight(ctx, session, mstRegion, affiRegion, weight)
}
