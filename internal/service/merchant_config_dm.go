package service

import (
	"context"
	"fmt"
	"time"

	commonCache "git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	internalMerchantConfigSettingPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_constraints.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type MerchantConfigServiceDm struct {
	commonCache      cache.CommonCache
	localCacheManger cache.MerchantConfigSettingLocalCacheManager
	redisCacheManger cache.MerchantConfigSettingRedisCacheManager
	merchantConfigDb *db.MerchantConfigDB
}

func (dm *MerchantConfigServiceDm) GetSingleMerchantConfigSettingInfoMap(ctx context.Context, merchantId uint64) (map[uint64]*internalMerchantConfigSettingPb.MerchantConfigSetting, error) {
	infoMap := dm.GetMerchantConfigSettingInfoMap(ctx, []uint64{merchantId})
	if infoMap[merchantId].Err != nil {
		return nil, infoMap[merchantId].Err
	}

	if infoMap[merchantId] == nil {
		return nil, cerr.New(fmt.Sprintf("failed to get merchant config for merchantId=%v", merchantId), uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_MERCHANT_CONFIG_SETTING))
	}
	return infoMap[merchantId].MerchantConfigSettingMap, nil
}

func (dm *MerchantConfigServiceDm) SetMerchantConfigSettings(ctx context.Context,
	merchantId uint64, settings []*internalMerchantConfigSettingPb.MerchantConfigSetting) error {

	err := dm.merchantConfigDb.SetMerchantConfigSettingList(ctx, merchantId, settings)
	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to SetMerchantConfigSettingList|merchantId=%d|settings=%v|err=%s",
			merchantId, cutil.LazyJSONEncoder(settings), err.Error()))
		return err
	}

	dm.purgeMerchantConfigSettingCache(ctx, merchantId)
	return nil
}

func (dm *MerchantConfigServiceDm) GetProfitRateLimit(ctx context.Context, region, merchantRegion string) ([]*internal.MerchantConstraints, error) {
	cacheKey := constant.GetProfitRateLimitCacheKey(merchantRegion)
	logging.GetLogger(ctx).Debug(fmt.Sprintf("GetProfitRateLimitList cache key = %v", cacheKey))
	profitRateLimits, err := dm.commonCache.GetProfitRateLimitList(ctx, cacheKey)
	if err == nil {
		logging.GetLogger(ctx).Debug("GetProfitRateLimitList fetched profit rate limit from cache")
		for _, prl := range profitRateLimits {
			logging.GetLogger(ctx).Debug(fmt.Sprintf("GetProfitRateLimitList profit rate limit = %v", cutil.JSONEncode(prl)))
		}
		return profitRateLimits, nil
	}
	profitRateLimits, err = dm.merchantConfigDb.GetProfitRateLimitList(ctx, region, merchantRegion)
	if err != nil {
		return nil, err
	}

	expireTime := time.Duration(config.GetCommonConfig().CbscProfitRateLimitCacheExpireSeconds) * time.Second
	_ = dm.commonCache.Set(ctx, cacheKey, profitRateLimits, expireTime)
	return profitRateLimits, nil
}

func (dm *MerchantConfigServiceDm) UpdateProfitRateLimit(ctx context.Context, region, merchantRegion string, profitRateMin, profitRateMax *float64, operator string) error {
	err := dm.merchantConfigDb.UpdateProfitRateLimit(ctx, region, merchantRegion, profitRateMin, profitRateMax, operator)
	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to update profit rate limit|region=%s|merchantRegion=%s|opreator=%s|err=%s",
			region, merchantRegion, operator, err.Error()))
		return err
	}
	dm.purgeProfitRateLimitCache(ctx, merchantRegion)
	return nil
}

func (dm *MerchantConfigServiceDm) purgeMerchantConfigSettingCache(ctx context.Context, merchantId uint64) {
	redisCacheKey := dm.redisCacheManger.Key(merchantId)
	_ = dm.redisCacheManger.Del(ctx, redisCacheKey)

	localCacheKey := dm.localCacheManger.Key(merchantId)
	_ = dm.localCacheManger.Del(ctx, localCacheKey)
}

func (dm *MerchantConfigServiceDm) purgeProfitRateLimitCache(ctx context.Context, merchantRegion string) {
	_ = dm.commonCache.Del(ctx, constant.GetProfitRateLimitCacheKey(merchantRegion))
}

func NewMerchantConfigService(localCacheManger cache.MerchantConfigSettingLocalCacheManager, redisCacheManger cache.MerchantConfigSettingRedisCacheManager, merchantConfigDb *db.MerchantConfigDB, commonCache cache.CommonCache) MerchantConfigService {
	return &MerchantConfigServiceDm{
		commonCache:      commonCache,
		localCacheManger: localCacheManger,
		redisCacheManger: redisCacheManger,
		merchantConfigDb: merchantConfigDb,
	}
}

func (dm *MerchantConfigServiceDm) GetMerchantConfigSettingInfoMap(ctx context.Context, merchantIdList []uint64) map[uint64]*MerchantConfigSettingInfo {
	merchantConfigInfoMap := make(map[uint64]*MerchantConfigSettingInfo)
	for _, merchantId := range merchantIdList {
		merchantConfigSettingMap, err := dm.getMerchantConfigSettingMap(ctx, merchantId)
		if err != nil {
			errMsg := fmt.Sprintf("failed to get merchant config setting, merchantId=%d, err=%s",
				merchantId, err.Error())
			logging.GetLogger(ctx).Error(errMsg)

			merchantConfigInfoMap[merchantId] = &MerchantConfigSettingInfo{
				Err: cerr.New(errMsg,
					uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_MERCHANT_CONFIG_SETTING)),
				MerchantConfigSettingMap: nil,
			}

			continue
		}
		merchantConfigInfoMap[merchantId] = &MerchantConfigSettingInfo{
			MerchantConfigSettingMap: merchantConfigSettingMap,
		}
	}

	return merchantConfigInfoMap
}

func (dm *MerchantConfigServiceDm) getMerchantConfigSettingMap(ctx context.Context, merchantId uint64) (map[uint64]*internalMerchantConfigSettingPb.MerchantConfigSetting, error) {
	merchantConfigSettingList, err := dm.getMerchantConfigSettingList(ctx, merchantId)
	if err != nil {
		return nil, err
	}

	merchantShopConfigSettingMap := make(map[uint64]*internalMerchantConfigSettingPb.MerchantConfigSetting)
	for _, m := range merchantConfigSettingList {
		merchantShopConfigSettingMap[m.GetShopId()] = m
	}
	return merchantShopConfigSettingMap, nil
}

func (dm *MerchantConfigServiceDm) getMerchantConfigSettingList(ctx context.Context, merchantId uint64) ([]*internalMerchantConfigSettingPb.MerchantConfigSetting, error) {
	// read from local cache
	localCacheKey := dm.localCacheManger.Key(merchantId)
	value, err := dm.localCacheManger.Get(ctx, localCacheKey)
	if err == nil {
		logging.GetLogger(ctx).Debug(fmt.Sprintf(
			"success to read from local cache for MerchantConfig, key=%s", localCacheKey))
		return value, nil
	}
	if err != commonCache.ErrCacheMiss {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to read from local cache for MerchantConfig, err=%s", err.Error()))
	}

	// if local cache miss, then read from redis
	redisCacheKey := dm.redisCacheManger.Key(merchantId)
	value, err = dm.redisCacheManger.Get(ctx, redisCacheKey)
	if err == nil { // if redis cache hit, then set local cache back
		logging.GetLogger(ctx).Debug(fmt.Sprintf(
			"success to read from redis for MerchantConfig, key=%s", redisCacheKey))

		setErr := dm.localCacheManger.Set(ctx, localCacheKey, value,
			time.Duration(config.GetCommonConfig().MerchantConfigSettingCacheExpireSecondForLocalCache)*time.Second)
		if setErr != nil {
			logging.GetLogger(ctx).Error(fmt.Sprintf(
				"failed to set local cache from redis for MerchantConfig, err=%s", setErr.Error()))
		} else {
			logging.GetLogger(ctx).Debug(fmt.Sprintf(
				"success to set local cache from redis for MerchantConfig, key=%s", localCacheKey))
		}
		return value, nil
	} else if err != commonCache.ErrCacheMiss {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to read from redis for MerchantConfig, err=%s", err.Error()))
	}

	// if redis cache miss, then read from MerchantConfigDB,
	// and set back to local cache and redis, including empty result from DB.
	value, err = dm.merchantConfigDb.GetMerchantConfigSettingList(ctx, merchantId)
	if err != nil {
		return nil, err
	}
	errForLocalCache := dm.localCacheManger.Set(ctx, localCacheKey, value,
		time.Duration(config.GetCommonConfig().MerchantConfigSettingCacheExpireSecondForLocalCache)*time.Second)
	if errForLocalCache != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to set local cache from db, key=%s, value=%s, err=%s",
			localCacheKey, cutil.JSONEncode(value), errForLocalCache.Error()))
	} else {
		logging.GetLogger(ctx).Debug(fmt.Sprintf(
			"success to set local cache from db, key=%s, value=%s",
			localCacheKey, cutil.JSONEncode(value)))
	}
	errForRedisCache := dm.redisCacheManger.Set(ctx, redisCacheKey, value,
		time.Duration(config.GetCommonConfig().MerchantConfigSettingCacheExpireSecondForRedis)*time.Second)
	if errForRedisCache != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to set redis cache, key=%s, value=%s, err=%s",
			redisCacheKey, cutil.JSONEncode(value), errForRedisCache.Error()))
	} else {
		logging.GetLogger(ctx).Debug(fmt.Sprintf(
			"success to set redis cache, key=%s, value=%s",
			redisCacheKey, cutil.JSONEncode(value)))
	}

	return value, nil
}
