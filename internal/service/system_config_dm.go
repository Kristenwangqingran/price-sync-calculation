package service

import (
	"context"
	"encoding/json"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	db2 "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type systemConfigService struct {
	systemConfigDB db2.SystemConfigDB
	cacheManager   cache.SystemConfigCacheManager
}

func NewSystemConfigService(systemConfigDB db2.SystemConfigDB, cacheManager cache.SystemConfigCacheManager) SystemConfigService {
	return &systemConfigService{
		systemConfigDB: systemConfigDB,
		cacheManager:   cacheManager,
	}
}

func (dm *systemConfigService) GetAllLocalPriceConfig(ctx context.Context) (map[string]map[string]*model.CommonPriceConfig, error) {
	record, err := dm.systemConfigDB.GetSystemConfigRecordByType(ctx, localPriceConfigNewType)
	if err != nil {
		return nil, err
	}
	localPriceConfigMap := make(map[string]map[string]*model.CommonPriceConfig)
	if err := json.Unmarshal([]byte(record.ConfigData), &localPriceConfigMap); err != nil {
		iErr := cerr.Wrap(err, "unmarshal local price config fail", uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
		logging.GetLogger(ctx).Error("GetLocalPriceConfigByRegion failed", ulog.Error(iErr))
		return nil, iErr
	}

	return localPriceConfigMap, nil
}

func (dm *systemConfigService) GetLocalPriceConfigByRegion(ctx context.Context,
	primaryRegion, affiRegion string) (*model.CommonPriceConfig, error) {

	cacheKey := dm.cacheManager.LocalSIPConfigKey(primaryRegion, affiRegion)
	localPriceConfig, err := dm.cacheManager.GetLocalSIPConfig(ctx, cacheKey)
	if err == nil && localPriceConfig != nil {
		return localPriceConfig, nil
	}

	localPriceConfigMap, err := dm.GetAllLocalPriceConfig(ctx)
	if err != nil {
		return nil, err
	}

	if _, existMst := localPriceConfigMap[primaryRegion]; !existMst {
		iErr := cerr.New("primary region config not found", uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
		logging.GetLogger(ctx).Error("GetLocalPriceConfigByRegion failed",
			ulog.String("primaryRegion", primaryRegion), ulog.Reflect("config", localPriceConfigMap), ulog.Error(iErr))
		return nil, iErr
	}
	if _, existAffi := localPriceConfigMap[primaryRegion][affiRegion]; !existAffi {
		iErr := cerr.New("affi region config not found", uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
		logging.GetLogger(ctx).Error("GetLocalPriceConfigByRegion failed",
			ulog.String("affiRegion", affiRegion), ulog.Reflect("config", localPriceConfigMap), ulog.Error(iErr))
		return nil, iErr
	}

	localPriceConfig = localPriceConfigMap[primaryRegion][affiRegion]
	_ = dm.cacheManager.SetLocalSIPConfig(ctx, cacheKey, localPriceConfig)

	return localPriceConfig, nil
}

func (dm *systemConfigService) GetChannelWhitelist(ctx context.Context) ([]int64, error) {
	cacheKey := dm.cacheManager.ChannelWhiteListKey()
	channelWhiteList, err := dm.cacheManager.GetChannelWhiteList(ctx, cacheKey)
	if err == nil && channelWhiteList != nil {
		return channelWhiteList, nil
	}

	record, err := dm.systemConfigDB.GetSystemConfigRecordByType(ctx, channelWhitelistType)
	if err != nil {
		return nil, err
	}

	channelWhiteList = make([]int64, 0)
	if err := json.Unmarshal([]byte(record.ConfigData), &channelWhiteList); err != nil {
		iErr := cerr.Wrap(err, "unmarshal channelWhiteList fail", uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
		logging.GetLogger(ctx).Error("GetChannelWhitelist failed", ulog.Error(iErr))
		return nil, iErr
	}

	_ = dm.cacheManager.SetChannelWhiteList(ctx, cacheKey, channelWhiteList)

	return channelWhiteList, nil
}

func (dm *systemConfigService) GetDefaultCBSIPPriceRateLimit(ctx context.Context) (configStr string, err error) {
	//TODO
	return configStr, nil
}
