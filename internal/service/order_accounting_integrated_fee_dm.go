package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/threadpool"

	"github.com/golang/protobuf/proto"

	commonCache "git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	moaif "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/marketplace_order_accounting_integrated_fee.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/slice"
)

type OrderAccountIntegratedFeeServiceDm struct {
	commissionFeeCacheManager            cache.CommissionRateCacheManager
	orderAccountIntegratedFeeSpexService spex.MarketplaceOrderAccountingIntegratedFee
	cache                                cache.CommonCache
}

func NewOrderAccountIntegratedFeeService(commissionFeeCacheManager cache.CommissionRateCacheManager, orderAccountIntegratedFeeSpexService spex.MarketplaceOrderAccountingIntegratedFee, commonCache cache.CommonCache) OrderAccountIntegratedFeeService {
	return &OrderAccountIntegratedFeeServiceDm{
		commissionFeeCacheManager:            commissionFeeCacheManager,
		orderAccountIntegratedFeeSpexService: orderAccountIntegratedFeeSpexService,
		cache:                                commonCache,
	}
}

func (dm *OrderAccountIntegratedFeeServiceDm) GetShopCommissionRateMap(ctx context.Context,
	shopIdRegionList []*model.ShopIdRegion) map[uint64]*ShopCommissionRateInfo {
	shopCommissionRateInfoMap := make(map[uint64]*ShopCommissionRateInfo)
	t := time.Now()

	wg := &sync.WaitGroup{}
	lock := &sync.Mutex{}

	for _, shopIdRegion := range shopIdRegionList {
		if shopIdRegion == nil {
			continue
		}

		wg.Add(1)
		shopId := shopIdRegion.ShopId
		region := shopIdRegion.Region
		err := threadpool.GetThreadPool().Do(ctx, func(cctx context.Context) {
			defer wg.Done()

			commissionRate, err := dm.GetCommissionRate(ctx, shopId, region)
			if err != nil {
				errMsg := fmt.Sprintf("failed to get commision rate, shopId=%d, err=%s",
					shopId, err.Error())
				logging.GetLogger(ctx).Error(errMsg)

				lock.Lock()
				shopCommissionRateInfoMap[shopId] = &ShopCommissionRateInfo{
					Err: cerr.New(errMsg,
						uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_SHOP_COMMISSION_RATE)),
				}
				lock.Unlock()

				return
			}

			lock.Lock()
			shopCommissionRateInfoMap[shopId] = &ShopCommissionRateInfo{
				CommissionRate: commissionRate,
			}
			lock.Unlock()
		})
		if err != nil {
			wg.Done()
			logging.GetLogger(ctx).Error("submit task to ThreadPool failed", ulog.Error(err))
		}
	}

	wg.Wait()

	logging.GetLogger(ctx).Info("GetShopCommissionRateMap finish", ulog.Float64("cost", time.Now().Sub(t).Seconds()))

	return shopCommissionRateInfoMap
}

func (dm *OrderAccountIntegratedFeeServiceDm) GetCommissionRate(ctx context.Context, shopId uint64, region string) (uint64, error) {
	// fetch from cache first
	key := dm.commissionFeeCacheManager.Key(shopId)
	commissionFee, err := dm.commissionFeeCacheManager.Get(ctx, key)
	if err == nil {
		return commissionFee, nil
	}
	if err != commonCache.ErrCacheMiss {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to get commision rate from cache, key=%s, err=%s", key, err.Error()))
	}

	// cache miss/failed to fetch from cache, then fetch from api, and set cache back
	ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, region)
	resp, err := dm.orderAccountIntegratedFeeSpexService.GetAppliableRulesByShop(ctxWithCID, &moaif.GetAppliableRulesByShopRequest{
		ShopId:  proto.Int64(int64(shopId)),
		FeeType: proto.Uint32(uint32(moaif.Constant_FEE_TYPE_COMMISSION_FEE)),
	})
	if err != nil {
		return 0, err
	}

	commissionFee = 0 // default value is 0
	applicableRuleDetailSlice := resp.GetRuleDetail()
	if len(applicableRuleDetailSlice) != 0 {
		sort.Sort(AppliableRuleDetailSlice(applicableRuleDetailSlice))
		commissionFee = applicableRuleDetailSlice[0].GetFeeRate() / 10
	}

	_ = dm.commissionFeeCacheManager.Set(ctx, key, commissionFee)
	return commissionFee, nil
}

func (dm *OrderAccountIntegratedFeeServiceDm) GetReferenceServiceFeeRate(ctx context.Context, shopId uint64, region string) (uint64, error) {
	cacheKey := constant.GetReferenceServiceFeeRateCacheKey(shopId)
	refServiceFee, err := dm.cache.GetReferenceServiceFeeRate(ctx, cacheKey)
	if err == nil {
		return refServiceFee, nil
	}

	req := &moaif.GetAppliableRulesByShopRequest{
		ShopId:  proto.Int64(int64(shopId)),
		FeeType: proto.Uint32(uint32(moaif.Constant_FEE_TYPE_SERVICE_FEE)),
	}

	regionCtx, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return 0, err
	}
	resp, err := dm.orderAccountIntegratedFeeSpexService.GetAppliableRulesByShop(regionCtx, req)
	if err != nil {
		return 0, err
	}

	appliableRuleDetailSlice := resp.GetRuleDetail()
	var referenceServiceFeeRateSum uint64
	for _, appliableRuleDetail := range appliableRuleDetailSlice {
		referenceServiceFeeRateSum = referenceServiceFeeRateSum + appliableRuleDetail.GetFeeRate()
	}
	referenceServiceFeeRate := referenceServiceFeeRateSum / 10

	_ = dm.cache.Set(ctx, cacheKey, referenceServiceFeeRate, time.Duration(config.GetCommonConfig().ReferenceServiceFeeRateCacheExpireSecond)*time.Second)

	return referenceServiceFeeRate, nil
}

func (dm *OrderAccountIntegratedFeeServiceDm) GetShopFee(ctx context.Context, region string, shopId int64, feeType model.ShopFeeType) (float64, error) {
	cacheKey := fmt.Sprintf("integrate_fee_%d_%d", shopId, feeType)
	feeString, err := dm.cache.GetShopFeeString(ctx, cacheKey)
	if err == nil {
		feeValue, err := strconv.ParseInt(feeString, 10, 64)
		if err != nil {
			return 0, err
		}
		return calcutil.ToRealPect(int(feeValue)), nil
	}

	req := &moaif.GetAppliableRulesByShopRequest{
		ShopId:  proto.Int64(int64(shopId)),
		FeeType: proto.Uint32(uint32(feeType)),
	}

	regionCtx, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return 0, err
	}
	resp, err := dm.orderAccountIntegratedFeeSpexService.GetAppliableRulesByShop(regionCtx, req)
	if err != nil {
		return 0, err
	}

	var feeValue int64
	if feeType == model.ShopFeeTypeServiceFee {
		feeValue, err = dm.calcServiceFee(ctx, resp.GetRuleDetail())
	} else {
		feeValue, err = dm.calcCommissionFee(ctx, resp.GetRuleDetail())
	}
	if err != nil {
		return 0, cerr.New(fmt.Sprintf("failed to calc shop fee, shopId=%v, feeType=%v", shopId, feeType), uint32(priceSyncPriceCalculationPb.Constant_ERROR_NOT_FOUND))
	}

	_ = dm.cache.Set(ctx, cacheKey, strconv.FormatInt(feeValue, 10), time.Duration(config.GetCommonConfig().CbSipShopFeeCacheExpireSeconds)*time.Second)
	return calcutil.ToRealPect(int(feeValue)), nil
}

func (dm *OrderAccountIntegratedFeeServiceDm) calcServiceFee(ctx context.Context, rules []*moaif.AppliableRuleDetail) (int64, error) {
	var feeRate uint64
	var exist bool
	for _, rule := range rules {
		if slice.ContainsUint32(rule.GetSettlementTypes(), uint32(moaif.Constant_SETTLEMENT_TYPE_SALES_ESCROW)) {
			exist = true
			feeRate += rule.GetFeeRate()
		}
	}

	if !exist {
		return 0, nil
	}

	return int64(feeRate), nil
}

func (dm *OrderAccountIntegratedFeeServiceDm) calcCommissionFee(ctx context.Context, rules []*moaif.AppliableRuleDetail) (int64, error) {
	var maxGroupId uint32
	var feeRate uint64
	for _, rule := range rules {
		if slice.ContainsUint32(rule.GetSettlementTypes(), uint32(moaif.Constant_SETTLEMENT_TYPE_SALES_ESCROW)) {
			if rule.GetGroupId() > maxGroupId {
				maxGroupId = rule.GetGroupId()
				feeRate = rule.GetFeeRate()
			}
		}
	}
	if maxGroupId == 0 {
		return 0, nil
	}

	return int64(feeRate), nil
}
