package db

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_v2_db"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
)

type MskuDB interface {
	GetByAffiMskuIds(ctx context.Context, mstShopId uint64, affiMskuIds []model.ItemModelId) ([]*sip_v2_db.MskuMapRecord, error)
}

type mskuDB struct {
	sipV2Repo sip_v2_db.SipV2Repo
}

func NewMskuDB(sipV2Repo sip_v2_db.SipV2Repo) MskuDB {
	return &mskuDB{
		sipV2Repo: sipV2Repo,
	}
}

func (dm *mskuDB) GetByAffiMskuIds(ctx context.Context, mstShopId uint64, affiMskuIds []model.ItemModelId) ([]*sip_v2_db.MskuMapRecord, error) {
	session := dm.sipV2Repo.DbSession()
	return dm.sipV2Repo.GetMtskuMapByAffiMskuIds(ctx, session, mstShopId, affiMskuIds)
}
