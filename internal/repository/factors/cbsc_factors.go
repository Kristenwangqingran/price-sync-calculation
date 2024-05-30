package factors

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/a_item"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/a_shop"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/http"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	internalExchangeRatePb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_exchange_rate.pb"
	internalMerchantConfigSettingPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
	internalMerchantConstraintsPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_constraints.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/account_service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/hpfn_config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/region_rate_table_config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/convutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/httpcliutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logicutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/threadpool"
)

type CalculationFactorsRepoImpl struct {
	httpCli httpcliutil.HTTPCli
	cache   cache.CommonCache

	aItemDataDM               a_item.AItemDataDM
	aShopDataDM               a_shop.AShopDataDM
	sipRepo                   sip_db.SipRepo
	hpfnConfigRepo            hpfn_config.HpfnConfigRepo
	regionRateTableConfigRepo region_rate_table_config.RegionRateTableConfigRepo

	shopCoreService             service.ShopCoreService
	exchangeRateService         service.ExchangeRateService
	integratedFeeService        service.OrderAccountIntegratedFeeService
	localSipSystemConfigService service.SystemConfigService
	itemService                 service.ItemService
	logisticService             service.LogisticService
	sipItemDataService          service.SipItemDataService
	listingSIPService           service.ListingSIPService
	merchantConfigService       service.MerchantConfigService
	shopMerchantService         service.ShopMerchantService
	accountServiceRepo          account_service.AccountServiceRepo
}

type CalculationFactorsRepoOpts struct {
	HttpCli                     httpcliutil.HTTPCli
	Cache                       cache.CommonCache
	SipRepo                     sip_db.SipRepo
	HpfnConfigRepo              hpfn_config.HpfnConfigRepo
	RegionRateTableConfigRepo   region_rate_table_config.RegionRateTableConfigRepo
	ShopCoreService             service.ShopCoreService
	ExchangeRateService         service.ExchangeRateService
	IntegratedFeeService        service.OrderAccountIntegratedFeeService
	LocalSipSystemConfigService service.SystemConfigService
	ItemService                 service.ItemService
	LogisticService             service.LogisticService
	SipItemDataService          service.SipItemDataService
	ListingSIPService           service.ListingSIPService
	MerchantConfigService       service.MerchantConfigService
	ShopMerchantService         service.ShopMerchantService
	AccountServiceRepo          account_service.AccountServiceRepo
}

func NewCalculationFactorsRepoImpl(deps *CalculationFactorsRepoOpts) *CalculationFactorsRepoImpl {
	return &CalculationFactorsRepoImpl{
		httpCli:                     deps.HttpCli,
		cache:                       deps.Cache,
		sipRepo:                     deps.SipRepo,
		hpfnConfigRepo:              deps.HpfnConfigRepo,
		regionRateTableConfigRepo:   deps.RegionRateTableConfigRepo,
		shopCoreService:             deps.ShopCoreService,
		exchangeRateService:         deps.ExchangeRateService,
		integratedFeeService:        deps.IntegratedFeeService,
		localSipSystemConfigService: deps.LocalSipSystemConfigService,
		itemService:                 deps.ItemService,
		logisticService:             deps.LogisticService,
		sipItemDataService:          deps.SipItemDataService,
		listingSIPService :          deps.ListingSIPService,
		merchantConfigService:       deps.MerchantConfigService,
		shopMerchantService:         deps.ShopMerchantService,
		accountServiceRepo:          deps.AccountServiceRepo,
	}
}

func (c *CalculationFactorsRepoImpl) GetCbscExchangeRate(ctx context.Context, merchantId uint64) (string, []*pb.CbscExchangeRate, error) {
	merchantCurrency, exchangeRateByRegion, err := c.GetExchangeRateMapForCbsc(ctx, merchantId)
	if err != nil {
		return "", nil, err
	}

	exchangeRateList := make([]*pb.CbscExchangeRate, 0)
	for region, exchangeRate := range exchangeRateByRegion {
		exchangeRateList = append(exchangeRateList, &pb.CbscExchangeRate{
			ExchangeRate: proto.Float64(exchangeRate),
			Region:       proto.String(region),
		})
	}
	return merchantCurrency, exchangeRateList, nil
}

func (c *CalculationFactorsRepoImpl) GetCbscFeeRateLimit(ctx context.Context, merchantId uint64, shopRegion *string) (*pb.CbscFeeRateLimit, error) {
	feeRateLimit := &pb.CbscFeeRateLimit{}

	merchantRegion, err := c.shopMerchantService.GetMerchantRegion(ctx, merchantId)
	if err != nil {
		return nil, err
	}

	cbscShopCommonConfig, err := c.GetCbscShopPriceCommonConfig(ctx, merchantRegion)
	if err != nil {
		return nil, err
	}

	feeRateLimit.ProfitRateLimit = make([]*pb.CbscProfitRateLimit, 0)
	allProfitRateLimit, err := c.GetProfitRateLimit(ctx, "", merchantRegion)
	if err != nil {
		return nil, err
	}

	profitRateLimitByRegion := make(map[string]*internalMerchantConstraintsPb.MerchantConstraints)
	for _, limit := range allProfitRateLimit {
		profitRateLimitByRegion[limit.GetRegion()] = limit
	}
	if shopRegion != nil {
		profitRateLimit, ok := profitRateLimitByRegion[*shopRegion]
		if !ok {
			return nil, cerr.New(fmt.Sprintf("cannot find profit rate limit for region=%v", *shopRegion), uint32(pb.Constant_ERROR_NOT_FOUND))
		}
		feeRateLimit.ProfitRateLimit = append(feeRateLimit.ProfitRateLimit, &pb.CbscProfitRateLimit{
			MinProfitRate: proto.Int64(profitRateLimit.GetProfitRateMin()),
			MaxProfitRate: proto.Int64(profitRateLimit.GetProfitRateMax()),
			Region:        shopRegion,
		})
	} else {
		for region, limit := range profitRateLimitByRegion {
			feeRateLimit.ProfitRateLimit = append(feeRateLimit.ProfitRateLimit, &pb.CbscProfitRateLimit{
				MinProfitRate: proto.Int64(limit.GetProfitRateMin()),
				MaxProfitRate: proto.Int64(limit.GetProfitRateMax()),
				Region:        proto.String(region),
			})
		}
	}

	if cbscShopCommonConfig != nil {
		feeRateLimit.ServiceFeeLimit = &pb.CbscServiceFeeRateLimit{
			MinServiceFeeRate: proto.Int64(int64(cbscShopCommonConfig.ServiceFeeRateMin)),
			MaxServiceFeeRate: proto.Int64(int64(cbscShopCommonConfig.ServiceFeeRateMax)),
		}
	} else {
		return nil, cerr.New(fmt.Sprintf(
			"cannot find service fee rate limit, merchantRegion=%s", merchantRegion),
			uint32(pb.Constant_ERROR_NOT_FOUND))
	}

	return feeRateLimit, nil
}

func (c *CalculationFactorsRepoImpl) GetCbscShopLevelFeeRate(ctx context.Context, merchantId uint64, mainAccountId *uint64, shopIds []uint64) ([]*pb.CbscShopLevelFeeRate, error) {
	merchantShopList, err := c.accountServiceRepo.GetAllShopList(ctx, mainAccountId, merchantId)
	if err != nil {
		return nil, err
	}
	merchantRegion, err := c.shopMerchantService.GetMerchantRegion(ctx, merchantId)
	if err != nil {
		return nil, err
	}

	cbscShopCommonConfig, err := c.GetCbscShopPriceCommonConfig(ctx, merchantRegion)
	if err != nil {
		return nil, err
	}

	queryShopIdSet := make(map[uint64]bool)
	for _, shopId := range shopIds {
		queryShopIdSet[shopId] = true
	}
	shopsNeedCheck := c.filterRequestShop(queryShopIdSet, merchantShopList)

	shopIdRegionPairs, err := c.shopCoreService.GetShopRegionByShopIdBatch(ctx, shopsNeedCheck)
	if err != nil {
		return nil, err
	}

	shopUserStatus, err := c.accountServiceRepo.GetUserStatusMap(ctx, shopIdRegionPairs)
	if err != nil {
		return nil, err
	}

	filteredShopIdRegions := c.filterDeleteOrBannedShopOrInvalidRegion(shopIdRegionPairs, shopUserStatus)
	if len(filteredShopIdRegions) == 0 {
		logging.GetLogger(ctx).Info("empty after filtered by shop status")
		return nil, nil
	}

	commissionFeeQueries := make([]model.GetCommissionRateRequest, 0)
	for _, shopIdRegionPair := range filteredShopIdRegions {
		commissionFeeQueries = append(commissionFeeQueries, model.GetCommissionRateRequest{
			ShopId:      shopIdRegionPair.ShopId,
			MpskuRegion: shopIdRegionPair.Region,
		})
	}

	commissionRates := c.GetCommissionRateBatchForCbsc(ctx, commissionFeeQueries)
	commissionRateByShopId := make(map[uint64]uint64)
	for i, commissionResult := range commissionRates {
		if commissionResult.Err != nil { // if occurs error, then use 0 as default value
			commissionRateByShopId[commissionFeeQueries[i].ShopId] = 0
		} else {
			commissionRateByShopId[commissionFeeQueries[i].ShopId] = commissionResult.CommissionRate
		}
	}

	referenceFeeByShopIdMap := c.GetReferenceServiceFeeRateListForCbsc(ctx, filteredShopIdRegions)

	merchantConfigMap, err := c.merchantConfigService.GetSingleMerchantConfigSettingInfoMap(ctx, merchantId)
	if err != nil {
		return nil, err
	}

	shopFeeRateList := make([]*pb.CbscShopLevelFeeRate, 0)
	for _, shopIdRegion := range filteredShopIdRegions {
		shopId := shopIdRegion.ShopId
		merchantConfig, _ := merchantConfigMap[shopId]

		profitRateStatus := int32(pb.Constant_FEE_RATE_SET)
		if merchantConfig == nil || merchantConfig.ProfitRate == nil {
			profitRateStatus = int32(pb.Constant_FEE_RATE_NOT_SET)
		}

		serviceFeeRateStatus := int32(pb.Constant_FEE_RATE_SET)
		if merchantConfig == nil || merchantConfig.ServiceFeeRate == nil {
			serviceFeeRateStatus = int32(pb.Constant_FEE_RATE_NOT_SET)
		}

		shopFeeRate := &pb.CbscShopLevelFeeRate{
			ShopId:                  proto.Int64(int64(shopId)),
			TransactionFeeRate:      proto.Int64(int64(cbscShopCommonConfig.TransactionFeeRate)),
			ProfitRateStatus:        proto.Int32(profitRateStatus),
			ServiceFeeRateStatus:    proto.Int32(serviceFeeRateStatus),
			CommissionRate:          proto.Int64(int64(commissionRateByShopId[shopId])),  // if not found, then use 0 as default value
			ReferenceServiceFeeRate: proto.Int64(int64(referenceFeeByShopIdMap[shopId])), // if not found, then use 0 as default value
		}

		// only fill if exist
		if merchantConfig != nil && merchantConfig.ProfitRate != nil {
			shopFeeRate.ProfitRate = proto.Int64(int64(merchantConfig.GetProfitRate()))
		}
		if merchantConfig != nil && merchantConfig.ServiceFeeRate != nil {
			shopFeeRate.ServiceFeeRate = proto.Int64(int64(merchantConfig.GetServiceFeeRate()))
		}

		shopFeeRateList = append(shopFeeRateList, shopFeeRate)
	}

	return shopFeeRateList, nil
}

func (c *CalculationFactorsRepoImpl) SetCbscPriceFactors(ctx context.Context, query model.SetCbscPriceFactorQuery) error {
	settings := make([]*internalMerchantConfigSettingPb.MerchantConfigSetting, 0)
	for _, shopSetting := range query.ShopSettings {
		settings = append(settings, &internalMerchantConfigSettingPb.MerchantConfigSetting{
			ShopId:         proto.Uint64(shopSetting.ShopId),
			Region:         proto.String(shopSetting.Region),
			ProfitRate:     shopSetting.ProfitRate,
			ServiceFeeRate: shopSetting.ServiceFeeRate,
		})
	}
	return c.merchantConfigService.SetMerchantConfigSettings(ctx, query.MerchantId, settings)
}

func (c *CalculationFactorsRepoImpl) filterDeleteOrBannedShopOrInvalidRegion(shopIdList []model.ShopIdRegion, shopUserStatus map[uint64]int32) []model.ShopIdRegion {
	res := make([]model.ShopIdRegion, 0)
	for _, pair := range shopIdList {
		status, ok := shopUserStatus[pair.ShopId]
		if ok && (status == model.StatusAccountDelete || status == model.StatusAccountBanned) {
			continue
		}
		res = append(res, pair)
	}
	return res
}

func (c *CalculationFactorsRepoImpl) filterRequestShop(queryShopIdSet map[uint64]bool, allMerchantShops []*model.CNSCShop) []uint64 {
	shopsNeedCheck := make([]uint64, 0)
	for _, shop := range allMerchantShops {
		if shop == nil || shop.ShopId == nil {
			continue
		}

		if len(queryShopIdSet) == 0 {
			shopsNeedCheck = append(shopsNeedCheck, shop.GetShopId())
			continue
		}

		if queryShopIdSet[shop.GetShopId()] { // filter if request has specified needed shops
			shopsNeedCheck = append(shopsNeedCheck, shop.GetShopId())
		}
	}
	return shopsNeedCheck
}

func (c *CalculationFactorsRepoImpl) GetCbscShopPriceCommonConfig(ctx context.Context, merchantRegion string) (*config.CBSCPriceFeeConfig, error) {
	cfg := config.GetCbscPriceFeeConfig(merchantRegion)
	if cfg == nil {
		return nil, cerr.New(fmt.Sprintf("failed to get cbsc shop price config for merchantRegion=%s", merchantRegion), uint32(pb.Constant_ERROR_PARAMS))
	}
	return cfg, nil
}

func (c *CalculationFactorsRepoImpl) GetProfitRateLimit(ctx context.Context, region string, merchantRegion string) ([]*internalMerchantConstraintsPb.MerchantConstraints, error) {
	return c.merchantConfigService.GetProfitRateLimit(ctx, region, merchantRegion)
}

func (c *CalculationFactorsRepoImpl) UpdateProfitRateLimit(ctx context.Context, region, merchantRegion string, profitRateMin, profitRateMax *float64, operator string) error {
	return c.merchantConfigService.UpdateProfitRateLimit(ctx, region, merchantRegion, profitRateMin, profitRateMax, operator)
}

func (c *CalculationFactorsRepoImpl) GetReferenceServiceFeeRateListForCbsc(ctx context.Context, shopIdList []model.ShopIdRegion) map[uint64]uint64 {
	lock := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	t := time.Now()
	result := make(map[uint64]uint64, 0)

	for _, pair := range shopIdList {
		wg.Add(1)
		shopId := pair.ShopId
		region := pair.Region

		err := threadpool.GetThreadPool().Do(ctx, func(cctx context.Context) {
			defer wg.Done()

			rate, err := c.integratedFeeService.GetReferenceServiceFeeRate(ctx, shopId, region)
			if err != nil {
				logging.GetLogger(ctx).Error(fmt.Sprintf("GetReferenceServiceFeeRate failed: err=%v", err))
				return
			}

			lock.Lock()
			result[shopId] = rate
			lock.Unlock()
		})
		if err != nil {
			wg.Done()
			logging.GetLogger(ctx).Error("submit GetReferenceServiceFeeRate to thread pool failed",
				ulog.Error(err))
		}
	}

	wg.Wait()
	logging.GetLogger(ctx).Info("GetReferenceServiceFeeRateListForCbsc finish", ulog.Float64("cost", time.Now().Sub(t).Seconds()))

	return result
}

func (c *CalculationFactorsRepoImpl) GetCbscPriceRateBatchForCbsc(ctx context.Context, merchantRegion string, merchantConfigMap map[uint64]*internalMerchantConfigSettingPb.MerchantConfigSetting, queries []model.GetCbscPriceRateRequest) ([]model.GetCbscPriceRateResult, error) {
	cbscPriceFeeConfig := config.GetCbscPriceFeeConfig(merchantRegion)
	if cbscPriceFeeConfig == nil {
		return nil, cerr.New(fmt.Sprintf("failed to get cbsc price config for merchantRegion=%v", merchantRegion), uint32(pb.Constant_ERROR_GET_MERCHANT_CONFIG_SETTING))
	}

	finalResult := make([]model.GetCbscPriceRateResult, len(queries))
	for i, query := range queries {
		if query.CommissionRateErr != nil {
			finalResult[i] = model.GetCbscPriceRateResult{
				Err: query.CommissionRateErr,
			}
			continue
		}

		shopId := query.ShopId
		merchantConfig, ok := merchantConfigMap[shopId]
		if !ok || merchantConfig == nil {
			finalResult[i] = model.GetCbscPriceRateResult{
				Err: cerr.New(fmt.Sprintf("cannof find merchant config for shopId=%v", shopId), uint32(pb.Constant_ERROR_GET_MERCHANT_CONFIG_SETTING)),
			}
			continue
		}

		commissionRate := query.CommissionRate
		priceRate := calcutil.GetCBSCDenominatorPriceRate(ctx, cbscPriceFeeConfig, merchantConfig, commissionRate)
		finalResult[i] = model.GetCbscPriceRateResult{
			CbscPriceRate: priceRate,
		}
	}

	return finalResult, nil
}

func (c *CalculationFactorsRepoImpl) GetProfitRateBatchForCbsc(ctx context.Context, merchantConfigMap map[uint64]*internalMerchantConfigSettingPb.MerchantConfigSetting, queries []model.GetCbscProfitRateRequest) ([]model.GetCbscProfitRateResult, error) {
	finalResult := make([]model.GetCbscProfitRateResult, len(queries))
	for i, query := range queries {
		shopId := query.ShopId

		merchantConfig, ok := merchantConfigMap[shopId]
		if !ok || merchantConfig == nil || merchantConfig.ProfitRate == nil {
			finalResult[i] = model.GetCbscProfitRateResult{
				Err: cerr.New(fmt.Sprintf("cannof find merchant config for shopId=%v", shopId), uint32(pb.Constant_ERROR_GET_MERCHANT_CONFIG_SETTING)),
			}
			continue
		}

		profitRate := calcutil.RoundIntToFloat(int64(*merchantConfig.ProfitRate), constant.PercentPrecisionBetweenMtskuAndMpsku, 4)

		finalResult[i] = model.GetCbscProfitRateResult{
			ProfitRate: profitRate,
		}
	}

	return finalResult, nil
}

func (c *CalculationFactorsRepoImpl) GetCommissionRateBatchForCbsc(ctx context.Context, queries []model.GetCommissionRateRequest) []model.GetCommissionRateResult {
	filteredQueries := model.FilterGetCommissionRateReqByShopId(queries)
	commissionRateQueries := make([]*model.ShopIdRegion, 0)
	for _, query := range filteredQueries {
		commissionRateQueries = append(commissionRateQueries, &model.ShopIdRegion{
			ShopId: query.ShopId,
			Region: query.MpskuRegion,
		})
	}
	commissionRateResultMap := c.integratedFeeService.GetShopCommissionRateMap(ctx, commissionRateQueries)

	finalResult := make([]model.GetCommissionRateResult, len(queries))
	for i, query := range queries {
		commissionRateRes, ok := commissionRateResultMap[query.ShopId]
		if !ok || commissionRateRes == nil {
			errMsg := fmt.Sprintf("failed to get commission rate for shopId=%v", query.ShopId)
			logging.GetLogger(ctx).Warn(errMsg)
			finalResult[i] = model.GetCommissionRateResult{
				Err: cerr.New(errMsg, uint32(pb.Constant_ERROR_GET_SHOP_COMMISSION_RATE)),
			}
			continue
		}
		if commissionRateRes.Err != nil {
			errMsg := fmt.Sprintf("failed to get commission rate for shopId=%v, err=%v", query.ShopId, commissionRateRes.Err)
			logging.GetLogger(ctx).Warn(errMsg)
			finalResult[i] = model.GetCommissionRateResult{
				Err: cerr.New(errMsg, uint32(pb.Constant_ERROR_GET_SHOP_COMMISSION_RATE)),
			}
			continue
		}
		finalResult[i] = model.GetCommissionRateResult{
			CommissionRate: commissionRateRes.CommissionRate,
		}
	}

	return finalResult
}

func (c *CalculationFactorsRepoImpl) GetExchangeRateMapForCbsc(ctx context.Context, merchantId uint64) (string, map[string]float64, error) {
	return c.exchangeRateService.GetMerchantExchangeRate(ctx, merchantId)
}

func (c *CalculationFactorsRepoImpl) GetMerchantExchangeRateInfo(ctx context.Context, merchantId uint64) (*internalExchangeRatePb.ExchangeRateInfo, error) {
	return c.exchangeRateService.GetMerchantExchangeRateInfo(ctx, merchantId)
}

func (c *CalculationFactorsRepoImpl) GetHidePriceForCbsc(ctx context.Context, queries []model.GetHidePriceForCbscRequest) ([]model.GetHidePriceForCbscResult, error) {
	shopIdsMap := make(map[uint64]string)
	shopRegionList := make([]string, 0)
	for _, q := range queries {
		shopIdsMap[q.ShopId] = q.Region
		shopRegionList = append(shopRegionList, q.Region)
	}

	is3PfShopMap := make(map[uint64]bool)
	for shopId, region := range shopIdsMap {
		_, is3pf, err := c.shopCoreService.IsSellerWarehouseShop(ctx, shopId, region)
		if err != nil {
			return nil, err
		}
		is3PfShopMap[shopId] = is3pf
	}

	regionChannelMap, err := c.logisticService.GetChannelInfoMapForRegions(ctx, shopRegionList)
	if err != nil {
		return nil, err
	}

	errHiddenPriceResults := make([]model.GetHidePriceForCbscResult, 0)
	skipHiddenPriceQueries := make([]model.GetHidePriceForCbscRequest, 0)
	slsHiddenPriceQueries := make([]model.GetHidePriceForCbscRequest, 0)
	for _, q := range queries {
		// for non 3pf shop
		if !is3PfShopMap[q.ShopId] {
			if len(q.EnabledChannelIdList) == 0 {
				if q.IsMtskuToMpsku { // if no available channels and MT->MP, return error
					errHiddenPriceResults = append(errHiddenPriceResults, model.GetHidePriceForCbscResult{
						QueryId: q.QueryId,
						Err: cerr.New("no available channel to calculate hidden fee",
							uint32(pb.Constant_CALCULATE_HIDDEN_FEE_NO_AVAILABLE_CHANNEL_ERROR)),
					})
					logging.GetLogger(ctx).Warn(fmt.Sprintf(
						"the shop is not 3pf, and no available channels for MT->MP, so return error, query=%+v", q))
				} else { // if no available channels and MP->MT, use 0 as hidden fee.
					skipHiddenPriceQueries = append(skipHiddenPriceQueries, q)
					logging.GetLogger(ctx).Debug(fmt.Sprintf(
						"the shop is not 3pf, and no available channels for MP->MT, so set hidden fee as 0, query=%+v", q))
				}
			} else { // if has available channels, add into sls query directly
				slsHiddenPriceQueries = append(slsHiddenPriceQueries, q)
				logging.GetLogger(ctx).Debug(fmt.Sprintf(
					"the shop is not 3pf, and has available channels, so directly add into sls query, query=%+v", q))
			}

			continue
		}

		// for 3pf shop
		// if cb channels exists, then calculate for these cb channels by sls.
		// otherwise, use 0 as hidden fee.
		skipHiddenPrice := true
		enabledChannelIds := make([]uint32, 0)
		if len(regionChannelMap[q.Region]) > 0 {
			for _, channelId := range q.EnabledChannelIdList {
				if channelInfo, exist := regionChannelMap[q.Region][uint64(channelId)]; exist {
					isLocalChannel := logicutil.IsChannelLocal(channelInfo.Tag)
					if !isLocalChannel {
						enabledChannelIds = append(enabledChannelIds, channelId)
					} else {
						logging.GetLogger(ctx).Debug(fmt.Sprintf("ignore due to local channel, under 3pf shop. channelId=%d", channelId))
					}
				} else {
					logging.GetLogger(ctx).Warn(fmt.Sprintf("cannot find this channel info, "+
						"channelId=%d, region=%s", channelId, q.Region))
				}
			}
		}
		if len(enabledChannelIds) > 0 {
			skipHiddenPrice = false
			q.EnabledChannelIdList = enabledChannelIds
		}

		if skipHiddenPrice {
			logging.GetLogger(ctx).Debug(fmt.Sprintf(
				"the shop is 3pf, but no cb channels, so set hidden fee as 0, query=%+v", q))
			skipHiddenPriceQueries = append(skipHiddenPriceQueries, q)
		} else {
			logging.GetLogger(ctx).Debug(fmt.Sprintf(
				"the shop is 3pf, and add into sls query for cb channels only, query=%+v", q))
			slsHiddenPriceQueries = append(slsHiddenPriceQueries, q)
		}
	}

	slsHiddenPriceResults, err := c.getHidePriceForCbscFromSls(ctx, slsHiddenPriceQueries)
	if err != nil {
		return nil, err
	}

	skipHiddenPriceResults := make([]model.GetHidePriceForCbscResult, len(skipHiddenPriceQueries))
	for idx, q := range skipHiddenPriceQueries {
		skipHiddenPriceResults[idx] = model.GetHidePriceForCbscResult{
			QueryId:   q.QueryId,
			HidePrice: float64(0),
		}
	}

	result := make([]model.GetHidePriceForCbscResult, 0)
	result = append(result, slsHiddenPriceResults...)
	result = append(result, skipHiddenPriceResults...)
	result = append(result, errHiddenPriceResults...)
	sort.Slice(result, func(i, j int) bool {
		return result[i].QueryId < result[j].QueryId
	})

	return result, nil
}

func (c *CalculationFactorsRepoImpl) getHidePriceForCbscFromSls(ctx context.Context, queries []model.GetHidePriceForCbscRequest) ([]model.GetHidePriceForCbscResult, error) {
	if len(queries) == 0 {
		return nil, nil
	}
	finalResults := make([]model.GetHidePriceForCbscResult, len(queries))

	hiddenFeeQueries := make([]*model.CalcHiddenFee, 0, len(queries))
	for _, query := range queries {
		weightInGram := calcutil.DbWeightToGram(int64(query.Weight))
		hiddenFeeQueries = append(hiddenFeeQueries, &model.CalcHiddenFee{
			ProductIDs: convutil.Uint32sToInt64s(query.EnabledChannelIdList),
			Region:     query.Region,
			ShopID:     query.ShopId,
			SkuInfos: []*model.SkuInfo{
				{
					ItemID:     query.ItemId,
					CategoryID: query.LeafCategoryId,
					Weight:     weightInGram,
					Quantity:   1,
				},
			},
			Direction:    1,
			WmsFlag:      0,
			DgFlag:       0,
			FallbackFlag: 0,
		})
	}

	hiddenFeeQueriesGroupByRegion, rawIndexes := model.GroupHiddenFeeQueriesByRegion(hiddenFeeQueries)

	for region, queryList := range hiddenFeeQueriesGroupByRegion {
		regionDomainUrl := config.GetShopeeRegionDomainUrl(region)
		ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, region)
		if err != nil {
			return nil, cerr.New(fmt.Sprintf("failed to fill CID, err=%s", err.Error()),
				uint32(pb.Constant_ERROR_INTERNAL))
		}

		req := model.BatchCalcHiddenFeeRequest{
			List: queryList,
		}

		resp, err := http.BatchCalcHiddenFee(c.httpCli, ctxWithCID, regionDomainUrl, &req)
		if err != nil {
			for i := range resp.Data {
				rawIndex := rawIndexes[region][i]
				finalResults[rawIndex] = model.GetHidePriceForCbscResult{
					QueryId: queries[rawIndex].QueryId,
					Err:     cerr.New(err.Error(), uint32(pb.Constant_CALCULATE_HIDDEN_FEE_ERROR)),
				}
			}
			continue
		}

		if resp.RetCode != 0 {
			errMsg := fmt.Sprintf("failed to calculate hidden fee, "+
				"code=%d, msg=%s",
				resp.RetCode, resp.Message)
			logging.GetLogger(ctx).Error(errMsg)

			for i := range resp.Data {
				rawIndex := rawIndexes[region][i]
				finalResults[rawIndex] = model.GetHidePriceForCbscResult{
					QueryId: queries[rawIndex].QueryId,
					Err:     cerr.New(errMsg, uint32(pb.Constant_CALCULATE_HIDDEN_FEE_ERROR)),
				}
			}
			continue
		}

		for i, channelResults := range resp.Data {
			rawIndex := rawIndexes[region][i]
			ignoreChannelErr := queries[rawIndex].IgnoreChannelErr
			var maxHiddenFee float64
			var iErr error

			for _, result := range channelResults {
				if result.RetCode != 0 {
					errMsg := fmt.Sprintf("failed to calculate hidden fee, "+
						"channelId=%v, code=%d, msg=%s, ignoreChannelErr=%v",
						result.ProductID, result.RetCode, result.Message, ignoreChannelErr)
					logging.GetLogger(ctx).Error(errMsg)

					if ignoreChannelErr {
						continue
					} else {
						iErr = cerr.New(errMsg, uint32(pb.Constant_CALCULATE_HIDDEN_FEE_ERROR))
						break
					}
				}

				if result.HiddenFee > maxHiddenFee {
					maxHiddenFee = result.HiddenFee
				}
			}

			if iErr != nil {
				finalResults[rawIndex] = model.GetHidePriceForCbscResult{
					QueryId: queries[rawIndex].QueryId,
					Err:     iErr,
				}
			} else {
				finalResults[rawIndex] = model.GetHidePriceForCbscResult{
					QueryId:   queries[rawIndex].QueryId,
					HidePrice: maxHiddenFee,
				}
			}
		}
	}

	return finalResults, nil
}

func (c *CalculationFactorsRepoImpl) GetOrderMartExchangeRate(ctx context.Context, srcCurrency string, dstCurrency string) (float64, error) {
	if srcCurrency == dstCurrency {
		return 1, nil
	}

	srcExchangeRateInfo, err := c.cache.GetOrderMartExchangeRate(ctx, constant.GetOrderMartExchangeRateCacheKey(srcCurrency))
	if err != nil {
		return 0, cerr.New(fmt.Sprintf("failed to get exchange rate for srcCurrency=%s", srcCurrency), uint32(pb.Constant_ERROR_CACHE))
	}

	dstExchangeRateInfo, err := c.cache.GetOrderMartExchangeRate(ctx, constant.GetOrderMartExchangeRateCacheKey(dstCurrency))
	if err != nil {
		return 0, cerr.New(fmt.Sprintf("failed to get exchange rate for dstCurrency=%s", dstCurrency), uint32(pb.Constant_ERROR_CACHE))
	}

	return dstExchangeRateInfo.ExchangeRate / srcExchangeRateInfo.ExchangeRate, nil
}
