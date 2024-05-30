package factors

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config/config_parser"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	internalExchangeRatePb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_exchange_rate.pb"
	ib "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/item_business.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_common_definition.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_v2_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/convutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

func (c *CalculationFactorsRepoImpl) GetShopSipRateConfigByPShopIdForCbSip(ctx context.Context, pShopId uint64) ([]model.AShopConfigInfo, error) {
	key := constant.GetShopMapCacheKey(pShopId)
	shopMaps, err := c.cache.GetAShopData(ctx, key)
	if err != nil && err != cache.ErrCacheMiss {
		logging.GetLogger(ctx).Error(fmt.Sprintf("failed to get shopMap from cache: err=%v", err), ulog.String("cacheKey", key))
	}

	if len(shopMaps) == 0 {
		aShopIds, err := c.shopCoreService.GetAShopIdsByPShopId(ctx, pShopId)
		if err != nil {
			return nil, err
		}
		if len(aShopIds) == 0 {
			return nil, nil
		}

		session := c.sipRepo.DbSession()
		shopMaps, err = c.sipRepo.GetShopMapWithoutOffboardByAShopIdsAndPShopId(ctx, session, pShopId, aShopIds)
		if err != nil {
			return nil, err
		}

		if len(shopMaps) > 0 {
			// 60s expiration
			err = c.cache.Set(ctx, key, shopMaps, time.Duration(config.GetCommonConfig().ShopMapCacheExpireSeconds)*time.Second)
			if err != nil {
				logging.GetLogger(ctx).Error("failed to set shopMap from cache", ulog.String("cacheKey", key), ulog.Error(err))
			}
		}
	}

	res := make([]model.AShopConfigInfo, 0)
	for _, shopMap := range shopMaps {
		res = append(res, model.AShopConfigInfo{
			AShopId: shopMap.GetAffiShopid(),
			SipRate: float64(shopMap.GetPriceRatio()) / constant.DbPectInflationFactor,
		})
	}
	return res, nil
}

func (c *CalculationFactorsRepoImpl) GetAllCountryMarginMapForCbSip(ctx context.Context) (map[string]map[string]float64, error) {
	countryMarginMap, err := c.getCountryMarginMapForCbSip(ctx)
	if err != nil {
		return nil, err
	}
	return countryMarginMap, nil
}

func (c *CalculationFactorsRepoImpl) GetAllExchangeRateMapForCbSip(ctx context.Context) (map[string]map[string]string, error) {
	session := c.sipRepo.DbSession()
	allExchangeRate, err := c.sipRepo.GetAllExchangeRate(ctx, session)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]string)
	for _, exchangeRate := range allExchangeRate {
		if _, ok := result[exchangeRate.SourceCurrency]; ok {
			result[exchangeRate.SourceCurrency][exchangeRate.TargetCurrency] = exchangeRate.ExchangeRate
		} else {
			result[exchangeRate.SourceCurrency] = make(map[string]string)
			result[exchangeRate.SourceCurrency][exchangeRate.TargetCurrency] = exchangeRate.ExchangeRate
		}
		value, err := strconv.ParseFloat(exchangeRate.ExchangeRate, 64)
		if err != nil {
			logging.GetLogger(ctx).Error(fmt.Sprintf("fail to ParseFloat(%s),err:%v", exchangeRate.ExchangeRate, err))
			return nil, err
		}
		if _, ok := result[exchangeRate.TargetCurrency]; ok {
			result[exchangeRate.TargetCurrency][exchangeRate.SourceCurrency] = convutil.FloatFormat(1.0/value, 10)
		} else {
			result[exchangeRate.TargetCurrency] = make(map[string]string)
			result[exchangeRate.TargetCurrency][exchangeRate.SourceCurrency] = convutil.FloatFormat(1.0/value, 10)
		}
	}
	for sourceCurrency := range result {
		result[sourceCurrency][sourceCurrency] = "1"
	}
	return result, nil
}

func (c *CalculationFactorsRepoImpl) GetCurrencyForCbSip(ctx context.Context, merchantId uint64, aRegion string, pItemInfo *ib.ProductInfo) (srcCurrency, dstCurrency string, err error) {
	// for srcCurrency, use P itemprice currency first, if not exists, then use merchant region
	pItemCurrency := c.pickCurrencyFromItemInfo(pItemInfo)
	if len(pItemCurrency) > 0 {
		srcCurrency = pItemCurrency
	} else {
		var exchangeRateInfo *internalExchangeRatePb.ExchangeRateInfo
		exchangeRateInfo, err = c.exchangeRateService.GetMerchantExchangeRateInfo(ctx, merchantId)
		if err != nil {
			return "", "", err
		}
		srcCurrency = exchangeRateInfo.GetCurrency()
	}
	if len(srcCurrency) == 0 {
		return "", "", cerr.New("srcCurrency is empty",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	// for dstCurrency
	dstCurrency = config.GetSIPRegionCommonConf().GetByRegion(aRegion).Currency
	if len(dstCurrency) == 0 {
		return "", "", cerr.New("dstCurrency is empty",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	return srcCurrency, dstCurrency, nil
}

func (c *CalculationFactorsRepoImpl) pickCurrencyFromItemInfo(itemInfo *ib.ProductInfo) string {
	for _, modelInfo := range itemInfo.GetModels() {
		for _, modelPrice := range modelInfo.GetPrice().GetOngoingPrices() {
			if modelPrice.GetRuleType() == uint32(price_common_definition.Constant_RULE_TYPE_SHOPEE_MANAGE_ITEM_PRICE) {
				currency := modelPrice.GetCurrency()
				if len(currency) > 0 {
					return currency
				}
			}
		}
	}
	return ""
}

func (c *CalculationFactorsRepoImpl) GetCountryMarginForCbSip(ctx context.Context, pRegion string, aRegion string) (float64, error) {
	countryMargin, err := c.getCountryMarginMapForCbSip(ctx)
	if err != nil {
		return 0, err
	}
	if countryMargin == nil || countryMargin[pRegion] == nil {
		return 0, nil
	}
	return countryMargin[pRegion][aRegion], nil
}

func (c *CalculationFactorsRepoImpl) getCountryMarginMapForCbSip(ctx context.Context) (map[string]map[string]float64, error) {
	key := constant.CountryMarginConfigCacheKey
	countryMarginStr, err := c.cache.GetCountryMargin(ctx, key)
	if err != nil || len(countryMarginStr) == 0 {
		session := c.sipRepo.DbSession()
		res, err := c.sipRepo.GetSystemConfigRecordByType(ctx, session, sip_db.SystemConfigTypeContryMargin)
		if err != nil {
			return nil, err
		}
		if res == nil {
			return c.buildDefaultCountryMargin(), nil
		}
		countryMarginStr = res.ConfigData
		_ = c.cache.Set(ctx, key, countryMarginStr, 60*time.Second)
	}

	countryMargin := make(map[string]map[string]float64)
	err = json.Unmarshal([]byte(countryMarginStr), &countryMargin)
	if err != nil {
		return nil, err
	}
	return countryMargin, nil
}

func (c *CalculationFactorsRepoImpl) buildDefaultCountryMargin() map[string]map[string]float64 {
	result := make(map[string]map[string]float64)
	for _, mstRegion := range config.GetSipRegionList() {
		result[mstRegion] = make(map[string]float64)
		for _, affiRegion := range config.GetSipRegionList() {
			if mstRegion == affiRegion {
				continue
			}
			result[mstRegion][affiRegion] = 0
		}
	}
	return result
}

func (c *CalculationFactorsRepoImpl) GetHandlingFeeForCbSip(ctx context.Context) (float64, error) {
	cfgVal, err := config.GetHandlingFeeConfig()
	if err != nil {
		return 0, err
	}
	return calcutil.ToRealPect(int(cfgVal)), nil
}

func (c *CalculationFactorsRepoImpl) GetShopCommissionFeeForCbSip(ctx context.Context, region string, shopId int64) (float64, error) {
	return c.integratedFeeService.GetShopFee(ctx, region, shopId, model.ShopFeeTypeCommissionFee)
}

func (c *CalculationFactorsRepoImpl) GetShopServiceFeeForCbSip(ctx context.Context, region string, shopId int64) (float64, error) {
	return c.integratedFeeService.GetShopFee(ctx, region, shopId, model.ShopFeeTypeServiceFee)
}

func (c *CalculationFactorsRepoImpl) GetHiddenPriceForCbSip(ctx context.Context, pItem *sip_v2_db.MstItemRecord, pShop *sip_db.MstShop, merchantRegion string, pRegion string, aRegion string, weight float64, psite bool) float64 {
	conf := c.getHiddenPriceConfigForCbSip(ctx, pItem, pShop, merchantRegion, pRegion, aRegion, weight, psite)
	if conf == nil {
		return 0
	}
	return c.calcHiddenPriceByHiddenPriceConfig(ctx, weight, conf)
}

func (c *CalculationFactorsRepoImpl) calcHiddenPriceByHiddenPriceConfig(ctx context.Context, weight float64, conf *model.HiddenPriceConf) float64 {
	if weight <= 0 {
		return 0
	}

	startPrice := calcutil.ToRealPrice(conf.StartPrice)
	price := calcutil.ToRealPrice(conf.Price)
	startWeight := calcutil.DbWeightToGram(conf.StartWeight)
	weightStep := calcutil.DbWeightToGram(conf.WeightStep)
	adjustment := calcutil.ToRealPrice(conf.Adjustment)

	var roundSize int64
	if conf.RoundSize <= 0 {
		roundSize = int64(math.Pow10(-int(conf.RoundSize)))
	} else {
		pow := math.Pow10(int(conf.RoundSize))
		startWeight *= pow
		roundSize = 1
		weight *= pow
		weightStep *= pow
	}

	res := startPrice + float64((calcutil.RoundUp(int(startWeight), int(roundSize), weight)-int(startWeight))/int(weightStep))*price + adjustment
	if res > 0 {
		return res
	} else {
		return 0
	}
}

func (c *CalculationFactorsRepoImpl) getHiddenPriceConfigForCbSip(ctx context.Context, pItem *sip_v2_db.MstItemRecord, pShop *sip_db.MstShop, merchantRegion string, pRegion string, aRegion string, weight float64, psite bool) *model.HiddenPriceConf {
	// first, try item level config
	conf := c.getItemLevelHiddenPriceConfig(ctx, pItem, pRegion, aRegion, weight, psite)
	if conf != nil {
		return conf
	}

	// second, try shop level config
	conf = c.getShopLevelHiddenPriceConfig(ctx, pShop, aRegion, weight, psite)
	if conf != nil {
		return conf
	}

	// try region level config
	conf = c.getRegionLevelHiddenPriceConfig(ctx, merchantRegion, pRegion, aRegion, weight, psite)
	if conf != nil {
		return conf
	}

	// try default level config
	if psite {
		return c.getDefaultLevelHiddenPriceConfig(ctx, pRegion, weight)
	}
	return c.getDefaultLevelHiddenPriceConfig(ctx, aRegion, weight)
}

func (c *CalculationFactorsRepoImpl) getHpfnConfigByRateTableConfig(ctx context.Context, cfg interface{}, pRegion string, aRegion string, weight float64, psite bool) *model.HiddenPriceConf {
	rateTableCfg := model.RateTableCfg{}
	switch v := cfg.(type) {
	case string:
		if len(v) == 0 {
			return nil
		}
		err := json.Unmarshal([]byte(v), &rateTableCfg)
		if err != nil {
			logging.GetLogger(ctx).Error(fmt.Sprintf("unmarshal rate table cfg=\"%s\" failed, err=%v", cfg, err))
			return nil
		}
	case *model.RateTableCfg:
		if v == nil {
			return nil
		}
		rateTableCfg = *v
	default:
		logging.GetLogger(ctx).Error("cfg must be string or *model.RateTableCfg")
		return nil
	}

	key := ""
	if psite {
		key = rateTableCfg.MstRateKey
	} else {
		affiKey, ok := rateTableCfg.AffiRateKey[aRegion]
		if !ok {
			return nil
		}
		key = affiKey
	}
	if len(key) == 0 {
		logging.GetLogger(ctx).Error(fmt.Sprintf("rate table=%v, key is empty", rateTableCfg))
		return nil
	}
	hpfnRate, err := c.hpfnConfigRepo.GetOne(ctx, key, int(calcutil.GramToDbWeight(weight)))
	if hpfnRate == nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf("hpfnkey=%s for weight=%f not found  err=%v", key, weight, err))
		return nil
	}
	res := &model.HiddenPriceConf{
		StartPrice:  hpfnRate.StartPrice,
		StartWeight: hpfnRate.StartWeight,
		RoundSize:   hpfnRate.RoundSize,
		Price:       hpfnRate.Price,
		WeightStep:  hpfnRate.WeightStep,
		Adjustment:  hpfnRate.Adjustment,
		HpfnKey:     key,
	}
	return res
}

func (c *CalculationFactorsRepoImpl) getItemLevelHiddenPriceConfig(ctx context.Context, pItem *sip_v2_db.MstItemRecord, pRegion string, aRegion string, weight float64, psite bool) *model.HiddenPriceConf {
	if pItem == nil {
		return nil
	}
	res := c.getHpfnConfigByRateTableConfig(ctx, pItem.RateTableCfg, pRegion, aRegion, weight, psite)
	if res == nil {
		return nil
	}
	if res.HpfnCfgLevel == 0 {
		res.HpfnCfgLevel = model.HpfnItemLevelCfg
	}
	logging.GetLogger(ctx).Info(fmt.Sprintf("HPFN rate table found item level config: pitemid=%d, affi_region=%s, weight=%f, psite=%t, hpfn config=%+v", pItem.ItemId, aRegion, weight, psite, res))
	return res
}

func (c *CalculationFactorsRepoImpl) getShopLevelHiddenPriceConfig(ctx context.Context, pShop *sip_db.MstShop, aRegion string, weight float64, psite bool) *model.HiddenPriceConf {
	if pShop == nil {
		return nil
	}

	res := c.getHpfnConfigByRateTableConfig(ctx, pShop.RateTableCfg, pShop.Country, aRegion, weight, psite)
	if res == nil {
		return nil
	}
	if res.HpfnCfgLevel == 0 {
		res.HpfnCfgLevel = model.HpfnShopLevelCfg
	}
	logging.GetLogger(ctx).Info(fmt.Sprintf("HPFN rate table found shop level config: mstShop=%v, weight=%f, psite=%t, hpfn config=%+v", pShop, weight, psite, res))
	return res
}

func (c *CalculationFactorsRepoImpl) getRegionLevelHiddenPriceConfig(ctx context.Context, merchantRegion, pRegion, aRegion string, weight float64, psite bool) *model.HiddenPriceConf {
	syscfg := c.regionRateTableConfigRepo.GetRegionRateTableConfig()
	if syscfg == nil {
		logging.GetLogger(ctx).Debug("region rate table not configured")
		return nil
	}
	if merchantRegion == "" {
		merchantRegion = "CN"
		logging.GetLogger(ctx).Info("merchentRegion is empty, use CN as default merchantRegion")
	}
	if merchantCfg, ok := syscfg[merchantRegion]; ok {
		if mstCfg, ok := merchantCfg[pRegion]; ok {
			res := c.getHpfnConfigByRateTableConfig(ctx, mstCfg, pRegion, aRegion, weight, psite)
			if res != nil {
				if res.HpfnCfgLevel == 0 {
					res.HpfnCfgLevel = model.HpfnRegionLevelCfg
				}
				logging.GetLogger(ctx).Info(fmt.Sprintf("HPFN rate table found region level config, merchant_region=%s, mst_region=%s, psite=%t, hpfn config=%+v", merchantRegion, pRegion, psite, res))
				return res
			}
		}
	}
	return nil
}

func (c *CalculationFactorsRepoImpl) getDefaultLevelHiddenPriceConfig(ctx context.Context, region string, weight float64) *model.HiddenPriceConf {
	config, err := config_parser.GetHpfnConfigMapFromConfig()
	if err != nil {
		return nil
	}
	cfgs, ok := config[region]
	if !ok {
		return nil
	}
	for _, cfg := range cfgs {
		dbWeight := calcutil.GramToDbWeight(weight)
		if cfg.WeightRange >= dbWeight {
			return &model.HiddenPriceConf{
				StartPrice:   cfg.StartPrice,
				StartWeight:  cfg.StartWeight,
				RoundSize:    cfg.RoundSize,
				Price:        cfg.Price,
				WeightStep:   cfg.WeightStep,
				Adjustment:   cfg.Adjustment,
				HpfnKey:      "-", // default
				HpfnCfgLevel: model.HpfnDefaultLevelCfg,
			}
		}
	}
	return nil
}

func (c *CalculationFactorsRepoImpl) GetExchangeRateForCbSip(ctx context.Context, sourceCurrency, targetCurrency string, needProcessPrecision bool) (float64, error) {
	if strings.EqualFold(sourceCurrency, targetCurrency) {
		return 1, nil
	}

	exchangeRateConfig := &model.ExchangeRateCacheData{}

	currencyPair := model.BuildCurrencyPair(sourceCurrency, targetCurrency)
	cacheKey := constant.GetExchangeRateCacheKey(currencyPair)
	cacheExchangeRateConfig, err := c.cache.GetString(ctx, cacheKey)
	if err != nil || cacheExchangeRateConfig == "" {
		session := c.sipRepo.DbSession()
		exchangeRateConfigDB, err := c.sipRepo.GetExchangeRateByCurrency(ctx, session, currencyPair)
		if err != nil {
			logging.GetLogger(ctx).Error(fmt.Sprintf("fail to find exchange rate config, err:%v", err))
			return 0, err
		}

		exchangeRateConfig.SourceCurrency = exchangeRateConfigDB.SourceCurrency
		exchangeRateConfig.TargetCurrency = exchangeRateConfigDB.TargetCurrency
		exchangeRateConfig.ExchangeRate = exchangeRateConfigDB.ExchangeRate
		exchangeRateConfigBytes, _ := json.Marshal(exchangeRateConfig)
		_ = c.cache.Set(ctx, cacheKey, string(exchangeRateConfigBytes), 60)
	} else {
		err = json.Unmarshal([]byte(cacheExchangeRateConfig), &exchangeRateConfig)
		if err != nil {
			return 0, err
		}
	}

	exchangeRate, err := strconv.ParseFloat(exchangeRateConfig.ExchangeRate, 64)
	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf("fail to ParseFloat(%s),err:%v", exchangeRateConfig.ExchangeRate, err))
		return 0, err
	}

	if exchangeRateConfig.SourceCurrency == sourceCurrency && exchangeRateConfig.TargetCurrency == targetCurrency {
		return exchangeRate, nil
	} else {
		// TODO: due to legacy code issue, sip item price & a price calculation has different exchange rate process logic
		//  so in this phase we just keep the old logic.
		//  for sip item price exchange rate, need to process the precision
		//  for a price calculate, don't need to handle the precision
		if !needProcessPrecision {
			return 1.0 / exchangeRate, nil
		}
		reversedExchangeRateStr := convutil.FloatFormat(1.0/exchangeRate, 10)
		reversedExchangeRate, err := strconv.ParseFloat(reversedExchangeRateStr, 64)
		if err != nil {
			return 0, err
		}
		return reversedExchangeRate, nil
	}
}
