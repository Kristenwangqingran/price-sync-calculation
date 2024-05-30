package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/shop_core.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/convutil"
)

type ShopCoreService interface {
	//IsOnboardedAShop returns true iif given shopId is AShop and is not offboarded
	IsAShopAndNotOffboarded(ctx context.Context, aShopId uint64) (bool, error)
	AreOnboardedAShops(ctx context.Context, aShopIds []uint64) (map[uint64]bool, error)
	GetAShopInfo(ctx context.Context, aShopId uint64) (*shop_core.GetAshopInfoResponse, error)
	IsAShopOffboarded(resp *shop_core.GetAshopInfoResponse) bool
	FilterOnboardedAShops(ctx context.Context, aShopIds []uint64) ([]uint64, error)
	GetPShopIdByAShopId(ctx context.Context, aShopId uint64) (uint64, error)
	GetPShopIdsByAShopIdsBatch(ctx context.Context, aShopIds []uint64) (map[uint64]uint64, error)
	GetAShopIdsByPShopId(ctx context.Context, pShopId uint64) ([]uint64, error)
	GetShopDetail(ctx context.Context, shopId uint64, region string) (*model.ShopDetail, error)
	GetShopRegionByShopId(ctx context.Context, shopId uint64) (string, error)
	GetShopRegionByShopIdBatch(ctx context.Context, shopIds []uint64) ([]model.ShopIdRegion, error)
	IsSellerWarehouseShop(ctx context.Context, shopId uint64, region string) (isWarehouse bool, is3pf bool, err error)
	GetShopWarehouseByShopId(ctx context.Context, shopId uint64, region string) ([]*shop_core.ShopWarehouse, error)
}

type shopCoreServiceImpl struct {
	shopCoreDataCache cache.ShopDataCacheManager
	shopCoreProxy     spex.ShopCore
}

func NewShopCoreService(shopCoreProxy spex.ShopCore, shopCoreDataCache cache.ShopDataCacheManager) ShopCoreService {
	return &shopCoreServiceImpl{
		shopCoreDataCache: shopCoreDataCache,
		shopCoreProxy:     shopCoreProxy,
	}
}

func (dm *shopCoreServiceImpl) GetShopRegionByShopId(ctx context.Context, shopId uint64) (string, error) {
	req := &shop_core.BatchGetShopRegionsRequest{
		ShopIdList: []int64{int64(shopId)},
	}

	resp, err := dm.shopCoreProxy.BatchGetShopRegions(ctx, req)
	if err != nil {
		return "", err
	}

	// shopCoreProxy guard only 1 ShopRegionPairs inside resp
	return strings.ToUpper(resp.ShopRegionPairs[0].GetRegion()), nil
}

func (dm *shopCoreServiceImpl) GetShopRegionByShopIdBatch(ctx context.Context, shopIds []uint64) ([]model.ShopIdRegion, error) {
	shopIdList := make([]int64, 0, len(shopIds))
	for _, shopId := range shopIds {
		shopIdList = append(shopIdList, int64(shopId))
	}
	req := &shop_core.BatchGetShopRegionsRequest{
		ShopIdList: shopIdList,
	}

	resp, err := dm.shopCoreProxy.BatchGetShopRegions(ctx, req)
	if err != nil {
		return nil, err
	}

	shopIdRegionPairs := make([]model.ShopIdRegion, 0)
	for _, pair := range resp.GetShopRegionPairs() {
		shopIdRegionPairs = append(shopIdRegionPairs, model.ShopIdRegion{
			ShopId: uint64(pair.GetShopId()),
			Region: strings.ToUpper(pair.GetRegion()),
		})
	}

	return shopIdRegionPairs, nil
}

// IsSellerWarehouseShop check if the shop type.
// if shop doesn't exist, will return false.
// - multiple warehouse: can refer https://confluence.shopee.io/display/SCPM/%5BSeller+Management%5D+%5BPRD%5D+Multi-Warehouse+Phase+2
// - 3pf: can refer https://confluence.shopee.io/display/SCPM/%5BSPML-17188%5DSupport+Upload+local+WH+stock+for+CB+3PF+sellers
func (dm *shopCoreServiceImpl) IsSellerWarehouseShop(ctx context.Context, shopId uint64, region string) (isWarehouse bool, is3pf bool, err error) {
	req := &shop_core.IsSellerWarehouseShopRequest{
		ShopId: proto.Int64(int64(shopId)),
		Region: proto.String(region),
	}

	ctx, err = cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return false, false,
			cerr.Wrap(err, fmt.Sprintf("add region(%s) to context failed in IsSellerWarehouseShop", region),
				uint32(pb.Constant_ERROR_INTERNAL))
	}

	result, _ := dm.shopCoreDataCache.CheckIsSellerWarehouseShopFromLocalCache(ctx, int64(shopId))
	if result != nil {
		return result.IsWarehouse, result.Is3PFWarehouse, nil
	}

	resp, err := dm.shopCoreProxy.IsSellerWarehouseShop(ctx, req)
	if err != nil {
		return false, false, err
	}

	_ = dm.shopCoreDataCache.SetIsSellerWarehouseShopToLocalCache(ctx, int64(shopId), &cache.ShopSellerWarehouseInfo{
		IsWarehouse:    resp.GetIsWarehouse(),
		Is3PFWarehouse: resp.GetIs_3PfWarehouse(),
	})

	return resp.GetIsWarehouse(), resp.GetIs_3PfWarehouse(), nil
}

func (dm *shopCoreServiceImpl) GetShopWarehouseByShopId(ctx context.Context, shopId uint64, region string) ([]*shop_core.ShopWarehouse, error) {
	req := &shop_core.GetShopWarehouseByShopIdRequest{
		ShopId: proto.Int64(int64(shopId)),
		Region: proto.String(region),
	}

	ctx, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return nil, cerr.Wrap(err, fmt.Sprintf("add region(%s) to context failed in GetShopWarehouseByShopId", region),
			uint32(pb.Constant_ERROR_INTERNAL))
	}

	resp, err := dm.shopCoreProxy.GetShopWarehouseByShopId(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetWarehouses(), nil
}

func (dm *shopCoreServiceImpl) GetAShopInfo(ctx context.Context, aShopId uint64) (*shop_core.GetAshopInfoResponse, error) {
	resp, err := dm.shopCoreProxy.GetAShopInfo(ctx, int64(aShopId))
	if err != nil {
		ulog.DefaultLoggerFromContext(ctx).Error("error when shop core proxy get a shop info", ulog.Uint64("aShopId", aShopId), ulog.Error(err))
		return nil, err
	}
	if resp == nil {
		return nil, cerr.New(fmt.Sprintf("cannot find aShopInfo for aShopId=%v", aShopId), uint32(pb.Constant_ERROR_NOT_FOUND))
	}
	return resp, nil
}

func (dm *shopCoreServiceImpl) IsAShopAndNotOffboarded(ctx context.Context, aShopId uint64) (bool, error) {
	resp, err := dm.GetAShopInfo(ctx, aShopId)
	if err != nil {
		return false, err
	}
	if dm.IsAShopOffboarded(resp) {
		return false, nil
	}
	return true, nil
}

func (dm *shopCoreServiceImpl) AreOnboardedAShops(ctx context.Context, aShopIds []uint64) (map[uint64]bool, error) {
	res := make(map[uint64]bool)
	for _, aShopId := range aShopIds {
		resp, err := dm.GetAShopInfo(ctx, aShopId)
		if err != nil {
			return nil, err
		}
		if dm.IsAShopOffboarded(resp) {
			continue
		}
		res[aShopId] = true
	}
	return res, nil
}

func (dm *shopCoreServiceImpl) IsAShopOffboarded(resp *shop_core.GetAshopInfoResponse) bool {
	return resp == nil || resp.GetOffboardTime() != 0
}

func (dm *shopCoreServiceImpl) FilterOnboardedAShops(ctx context.Context, aShopIds []uint64) ([]uint64, error) {
	onboardedAShopSet, err := dm.AreOnboardedAShops(ctx, aShopIds)
	if err != nil {
		return nil, err
	}
	onboardedAShops := make([]uint64, 0, len(onboardedAShopSet))

	for aShopId := range onboardedAShopSet {
		onboardedAShops = append(onboardedAShops, aShopId)
	}
	return onboardedAShops, nil
}

func (dm *shopCoreServiceImpl) GetPShopIdByAShopId(ctx context.Context, aShopId uint64) (uint64, error) {
	pShopId, err := dm.shopCoreProxy.GetPShopIdByAShopId(ctx, aShopId)
	if err != nil {
		ulog.DefaultLoggerFromContext(ctx).Error("error shop core proxy get p shop id from a shop id", ulog.Error(err))
		return 0, err
	}
	return uint64(pShopId), nil
}

func (dm *shopCoreServiceImpl) GetPShopIdsByAShopIdsBatch(ctx context.Context, aShopIds []uint64) (map[uint64]uint64, error) {
	paMaps, err := dm.shopCoreProxy.GetPShopIdsByAShopIdsBatch(ctx, aShopIds)
	if err != nil {
		ulog.DefaultLoggerFromContext(ctx).Error("error shop core proxy batch get p shop ids from a shop ids", ulog.Error(err))
		return nil, err
	}
	aToPShopIdMap := make(map[uint64]uint64)
	for _, paRelation := range paMaps {
		aToPShopIdMap[uint64(paRelation.GetAshopId())] = uint64(paRelation.GetPshopId())
	}
	return aToPShopIdMap, nil
}

func (dm *shopCoreServiceImpl) GetAShopIdsByPShopId(ctx context.Context, pShopId uint64) ([]uint64, error) {
	rawAShopIds, err := dm.shopCoreProxy.GetAShopIdsByPShopId(ctx, pShopId)
	if err != nil {
		return nil, err
	}

	return convutil.Int64sToUint64s(rawAShopIds), nil
}

func (dm *shopCoreServiceImpl) GetShopDetail(ctx context.Context, shopId uint64, region string) (*model.ShopDetail, error) {
	cacheKey := dm.shopCoreDataCache.ShopDetailKey(int64(shopId))
	shopDetail, err := dm.shopCoreDataCache.GetShopDetail(ctx, cacheKey)
	if err == nil && shopDetail != nil {
		return shopDetail, nil
	}

	ctx, err = cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return nil, cerr.Wrap(err, fmt.Sprintf("add region(%s) to context failed", region), uint32(pb.Constant_ERROR_INTERNAL))
	}
	req := &shop_core.GetShopRequest{
		Shopid: proto.Int64(int64(shopId)),
	}

	resp, err := dm.shopCoreProxy.GetShop(ctx, req)
	if err != nil {
		return nil, err
	}

	shopDetail = &model.ShopDetail{
		ShopId:            resp.GetShop().GetShopid(),
		PickupAddressId:   resp.GetShop().GetPickupAddressId(),
		UserId:            resp.GetShop().GetUserid(),
		Name:              resp.GetShop().GetName(),
		Region:            resp.GetShop().GetRegion(),
		Status:            int(resp.GetShop().GetStatus()),
		Ctime:             int64(resp.GetShop().GetCtime()),
		Mtime:             int64(resp.GetShop().GetMtime()),
		CbOption:          int(resp.GetShop().GetCbOption()),
		Cover:             resp.GetShop().GetCover(),
		Description:       resp.GetShop().GetDescription(),
		IsBbcSeller:       resp.GetShop().GetIsBbcSeller(),
		IsSipPrimary:      resp.GetShop().GetIsSipPrimary(),
		IsSipAffiliated:   resp.GetShop().GetIsSipAffiliated(),
		IsSipCb:           resp.GetShop().GetIsSipCb(),
		Covers:            resp.GetShop().GetCovers(),
		UpdateShopCovers:  resp.GetShop().GetUpdateShopCovers(),
		ReturnAddressId:   resp.GetShop().GetReturnAddressId(),
		CbReturnAddressId: resp.GetShop().GetCbReturnAddressId(),
		SipPrimaryRegion:  resp.GetShop().GetSipPrimaryRegion(),
	}

	_ = dm.shopCoreDataCache.SetShopDetail(ctx, cacheKey, shopDetail)

	return shopDetail, nil
}
