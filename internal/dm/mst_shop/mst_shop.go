package mst_shop

import (
	"context"
	"fmt"
	"time"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type MSTShopDM interface {
	GetPShopInfoWithCache(ctx context.Context, pShopId uint64) (*sip_db.MstShop, error)
	GetPShopInfo(ctx context.Context, pShopId uint64) (*sip_db.MstShop, error)
	GetPShopInfoBatch(ctx context.Context, pShopIds []uint64) ([]*sip_db.MstShop, error)
}

type mstShopDMImpl struct {
	cache cache.CommonCache
	db    db.MSTShopDB
}

func NewMSTShopDM(cache cache.CommonCache, db db.MSTShopDB) MSTShopDM {
	return &mstShopDMImpl{
		cache: cache,
		db:    db,
	}
}

func (m *mstShopDMImpl) GetPShopInfoWithCache(ctx context.Context, pShopId uint64) (*sip_db.MstShop, error) {
	cacheKey := constant.GetMstShopCacheKey(pShopId)
	mstShop, cacheErr := m.cache.GetMstShop(ctx, cacheKey)
	if cacheErr != nil {
		logging.GetLogger(ctx).Warn(fmt.Sprintf("failed to get mstShop from cache, key=%v", cacheKey))
	}
	var err error
	if mstShop == nil {
		mstShop, err = m.db.GetByPShopId(ctx, pShopId)
		if err != nil {
			return nil, err
		}
	}

	if mstShop == nil {
		return nil, cerr.New(fmt.Sprintf("cannot find pShopInfo for pShopId=%v", pShopId), uint32(pb.Constant_ERROR_NOT_FOUND))
	}

	err = m.cache.Set(ctx, cacheKey, mstShop, time.Duration(config.GetCommonConfig().MstShopInfoExpireSeconds))
	if err != nil {
		logging.GetLogger(ctx).Warn(fmt.Sprintf("failed to set mstShop to cache, key=%v", cacheKey))
	}
	return mstShop, nil
}

func (m *mstShopDMImpl) GetPShopInfo(ctx context.Context, pShopId uint64) (*sip_db.MstShop, error) {
	mstShop, err := m.db.GetByPShopId(ctx, pShopId)
	if err != nil {
		return nil, err
	}
	return mstShop, nil
}

func (m *mstShopDMImpl) GetPShopInfoBatch(ctx context.Context, pShopIds []uint64) ([]*sip_db.MstShop, error) {
	mstShops, err := m.db.GetByShopIdBatch(ctx, pShopIds)
	if err != nil {
		return nil, err
	}
	return mstShops, nil
}
