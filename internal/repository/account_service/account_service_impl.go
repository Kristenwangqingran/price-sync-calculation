package account_service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/wire"

	commoncache "git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/http"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	accountCore "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/account_core.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/shop_core.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/shop_merchant.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/httpcliutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/threadpool"
)

var ProviderSet = wire.NewSet(
	NewAccountServiceImpl,
	wire.Struct(new(AccountServiceImplOpts), "*"),
	wire.Bind(new(AccountServiceRepo), new(*AccountServiceImpl)),
)

type AccountServiceImpl struct {
	cache                cache.CommonCache
	httpCli              httpcliutil.HTTPCli
	shopMerchantService  spex.ShopMerchant
	accountCoreService   spex.AccountCore
	shopCoreService      spex.ShopCore
	shopDataCacheManager cache.ShopDataCacheManager
}

type AccountServiceImplOpts struct {
	Cache                cache.CommonCache
	HttpCli              httpcliutil.HTTPCli
	ShopMerchantService  spex.ShopMerchant
	AccountCoreService   spex.AccountCore
	ShopCoreService      spex.ShopCore
	ShopDataCacheManager cache.ShopDataCacheManager
}

func NewAccountServiceImpl(deps *AccountServiceImplOpts) *AccountServiceImpl {
	return &AccountServiceImpl{
		cache:                deps.Cache,
		httpCli:              deps.HttpCli,
		shopMerchantService:  deps.ShopMerchantService,
		accountCoreService:   deps.AccountCoreService,
		shopCoreService:      deps.ShopCoreService,
		shopDataCacheManager: deps.ShopDataCacheManager,
	}
}

func (a *AccountServiceImpl) GetUserStatusMap(ctx context.Context, shopIdRegions []model.ShopIdRegion) (map[uint64]int32, error) {
	result := make(map[uint64]int32)
	if len(shopIdRegions) == 0 {
		return result, nil
	}

	keys := make([]string, 0, len(shopIdRegions))
	for _, shopIdRegion := range shopIdRegions {
		key := constant.WrapRedisKey("UserStatus", strconv.FormatUint(shopIdRegion.ShopId, 10))
		keys = append(keys, key)
	}

	values, founds, err := a.cache.GetInt32sWithLocal(ctx, keys)
	if err != nil {
		logging.GetLogger(ctx).Error("get user status map from cache failed", ulog.Error(err))
	}

	remainingShopIdRegions := make([]model.ShopIdRegion, 0)
	originalIndexesOfNotFound := make([]int, 0)
	for i, shopIdRegion := range shopIdRegions {
		// not found in cache, get from account service
		if !founds[i] {
			remainingShopIdRegions = append(remainingShopIdRegions, shopIdRegion)
			originalIndexesOfNotFound = append(originalIndexesOfNotFound, i)
			continue
		}

		// found in cache, set to result
		result[shopIdRegion.ShopId] = values[i]
	}

	shopIdRegionPairsGroupByRegion := model.GroupShopIdRegionPairsByRegion(remainingShopIdRegions)
	cacheMissedResult := make(map[string]int32)
	for region, pairs := range shopIdRegionPairsGroupByRegion {
		shopIds := make([]int64, 0)
		for _, pair := range pairs {
			shopIds = append(shopIds, int64(pair.ShopId))
		}

		regionalCtx, err := cidutil.FillCtxWithNewCID(ctx, region)
		if err != nil {
			return nil, err
		}

		userIds, err := a.GetUserIdByShopIdBatch(regionalCtx, shopIds)
		if err != nil {
			return nil, err
		}

		userList, err := a.GetAccountBatch(regionalCtx, userIds)
		if err != nil {
			return nil, cerr.Wrap(err, "failed to get account batch", uint32(pb.Constant_ERROR_EXTERNAL))
		}
		for _, userDetail := range userList {
			result[uint64(userDetail.GetShopid())] = userDetail.GetStatus()

			cacheKey := constant.WrapRedisKey("UserStatus", strconv.FormatUint(uint64(userDetail.GetShopid()), 10))
			cacheMissedResult[cacheKey] = userDetail.GetStatus()
		}
	}

	err = a.cache.SetInt32sWithLocal(ctx, cacheMissedResult, time.Duration(config.GetCommonConfig().UserStatusMapLocalCacheExpireSeconds)*time.Second, time.Duration(config.GetCommonConfig().UserStatusMapRemoteCacheExpireSeconds)*time.Second)
	if err != nil {
		logging.GetLogger(ctx).Error("set user status map from cache failed", ulog.Error(err))
	}

	return result, nil
}

// GetUserIdByShopIdBatch no req & resp order guarantee
func (a *AccountServiceImpl) GetUserIdByShopIdBatch(ctx context.Context, shopIds []int64) ([]int64, error) {
	result := make([]int64, 0)
	missedShopIds := make([]int64, 0)
	for _, shopId := range shopIds {
		userId, err := a.shopDataCacheManager.GetUserIdByShopIdFromLocalCache(ctx, shopId)
		if err != nil {
			if err == commoncache.ErrCacheMiss {
				missedShopIds = append(missedShopIds, shopId)
				continue
			} else {
				return nil, err
			}
		}

		// cache found
		result = append(result, userId)
	}

	spexResult, err := a.getUserIdByShopIdBatchFromSpex(ctx, missedShopIds)
	if err != nil {
		return nil, err
	}

	for shopId, userId := range spexResult {
		_ = a.shopDataCacheManager.SetUserIdOfShopIdToLocalCache(ctx, shopId, userId)
		result = append(result, userId)
	}

	return result, nil
}

func (a *AccountServiceImpl) getUserIdByShopIdBatchFromSpex(ctx context.Context, shopIds []int64) (map[int64]int64, error) {
	t := time.Now()
	result := make(map[int64]int64)
	var gErr error

	lock := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	batchSize := int(config.GetBatchConfig().MaxBatchSizeForGetUserIdByShopId)
	for i := 0; i < len(shopIds); i += batchSize {
		wg.Add(1)

		j := i + batchSize
		if j > len(shopIds) {
			j = len(shopIds)
		}

		subShopIds := shopIds[i:j]
		err := threadpool.GetThreadPool().Do(ctx, func(cctx context.Context) {
			defer wg.Done()

			r, err := a.doGetUserIdByShopIdBatch(ctx, subShopIds)
			if err != nil {
				lock.Lock()
				gErr = err
				lock.Unlock()
				return
			}

			if len(r) != len(subShopIds) {
				lock.Lock()
				gErr = cerr.New("shop.core.get_shop_userid_list req/resp length mismatch", uint32(pb.Constant_ERROR_EXTERNAL))
				lock.Unlock()
				return
			}

			lock.Lock()
			for idx, shopId := range subShopIds {
				result[shopId] = r[idx]
			}
			lock.Unlock()
		})

		if err != nil {
			wg.Done()
			logging.GetLogger(ctx).Error("GetUserIdByShopIdBatch: submit task to thread pool failed", ulog.Error(err))
			return nil, err
		}
	}

	if gErr != nil {
		return nil, gErr
	}

	wg.Wait()
	logging.GetLogger(ctx).Info("getUserIdByShopIdBatchFromSpex finish",
		ulog.Float64("cost", time.Now().Sub(t).Seconds()), ulog.Int("shop_ids_size", len(shopIds)))
	return result, nil
}

func (a *AccountServiceImpl) GetAccountBatch(ctx context.Context, userIds []int64) ([]*accountCore.UserDetail, error) {
	t := time.Now()
	result := make([]*accountCore.UserDetail, 0)
	var gErr error
	lock := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	batchSize := int(config.GetBatchConfig().MaxBatchSizeForGetAccount)
	for i := 0; i < len(userIds); i += batchSize {
		wg.Add(1)

		j := i + batchSize
		if j > len(userIds) {
			j = len(userIds)
		}

		subUserIds := userIds[i:j]

		err := threadpool.GetThreadPool().Do(ctx, func(cctx context.Context) {
			defer wg.Done()

			req := &accountCore.GetAccountBatchRequest{
				UseridList:  subUserIds,
				NeedDeleted: proto.Bool(true),
			}

			resp, err := a.accountCoreService.GetAccountBatch(ctx, req)
			if err != nil {
				lock.Lock()
				gErr = cerr.Wrap(err, "failed to get account batch", uint32(pb.Constant_ERROR_EXTERNAL))
				lock.Unlock()
				return
			}

			// order is not reliable
			lock.Lock()
			result = append(result, resp.GetUserList()...)
			lock.Unlock()
		})

		if err != nil {
			wg.Done()
			logging.GetLogger(ctx).Error("GetAccountBatch: submit task to thread pool failed", ulog.Error(err))
			return nil, err
		}
	}

	if gErr != nil {
		return nil, gErr
	}

	wg.Wait()

	logging.GetLogger(ctx).Info("GetAccountBatch finish",
		ulog.Float64("cost", time.Now().Sub(t).Seconds()), ulog.Int("user_ids_size", len(userIds)))
	return result, nil
}

func (a *AccountServiceImpl) doGetUserIdByShopIdBatch(ctx context.Context, shopIds []int64) ([]int64, error) {
	req := &shop_core.GetShopUseridListRequest{
		ShopidList: shopIds,
	}
	resp, err := a.shopCoreService.GetUserIdByShopIdBatch(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetUseridList(), nil
}

func (a *AccountServiceImpl) GetAllShopList(ctx context.Context, accountId *uint64, merchantId uint64) ([]*model.CNSCShop, error) {
	if accountId == nil || *accountId == 0 {
		return a.GetMerchantShopList(ctx, merchantId)
	}
	return a.GetAllCnscShops(ctx, *accountId, merchantId)
}

func (a *AccountServiceImpl) GetAllCnscShops(ctx context.Context, accountId uint64, merchantId uint64) ([]*model.CNSCShop, error) {
	cacheKey := constant.GetAllCnscShopsCacheKey(merchantId)
	shops, err := a.cache.GetAllCnscShopsWithLocal(ctx, cacheKey, time.Duration(config.GetCommonConfig().CnscShopsLocalCacheExpireSeconds)*time.Second)
	if err == nil {
		return shops, nil
	}

	resp := &model.GetCnscShopsByAccountIdResponse{}
	globalCtx, err := cidutil.FillCtxWithNewCID(ctx, cidutil.GlobalCID)
	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"account_id":  strconv.FormatUint(accountId, 10),
		"merchant_id": strconv.FormatUint(merchantId, 10),
	}
	httpResponse, err := http.GetSubAccountHttp(globalCtx, a.httpCli, constant.GetAllCnscShops, params)
	if err != nil || !httpResponse.IsOk() {
		return nil, cerr.New(fmt.Sprintf("failed to get cnsc shops, param=%+v, err=%v", params, err), uint32(pb.Constant_ERROR_HTTP_API))
	}
	body, err := httpResponse.Unmarshal()
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to unmarshal http response, param=%+v", params), uint32(pb.Constant_ERROR_MARSHAL))
	}

	if err = json.Unmarshal(body.([]byte), &resp); err != nil {
		return nil, cerr.Wrap(err, "failed to unmarshal GetCnscShopsByAccountIdResponse response", uint32(pb.Constant_ERROR_MARSHAL))
	}

	_ = a.cache.SetWithLocal(ctx, cacheKey, resp.Shops, time.Duration(config.GetCommonConfig().CnscShopsRemoteCacheExpireSeconds)*time.Second, time.Duration(config.GetCommonConfig().CnscShopsLocalCacheExpireSeconds)*time.Second)
	return resp.Shops, nil

}

func (a *AccountServiceImpl) GetMerchantShopList(ctx context.Context, merchantId uint64) ([]*model.CNSCShop, error) {
	cacheKey := constant.GetMerchantShopListCacheKey(merchantId)
	localCacheExpire := time.Duration(config.GetCommonConfig().MerchantShopListLocalCacheExpireSeconds) * time.Second
	shopIdList, err := a.cache.GetMerchantShopListWithLocal(ctx, cacheKey, localCacheExpire)
	if err == nil {
		return model.NewCnscShopListFromShopIdList(shopIdList), nil
	}

	ctx, _ = cidutil.FillCtxWithNewCID(ctx, cidutil.GlobalCID)

	req := &shop_merchant.GetShopIdListByMerchantRequest{
		MerchantId: proto.Int64(int64(merchantId)),
	}

	resp, err := a.shopMerchantService.GetShopIDListByMerchant(ctx, req)
	if err != nil {
		return nil, cerr.Wrap(err, "failed to get merchant shop", uint32(pb.Constant_ERROR_EXTERNAL))
	}

	allShopList := make([]uint64, 0)
	for _, shopId := range resp.GetShopIdList() {
		allShopList = append(allShopList, uint64(shopId))
	}

	checkResult, err := a.CheckCnscWhiteList(ctx, allShopList)
	if err != nil {
		return nil, err
	}

	for _, shopId := range allShopList {
		if checkResult[shopId] {
			shopIdList = append(shopIdList, shopId)
		}
	}

	_ = a.cache.SetWithLocal(ctx, cacheKey, shopIdList, time.Duration(config.GetCommonConfig().MerchantShopListRemoteCacheExpireSeconds)*time.Second, localCacheExpire)
	return model.NewCnscShopListFromShopIdList(shopIdList), nil
}

func (a *AccountServiceImpl) CheckCnscWhiteList(ctx context.Context, shopIdList []uint64) (map[uint64]bool, error) {
	batchSize := int(config.GetBatchConfig().MaxBatchSizeForCnscCheckWhiteList)
	res := make(map[uint64]bool)
	for i := 0; i < len(shopIdList); i += batchSize {
		upperBound := i + batchSize
		if upperBound > len(shopIdList) {
			upperBound = len(shopIdList)
		}

		curShopIdList := make([]uint32, 0)
		for _, s := range shopIdList[i:upperBound] {
			curShopIdList = append(curShopIdList, uint32(s))
		}

		singleRes, err := a.shopMerchantService.BatchCheckMerchantShopCbsc(ctx,
			&shop_merchant.BatchCheckMerchantShopCbscRequest{
				ShopidList: curShopIdList,
			})
		if err != nil {
			return nil, err
		}
		for _, v := range singleRes.GetCbscResultList() {
			res[uint64(v.GetShopid())] = v.GetIsCbsc()
		}
	}
	return res, nil
}
