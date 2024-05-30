package config

import (
	"git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/common/gdbc/hardy"
	"git.garena.com/shopee/common/uniconfig"
)

type Config struct {
	CommonCfg                   *CommonConfig
	RedisCacheConfig            *CacheConfig
	HTTPApiConfig               *HTTPApiConfig
	BatchCfg                    *BatchConfig
	MerchantConfigDBConfig      *hardy.Config
	PriceSyncDBConfig           *hardy.Config
	LocalCacheForMerchantConfig *cache.InMemoryCacheConfig
	LocalCacheForShop           *cache.InMemoryCacheConfig
	CbscPriceConfig             *CbscPriceConfig
	OPLConfig                   *OPLConfig
	SipDBCfg                    *hardy.Config
	SipV2DBCfg                  *hardy.Config
	AuditLogDBCfg               *hardy.Config
	SIPMigrationCfg             *SIPMigrationConfig
	LocalSipPriceConfig         *LocalSipPriceConfig
	DataInfraConfig             *DataInfraConfig
	ThreadPoolConfig            *ThreadPoolConfig
}

var (
	confVal = &Config{}
)

// InitializeSpexConfigs init configs after creating the service
func InitializeSpexConfigs() {
	bindSpexConfigs()
	getAndCastSpexConfigs()
	applySpexConfigs()
	watchSpexConfigs()
}

func bindSpexConfigs() {
	bindConfig(KeyCommon, &CommonConfig{})
	bindConfig(KeyRedis, &CacheConfig{})
	bindConfig(KeyHttpApi, &HTTPApiConfig{})
	bindConfig(KeyBatch, &BatchConfig{})
	bindConfig(KeyPriceSyncDB, &hardy.Config{})
	bindConfig(KeyMerchantConfigDB, &hardy.Config{})
	bindConfig(KeyCbscPriceConfig, &CbscPriceConfig{})
	bindConfig(KeyLocalSipPriceConfig, &LocalSipPriceConfig{})
	bindConfig(KeySipDB, &hardy.Config{})
	bindConfig(KeySipV2DB, &hardy.Config{})
	bindConfig(KeyAuditLogDB, &hardy.Config{})
	bindConfig(KeySIPMigrationConfig, &SIPMigrationConfig{})
	bindConfig(KeyMerchantConfigLocalCache, &cache.InMemoryCacheConfig{})
	bindConfig(KeyShopConfigLocalCache, &cache.InMemoryCacheConfig{})
	bindConfig(KeyDataInfraConfig, &DataInfraConfig{})
	bindConfig(keyThreadPoolConfig, &ThreadPoolConfig{})
	bindConfig(KeyOPLConfig, &OPLConfig{})
}

func bindConfig(key string, proto interface{}) {
	err := uniconfig.BindProto(key, proto)
	if err != nil {
		panic(err)
	}
}

func getAndCastSpexConfigs() {
	rawSpexConfigForCommon := getConfig(KeyCommon)
	confVal.CommonCfg, _ = rawSpexConfigForCommon.(*CommonConfig)

	rawSpexConfigForRedis := getConfig(KeyRedis)
	confVal.RedisCacheConfig, _ = rawSpexConfigForRedis.(*CacheConfig)

	rawSpexConfigForHttpApi := getConfig(KeyHttpApi)
	confVal.HTTPApiConfig, _ = rawSpexConfigForHttpApi.(*HTTPApiConfig)

	rawSpexConfigForBatch := getConfig(KeyBatch)
	confVal.BatchCfg, _ = rawSpexConfigForBatch.(*BatchConfig)

	rawPriceSyncDBConfig := getConfig(KeyPriceSyncDB)
	confVal.PriceSyncDBConfig, _ = rawPriceSyncDBConfig.(*hardy.Config)

	rawDBConfig := getConfig(KeyMerchantConfigDB)
	confVal.MerchantConfigDBConfig, _ = rawDBConfig.(*hardy.Config)

	rawCbscPriceConfig := getConfig(KeyCbscPriceConfig)
	confVal.CbscPriceConfig, _ = rawCbscPriceConfig.(*CbscPriceConfig)

	rawLocalSipPriceConfig := getConfig(KeyLocalSipPriceConfig)
	confVal.LocalSipPriceConfig, _ = rawLocalSipPriceConfig.(*LocalSipPriceConfig)

	rawSipDBConfig := getConfig(KeySipDB)
	confVal.SipDBCfg, _ = rawSipDBConfig.(*hardy.Config)

	rawSipV2DBConfig := getConfig(KeySipV2DB)
	confVal.SipV2DBCfg, _ = rawSipV2DBConfig.(*hardy.Config)

	rawAuditLogDBConfig := getConfig(KeyAuditLogDB)
	confVal.AuditLogDBCfg, _ = rawAuditLogDBConfig.(*hardy.Config)

	rawSIPMigrationConfig := getConfig(KeySIPMigrationConfig)
	confVal.SIPMigrationCfg, _ = rawSIPMigrationConfig.(*SIPMigrationConfig)

	rawMerchantConfigLocalCache := getConfig(KeyMerchantConfigLocalCache)
	confVal.LocalCacheForMerchantConfig, _ = rawMerchantConfigLocalCache.(*cache.InMemoryCacheConfig)

	rawShopConfigLocalCache := getConfig(KeyShopConfigLocalCache)
	confVal.LocalCacheForShop, _ = rawShopConfigLocalCache.(*cache.InMemoryCacheConfig)

	rawDataInfraCfg := getConfig(KeyDataInfraConfig)
	confVal.DataInfraConfig, _ = rawDataInfraCfg.(*DataInfraConfig)

	rawThreadPoolConfig := getConfig(keyThreadPoolConfig)
	confVal.ThreadPoolConfig, _ = rawThreadPoolConfig.(*ThreadPoolConfig)

	rawOPLConfig := getConfig(KeyOPLConfig)
	confVal.OPLConfig, _ = rawOPLConfig.(*OPLConfig)
}

func getConfig(key string) interface{} {
	rawSpexConfig, err := uniconfig.Get(key)
	if err != nil {
		panic(err)
	}
	return rawSpexConfig
}

func applySpexConfigs() {
	applyRedisCacheConfig()
	applyBatchConfig()
	applyMerchantConfigDBConfig()
	applyPriceSyncDBConfig()
	applySipDBConfig()
	applySipV2DBConfig()
	applyAuditLogConfig()
	applyLocalCacheCacheConfig()
	applyThreadPoolConfig()
}

func watchSpexConfigs() {
	watchConfig(KeyCommon, onCommonConfigUpdate)
	watchConfig(KeyHttpApi, onHTTPApiConfigUpdate)
	watchConfig(KeyBatch, onBatchConfigUpdate)
	watchConfig(KeyCbscPriceConfig, onCbscPriceConfigUpdate)
	watchConfig(KeyLocalSipPriceConfig, onLocalSipPriceConfigUpdate)
	watchConfig(KeySIPMigrationConfig, onSIPMigrationConfigUpdate)
	watchConfig(KeyDataInfraConfig, onDataInfraConfigUpdate)
	watchConfig(keyThreadPoolConfig, onThreadPoolConfigUpdate)
	watchConfig(KeyOPLConfig, onOPLConfigUpdate)
}

func watchConfig(key string, callback uniconfig.EventCallback) {
	err := uniconfig.WatchKey(key, callback)
	if err != nil {
		panic(err)
	}
}

// SetGlobal replaces the global config with the given one.
// Currently used for self test.
func SetGlobal(cfg *Config) {
	confVal = cfg
}
