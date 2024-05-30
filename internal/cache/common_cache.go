package cache

import (
	"context"
	"fmt"
	"time"

	"git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_constraints.pb"
	internal_sip "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type CommonCache interface {
	Set(ctx context.Context, key string, value interface{}, expire time.Duration) error
	SetWithLocal(ctx context.Context, key string, value interface{}, expire, localExpire time.Duration) error

	Del(ctx context.Context, key string) error
	DelWithLocal(ctx context.Context, key string) error

	GetString(ctx context.Context, key string) (string, error)

	GetInt32sWithLocal(ctx context.Context, keys []string) (vals []int32, found []bool, err error)
	SetInt32sWithLocal(ctx context.Context, kvs map[string]int32, localExpire time.Duration, remoteExpire time.Duration) error

	GetReferenceServiceFeeRate(ctx context.Context, key string) (uint64, error)
	GetProfitRateLimitList(ctx context.Context, key string) ([]*internal.MerchantConstraints, error)
	GetMerchantShopListWithLocal(ctx context.Context, key string, localExpire time.Duration) ([]uint64, error)
	GetAllCnscShopsWithLocal(ctx context.Context, key string, localExpire time.Duration) ([]*model.CNSCShop, error)
	GetShopFeeString(ctx context.Context, key string) (string, error)
	GetCountryMargin(ctx context.Context, key string) (string, error)
	GetAShopData(ctx context.Context, key string) ([]*internal_sip.AShopData, error)
	GetMstShop(ctx context.Context, key string) (*sip_db.MstShop, error)
	GetRegionLevelChannelInfoMap(ctx context.Context, key string) (map[uint64]*model.ChannelInfo, error)

	GetOrderMartExchangeRate(ctx context.Context, key string) (*model.OrderMartExchangeRate, error)
	SetOrderMartExchangeRateBatch(ctx context.Context, dataList []*model.OrderMartExchangeRate) bool
}

type CommonCacheImpl struct {
	remoteStore cache.Cache
	localCache  cache.Cache
}

func NewCommonCacheImpl(remoteStore config.RedisSession, localStore config.LocalCacheSession) *CommonCacheImpl {
	return &CommonCacheImpl{
		remoteStore: remoteStore,
		localCache:  localStore,
	}
}

func (c *CommonCacheImpl) GetMstShop(ctx context.Context, key string) (*sip_db.MstShop, error) {
	var receiver *sip_db.MstShop
	err := c.Get(ctx, key, &receiver)
	return receiver, err
}

func (c *CommonCacheImpl) GetAShopData(ctx context.Context, key string) ([]*internal_sip.AShopData, error) {
	var receiver []*internal_sip.AShopData
	err := c.Get(ctx, key, &receiver)
	return receiver, err
}

func (c *CommonCacheImpl) GetInt32sWithLocal(ctx context.Context, keys []string) (vals []int32, founds []bool, err error) {
	receivers := make(map[string]interface{})
	vals = make([]int32, len(keys))
	founds = make([]bool, len(keys))
	for _, key := range keys {
		receivers[key] = new(int32)
	}
	err = c.localCache.GetMany(ctx, receivers)

	if err == nil {
		for i, key := range keys {
			if v, exist := receivers[key]; exist {
				intVal, ok := v.(*int32)
				if ok && intVal != nil {
					founds[i] = true
					vals[i] = *intVal
				}
			}
		}
	}

	localMissKeys := make([]string, 0)
	originalIndexes := make([]int, 0)
	for i, found := range founds {
		if !found {
			localMissKeys = append(localMissKeys, keys[i])
			originalIndexes = append(originalIndexes, i)
		}
	}

	if len(localMissKeys) == 0 {
		return vals, founds, nil
	}

	remoteReceiver := make(map[string]interface{})
	for _, key := range localMissKeys {
		remoteReceiver[key] = new(int32)
	}
	err = c.remoteStore.GetMany(ctx, remoteReceiver)
	if err != nil {
		return vals, founds, err
	}
	for i, key := range localMissKeys {
		if v, exist := remoteReceiver[key]; exist {
			intVal, ok := v.(*int32)
			if ok && intVal != nil {
				founds[originalIndexes[i]] = true
				vals[originalIndexes[i]] = *intVal
			}
		}
	}
	return vals, founds, nil
}

func (c *CommonCacheImpl) SetInt32sWithLocal(ctx context.Context, kvs map[string]int32, localExpire time.Duration, remoteExpire time.Duration) error {
	storeKvs := make(map[string]interface{})
	for k, v := range kvs {
		storeKvs[k] = v
	}
	err := c.localCache.SetMany(ctx, storeKvs, localExpire)
	if err != nil {
		return err
	}
	err = c.remoteStore.SetMany(ctx, storeKvs, remoteExpire)
	if err != nil {
		return err
	}

	return nil
}

func (c *CommonCacheImpl) GetString(ctx context.Context, key string) (string, error) {
	var val string
	err := c.Get(ctx, key, &val)
	if err != nil {
		return "", err
	}

	return val, nil
}

func (c *CommonCacheImpl) GetCountryMargin(ctx context.Context, key string) (string, error) {
	var val string
	err := c.Get(ctx, key, &val)
	if err != nil {
		return "", err
	}

	return val, nil
}

func (c *CommonCacheImpl) GetShopFeeString(ctx context.Context, key string) (string, error) {
	var res string
	err := c.Get(ctx, key, &res)
	if err != nil {
		return res, err
	}
	return res, nil
}

func (c *CommonCacheImpl) GetAllCnscShopsWithLocal(ctx context.Context, key string, localExpire time.Duration) ([]*model.CNSCShop, error) {
	var shopList []*model.CNSCShop
	err := c.GetWithLocal(ctx, key, &shopList, localExpire)
	if err != nil {
		return nil, err
	}
	return shopList, nil
}

func (c *CommonCacheImpl) GetMerchantShopListWithLocal(ctx context.Context, key string, localExpire time.Duration) ([]uint64, error) {
	var shopIdList []uint64
	err := c.GetWithLocal(ctx, key, &shopIdList, localExpire)
	if err != nil {
		return nil, err
	}
	return shopIdList, nil
}

func (c *CommonCacheImpl) GetProfitRateLimitList(ctx context.Context, key string) ([]*internal.MerchantConstraints, error) {
	var profitRateLimits []*internal.MerchantConstraints
	err := c.Get(ctx, key, &profitRateLimits)
	if err != nil {
		if err != cache.ErrCacheMiss {
			logging.GetLogger(ctx).Error(fmt.Sprintf("error getting profit rate limit from cache, err=%v", err))
		}
		return nil, err
	}
	return profitRateLimits, nil
}

func (c *CommonCacheImpl) GetReferenceServiceFeeRate(ctx context.Context, key string) (uint64, error) {
	var refServiceFee uint64
	err := c.Get(ctx, key, &refServiceFee)
	if err != nil {
		return 0, err
	}
	return refServiceFee, nil
}

func (c *CommonCacheImpl) GetRegionLevelChannelInfoMap(ctx context.Context, key string) (map[uint64]*model.ChannelInfo, error) {
	var channelInfoMap map[uint64]*model.ChannelInfo
	err := c.Get(ctx, key, &channelInfoMap)
	if err != nil {
		if err != cache.ErrCacheMiss {
			logging.GetLogger(ctx).Warn(fmt.Sprintf("failed to get region level channel info from cache, key=%v, err=%s", key, err.Error()))
		}
		return nil, err
	}

	logging.GetLogger(ctx).Debug(fmt.Sprintf("success to get region level channel info from cache, "+
		"key=%v, val=%+v", key, cutil.JSONEncode(channelInfoMap)))
	return channelInfoMap, nil
}

func (c *CommonCacheImpl) GetOrderMartExchangeRate(ctx context.Context, key string) (*model.OrderMartExchangeRate, error) {
	var orderMartExchangeRate *model.OrderMartExchangeRate
	err := c.Get(ctx, key, &orderMartExchangeRate)
	if err != nil {
		if err != cache.ErrCacheMiss {
			logging.GetLogger(ctx).Error(fmt.Sprintf("failed to get order mart exchange rate from cache, key=%v, err=%s", key, err.Error()))
		} else {
			logging.GetLogger(ctx).Warn(fmt.Sprintf("cache miss for get order mart exchange rate, key=%s", key))
		}
		return nil, err
	}

	logging.GetLogger(ctx).Info(fmt.Sprintf("success to fetch order mart exchange rate from cache, "+
		"key=%s, value=%v", key, cutil.JSONEncode(orderMartExchangeRate)))
	return orderMartExchangeRate, nil
}

func (c *CommonCacheImpl) SetOrderMartExchangeRateBatch(ctx context.Context, dataList []*model.OrderMartExchangeRate) bool {
	success := false
	for _, data := range dataList {
		key := constant.GetOrderMartExchangeRateCacheKey(data.Currency)
		err := c.Set(ctx, key, data, cache.NoExpiration)
		if err != nil {
			logging.GetLogger(ctx).Error(fmt.Sprintf("failed to set order mart exchange rate into cache, "+
				"key=%v, val=%s, err=%s", key, cutil.JSONEncode(data), err.Error()))
			continue
		} else {
			logging.GetLogger(ctx).Info(fmt.Sprintf("success to set order mart exchange rate into cache, "+
				"key=%s, value=%v", key, cutil.JSONEncode(data)))
			success = true
		}
	}
	return success
}

func (c *CommonCacheImpl) Get(ctx context.Context, key string, receiver interface{}) error {
	return c.get(ctx, c.remoteStore, key, receiver)
}

func (c *CommonCacheImpl) GetWithLocal(ctx context.Context, key string, receiver interface{}, localExpire time.Duration) error {
	err := c.get(ctx, c.localCache, key, receiver)
	if err == nil {
		return nil
	}

	err = c.get(ctx, c.remoteStore, key, receiver)
	if err == nil {
		_ = c.set(ctx, c.localCache, key, receiver, localExpire)
	}

	return err
}

func (c *CommonCacheImpl) SetWithLocal(ctx context.Context, key string, value interface{}, expire, localExpire time.Duration) error {
	_ = c.set(ctx, c.localCache, key, value, localExpire)
	return c.set(ctx, c.remoteStore, key, value, expire)
}

func (c *CommonCacheImpl) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	err := c.set(ctx, c.remoteStore, key, value, expire)
	if err != nil {
		errMsg := fmt.Sprintf("falied to set cache value, err=%v, key=%s, value=%+v",
			err.Error(), key, value)
		logging.GetLogger(ctx).Error(errMsg)
		return err
	}

	return nil
}

func (c *CommonCacheImpl) DelWithLocal(ctx context.Context, key string) error {
	_ = c.del(ctx, c.localCache, key)
	return c.del(ctx, c.remoteStore, key)
}

func (c *CommonCacheImpl) Del(ctx context.Context, key string) error {
	err := c.del(ctx, c.remoteStore, key)
	if err != nil {
		errMsg := fmt.Sprintf("falied to delete cache value, err=%v, key=%s",
			err.Error(), key)
		logging.GetLogger(ctx).Error(errMsg)
		return err
	}

	return nil
}

func (c *CommonCacheImpl) get(ctx context.Context, storage cache.Cache, key string, receiver interface{}) error {
	err := storage.Get(ctx, key, receiver)
	if err != nil {
		return err
	}

	return nil
}

func (c *CommonCacheImpl) set(ctx context.Context, storage cache.Cache, key string, value interface{}, expire time.Duration) error {
	err := storage.Set(ctx, key, value, expire)
	if err != nil {
		return err
	}

	return nil
}

func (c *CommonCacheImpl) del(ctx context.Context, storage cache.Cache, key string) error {
	err := storage.Delete(ctx, key)
	if err != nil {
		return err
	}

	return nil
}
