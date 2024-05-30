package shop_ops_audit_log

import (
	"context"
	"fmt"
	"time"

	"git.garena.com/shopee/common/gdbc/gdbc"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/infra/snowflake"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/orm"
)

type ShopOpsAuditLogRepo interface {
	orm.DbSessionFactory
}

func NewAuditLogRepo() *ShopOpsAuditLogRepoImpl {
	return &ShopOpsAuditLogRepoImpl{}
}

type ShopOpsAuditLogRepoImpl struct {
}

func (a *ShopOpsAuditLogRepoImpl) DbSession() orm.DbSession {
	return (*gdbc.DB)(config.GetAuditLogDBClient())
}

func (a *ShopOpsAuditLogRepoImpl) Insert(ctx context.Context, entry *OpsShopLog) error {
	t := time.Now()
	currTime := t.Unix()
	entry.Ctime = currTime
	entry.Mtime = currTime
	if entry.Id == 0 {
		nextId, err := snowflake.GetGenIDWorker().NextId(ctx)
		if err != nil {
			return cerr.Wrap(err, "error generating next shop_ops_audit_log id", uint32(price_sync_price_calculation.Constant_ERROR_INTERNAL))
		}
		entry.Id = nextId
	}
	_, err := a.DbSession().Create(entry).Do(ctx)
	if err != nil {
		return cerr.Wrap(err, fmt.Sprintf("error creating shop_ops_audit_log, log=%v", entry), uint32(price_sync_price_calculation.Constant_ERROR_INTERNAL))
	}
	return nil
}
