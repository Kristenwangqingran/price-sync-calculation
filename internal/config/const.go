package config

// spex config field name
const (
	KeyCommon                   = "common"
	KeyRedis                    = "redis"
	KeyHttpApi                  = "http_api"
	KeyBatch                    = "batch"
	KeyPriceSyncDB              = "price_sync_db"
	KeyMerchantConfigDB         = "merchant_config_db"
	KeyCbscPriceConfig          = "cbsc_price_config"
	KeyMerchantConfigLocalCache = "merchant_config_local_cache"
	KeyShopConfigLocalCache     = "shop_local_cache"
	KeyLocalSipPriceConfig      = "local_sip_price_config"
	KeyDataInfraConfig          = "data_infra_config"
	keyThreadPoolConfig         = "thread_pool_config"
	KeyOPLConfig                = "opl"

	KeySipDB      = "sip_db"
	KeySipV2DB    = "sip_v2_db"
	KeyAuditLogDB = "audit_log_db"

	// TODO: update config architecture @wang.zhong
	KeySIPMigrationConfig = "sip_migration_config"
)
