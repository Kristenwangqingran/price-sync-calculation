package db

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
)

type SystemConfigDB interface {
	GetSystemConfigRecordByType(ctx context.Context, configType int) (*sip_db.SystemConfigRecord, error)
}

type systemConfigDB struct {
	sipRepo sip_db.SipRepo
}

func NewSystemConfigDB(sipRepo sip_db.SipRepo) SystemConfigDB {
	return &systemConfigDB{
		sipRepo: sipRepo,
	}
}

func (db *systemConfigDB) GetSystemConfigRecordByType(ctx context.Context, configType int) (*sip_db.SystemConfigRecord, error) {
	session := db.sipRepo.DbSession()
	return db.sipRepo.GetSystemConfigRecordByType(ctx, session, configType)
}
