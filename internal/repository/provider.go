package repository

import (
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/account_service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/edit_item_price_allow_list"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/factors"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/hpfn_config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/price_sync_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/region_rate_table_config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/shop_ops_audit_log"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_v2_db"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	price_sync_db.ProviderSet,
	sip_db.ProviderSet,
	sip_v2_db.ProviderSet,
	hpfn_config.ProviderSet,
	account_service.ProviderSet,
	factors.ProviderSet,
	region_rate_table_config.ProviderSet,
	edit_item_price_allow_list.ProviderSet,
	shop_ops_audit_log.NewAuditLogRepo,
)
