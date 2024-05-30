package db

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
)

type LocalHiddenPriceConfigDB interface {
	// the weight unit is gram
	GetHiddenPriceConfigRecordByWeight(ctx context.Context, mstRegion, affiRegion string, weight int64) (*sip_db.LocalHiddenPriceConfigRecord, error)
}

type localHiddenPriceConfigDB struct {
	sipRepo sip_db.SipRepo
}

func NewLocalHiddenPriceConfigDB(sipRepo sip_db.SipRepo) LocalHiddenPriceConfigDB {
	return &localHiddenPriceConfigDB{
		sipRepo: sipRepo,
	}
}

// the weight unit is gram
func (db *localHiddenPriceConfigDB) GetHiddenPriceConfigRecordByWeight(ctx context.Context, mstRegion, affiRegion string, weight int64) (*sip_db.LocalHiddenPriceConfigRecord, error) {
	session := db.sipRepo.DbSession()
	return db.sipRepo.GetHiddenPriceConfigRecordByWeight(ctx, session, mstRegion, affiRegion, weight)
}
