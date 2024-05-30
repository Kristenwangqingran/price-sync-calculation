package db

import (
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/shop_ops_audit_log"
)

type ShopOpsAuditLogDB interface {
}

type shopOpsAuditLogDBImpl struct {
	repo shop_ops_audit_log.ShopOpsAuditLogRepo
}

func NewShopOpsAuditLogDB(repo shop_ops_audit_log.ShopOpsAuditLogRepo) ShopOpsAuditLogDB {
	return &shopOpsAuditLogDBImpl{
		repo: repo,
	}
}

func (s *shopOpsAuditLogDBImpl) InsertShopOpsAuditLog() {

}
