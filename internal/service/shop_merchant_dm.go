package service

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	shopMerchantPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/shop_merchant.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type ShopMerchantServiceDm struct {
	shopMerchantSpexService spex.ShopMerchant
}

func NewShopMerchantService(shopMerchant spex.ShopMerchant) ShopMerchantService {
	return &ShopMerchantServiceDm{
		shopMerchantSpexService: shopMerchant,
	}
}

func (dm *ShopMerchantServiceDm) CheckMerchantShopCbsc(ctx context.Context, shopId uint64) (bool, error) {
	resp, err := dm.shopMerchantSpexService.CheckMerchantShopCbsc(ctx, &shopMerchantPb.CheckMerchantShopCbscRequest{
		Shopid: proto.Uint32(uint32(shopId)),
	})
	if err != nil {
		return false, cerr.Wrap(err, "failed to check merchant shop cbsc", uint32(pb.Constant_ERROR_EXTERNAL))
	}
	return resp.GetCbscResult().GetIsCbsc(), nil
}

func (dm *ShopMerchantServiceDm) GetMerchantRegion(ctx context.Context, merchantId uint64) (string, error) {
	globalCtx, _ := cidutil.FillCtxWithNewCID(ctx, cidutil.GlobalCID)
	resp, err := dm.shopMerchantSpexService.GetMerchant(globalCtx, &shopMerchantPb.GetMerchantRequest{
		MerchantId: proto.Int64(int64(merchantId)),
	})
	if err != nil {
		return "", cerr.New(fmt.Sprintf("%v: failed to get merchant region, merchantId=%v", err, merchantId), uint32(pb.Constant_ERROR_GET_MERCHANT_REGION))
	}
	return resp.GetMerchant().GetRegion(), nil
}

func (dm *ShopMerchantServiceDm) GetMerchantRegionInfoMap(ctx context.Context, merchantIdList []uint64) map[uint64]*MerchantRegionInfo {
	merchantRegionMap := dm.getMerchantRegionMap(ctx, merchantIdList)

	merchantRegionInfoMap := make(map[uint64]*MerchantRegionInfo)
	for _, merchantId := range merchantIdList {
		merchantRegion := merchantRegionMap[merchantId]

		if merchantRegion == "" { // fail to get
			merchantRegionInfoMap[merchantId] = &MerchantRegionInfo{
				Err: cerr.New("failed to get the merchant region",
					uint32(pb.Constant_ERROR_GET_MERCHANT_REGION)),
				MerchantRegion: "",
			}
		} else { // success to get
			merchantRegionInfoMap[merchantId] = &MerchantRegionInfo{
				MerchantRegion: merchantRegion,
			}
		}
	}

	return merchantRegionInfoMap
}

func (dm *ShopMerchantServiceDm) getMerchantRegionMap(ctx context.Context, merchantIdList []uint64) map[uint64]string {
	merchantIdListReq := make([]int64, len(merchantIdList))
	for idx, m := range merchantIdList {
		merchantIdListReq[idx] = int64(m)
	}

	merchantInfoMap := dm.GetMerchantInfoMap(ctx, merchantIdListReq)

	merchantRegionMap := make(map[uint64]string)
	for _, m := range merchantInfoMap {
		merchantRegionMap[uint64(m.GetMerchantId())] = m.GetRegion()
	}

	return merchantRegionMap
}

func (dm *ShopMerchantServiceDm) GetMerchantInfoMap(ctx context.Context, merchantIdList []int64) map[int64]*shopMerchantPb.Merchant {
	if len(merchantIdList) == 0 {
		return nil
	}

	batchSize := config.GetBatchConfig().MaxBatchSizeForShopGetMerchantList
	merchantInfoMap := make(map[int64]*shopMerchantPb.Merchant)
	for start := 0; start < len(merchantIdList); {
		end := start + int(batchSize)
		if end > len(merchantIdList) {
			end = len(merchantIdList)
		}

		req := &shopMerchantPb.GetMerchantListRequest{
			MerchantIdList: merchantIdList[start:end],
		}

		// if not found, the api will return nil for that element
		ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, cidutil.GlobalCID)
		resp, err := dm.shopMerchantSpexService.GetMerchantList(ctxWithCID, req)
		if err != nil {
			logging.GetLogger(ctx).Error(
				fmt.Sprintf("failed to GetMerchantList, req=%s, err=%s",
					cutil.JSONEncode(req), err.Error()),
			)
		}

		for _, m := range resp.GetMerchantList() {
			merchantInfoMap[m.GetMerchantId()] = m
		}

		start += int(batchSize)
	}

	return merchantInfoMap
}

func (dm *ShopMerchantServiceDm) GetMerchantInfoByShopId(ctx context.Context, shopId uint64) (*shopMerchantPb.Merchant, error) {
	if shopId == 0 {
		return nil, cerr.New("invalid shop id 0", uint32(pb.Constant_ERROR_PARAMS))
	}

	req := &shopMerchantPb.GetMerchantRequest{
		Shopid:         proto.Uint32(uint32(shopId)),
		UpgradedStatus: proto.Int32(int32(shopMerchantPb.Constant_UPGRADED_STATUS_ALL)),
	}

	ctx, _ = cidutil.FillCtxWithNewCID(ctx, cidutil.GlobalCID)

	resp, err := dm.shopMerchantSpexService.GetMerchant(ctx, req)
	if err != nil {
		return nil, cerr.New(err.Error(), uint32(pb.Constant_ERROR_EXTERNAL))
	}

	return resp.GetMerchant(), nil
}
