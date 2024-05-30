package common_sip_logic

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/wire"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/a_item"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/a_shop"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/mst_shop"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
)

type CommonSIPLogicImpl struct {
	ashopDataDM         a_shop.AShopDataDM
	shopCoreService     service.ShopCoreService
	itemDiscountService service.ItemDiscountService
	aItemDataDM         a_item.AItemDataDM
	mstShopDM           mst_shop.MSTShopDM
}

func NewCommonSIPLogicImpl(
	aShopDataDM a_shop.AShopDataDM,
	shopCoreService service.ShopCoreService,
	aItemDataService a_item.AItemDataDM,
	mstShopDM mst_shop.MSTShopDM,
) *CommonSIPLogicImpl {
	return &CommonSIPLogicImpl{
		ashopDataDM:     aShopDataDM,
		shopCoreService: shopCoreService,
		aItemDataDM:     aItemDataService,
		mstShopDM:       mstShopDM,
	}
}

var ProviderSet = wire.NewSet(
	NewCommonSIPLogicImpl,
)

func (impl *CommonSIPLogicImpl) GetAShopMarginBatch(ctx context.Context, affiShopIds []uint64) (map[uint64]int64, error) {
	onboardedAShops, err := impl.shopCoreService.FilterOnboardedAShops(ctx, affiShopIds)
	if err != nil {
		return nil, err
	}
	shopMargins, err := impl.ashopDataDM.GetAShopMarginBatch(ctx, onboardedAShops)

	if err != nil {
		return nil, err
	}
	return shopMargins, nil
}

func (impl *CommonSIPLogicImpl) SetAShopMargin(ctx context.Context, affiShopId uint64, margin int64) error {
	aShopInfo, err := impl.shopCoreService.GetAShopInfo(ctx, affiShopId)
	if err != nil {
		return cerr.Wrap(err, fmt.Sprintf("a shop not found, aShopId=%d", affiShopId), uint32(pb.Constant_ERROR_NOT_FOUND))
	}
	if impl.shopCoreService.IsAShopOffboarded(aShopInfo) {
		return cerr.Wrap(fmt.Errorf("a shop is already offboarded, aShopId=%d", affiShopId), "", uint32(pb.Constant_ERROR_NOT_FOUND))
	}
	pShopId, err := impl.shopCoreService.GetPShopIdByAShopId(ctx, affiShopId)
	if err != nil {
		return err
	}
	pShopData, err := impl.mstShopDM.GetPShopInfoWithCache(ctx, pShopId)
	if err != nil {
		return err
	}
	if shopMarginValidationErr := impl.validateAShopMargin(margin, pShopData.IsCbShop()); shopMarginValidationErr != nil {
		return cerr.Wrap(err, "", uint32(pb.Constant_ERROR_PARAMS))
	}
	return impl.ashopDataDM.SetAShopDataShopMargin(ctx, affiShopId, pShopId, int32(margin))
}

func (impl *CommonSIPLogicImpl) validateAShopMargin(shopMargin int64, isPShopCB bool) error {
	realVal := calcutil.ToRealPect(int(shopMargin))
	if isPShopCB {
		if realVal < -1 || realVal > 2 {
			return fmt.Errorf("CB SIP Ashop margin should be >= -1 and <= 2, got=%d, isPShopCB=%v", shopMargin, isPShopCB)
		}
	} else {
		if realVal <= 0 || realVal > 10 {
			return fmt.Errorf("local SIP Ashop margin should be > 0 and <= 10, got=%d, isPShopCB=%v", shopMargin, isPShopCB)
		}
	}
	return nil
}

func (impl *CommonSIPLogicImpl) GetAShopPriceRatioBatch(ctx context.Context, affiShopIds []uint64) (map[uint64]int64, error) {
	//TODO urg: confirm if onboarded shop only satisfies listing side need
	onboardedAShops, err := impl.shopCoreService.FilterOnboardedAShops(ctx, affiShopIds)
	if err != nil {
		return nil, err
	}
	shopPriceRatios, err := impl.ashopDataDM.GetAShopPriceRatioBatch(ctx, onboardedAShops)
	if err != nil {
		return nil, err
	}
	return shopPriceRatios, nil
}

func (impl *CommonSIPLogicImpl) GetAItemMarginBatch(ctx context.Context, affiShopIDToItemIdsMap map[uint64][]uint64) (map[uint64]int32, error) {
	//TODO urg: confirm if onboarded shop only satisfies listing side need
	affiShopIds := make([]uint64, 0, len(affiShopIDToItemIdsMap))
	for affiShopId := range affiShopIDToItemIdsMap {
		affiShopIds = append(affiShopIds, affiShopId)
	}
	onboardedAShopIds, err := impl.shopCoreService.FilterOnboardedAShops(ctx, affiShopIds)
	if err != nil {
		return nil, err
	}
	//TODO: remove logic to obtain mst shopid after db split is done
	aToPShopIdMap, err := impl.shopCoreService.GetPShopIdsByAShopIdsBatch(ctx, onboardedAShopIds)
	if err != nil {
		return nil, err
	}
	completeAItemIdToMarginMap := make(map[uint64]int32)
	for ashopId, pShopId := range aToPShopIdMap {
		aItemIdToMarginMap, err := impl.aItemDataDM.GetAItemMarginBatch(ctx, pShopId, affiShopIDToItemIdsMap[ashopId])
		for aItemId, margin := range aItemIdToMarginMap {
			completeAItemIdToMarginMap[aItemId] = margin
		}
		if err != nil {
			return nil, err
		}
	}
	return completeAItemIdToMarginMap, nil
}

func (impl *CommonSIPLogicImpl) GetAItemRealWeight(ctx context.Context, affiShopId, affiItemId uint64) (int32, error) {
	aShopInfo, err := impl.shopCoreService.GetAShopInfo(ctx, affiShopId)
	if err != nil {
		return 0, err
	}
	if impl.shopCoreService.IsAShopOffboarded(aShopInfo) {
		ulog.DefaultLoggerFromContext(ctx).Info("A shop has already been offboarded, ignored", ulog.Uint64("aShopId", affiShopId))
		return 0, nil
	}
	pShopId, err := impl.shopCoreService.GetPShopIdByAShopId(ctx, affiShopId)
	if err != nil {
		return 0, err
	}
	itemRealWeight, err := impl.aItemDataDM.GetAItemRealWeight(ctx, pShopId, affiItemId)
	if err != nil {
		return 0, err
	}
	return itemRealWeight, nil
}

func (impl *CommonSIPLogicImpl) GetPShopOpsPriceRatioSettingBatch(ctx context.Context, pShopId []uint64) ([]*pb.PShopOpsPriceRatioSetting, error) {
	mstShops, err := impl.mstShopDM.GetPShopInfoBatch(ctx, pShopId)
	if err != nil {
		return nil, err
	}

	pShopOpsPriceRatioSettings := make([]*pb.PShopOpsPriceRatioSetting, 0, len(mstShops))
	for _, mstShop := range mstShops {
		isControlledByOps, startTime, endTime := model.GetPShopOpsPriceRatioSettingFromMstShopRecord(mstShop)
		ratioSetting := &pb.PShopOpsPriceRatioSetting{
			IsControlledByOps: proto.Bool(isControlledByOps),
			StartTime:         proto.Int64(startTime),
			EndTime:           proto.Int64(endTime),
		}
		pShopOpsPriceRatioSettings = append(pShopOpsPriceRatioSettings, ratioSetting)
	}
	return pShopOpsPriceRatioSettings, nil
}

func (impl *CommonSIPLogicImpl) SetAItemMargin(ctx context.Context, affiShopId, affiItemId uint64, aItemMargin int64) error {
	pShopId, err := impl.shopCoreService.GetPShopIdByAShopId(ctx, affiShopId)
	if err != nil {
		return err
	}

	pShopData, err := impl.mstShopDM.GetPShopInfoWithCache(ctx, pShopId)
	if err != nil {
		return err
	}
	realAItemMarginVal := calcutil.ToRealRatio(aItemMargin)
	if pShopData.IsCbShop() {
		if realAItemMarginVal < -200 || realAItemMarginVal > 1000 {
			return cerr.Wrap(fmt.Errorf("margin should be >= -200 and <= 1000 since primaryShop is CB, pShopId=%d, given aItemMargin=%f", pShopId, aItemMargin), "", uint32(pb.Constant_ERROR_PARAMS))
		}
	} else {
		if realAItemMarginVal <= 0 || realAItemMarginVal > 10 {
			return cerr.Wrap(fmt.Errorf("margin should be > 0 and <= 10 since primaryShop is local, pShopId=%d, given aItemMargin=%f", pShopId, aItemMargin), "", uint32(pb.Constant_ERROR_PARAMS))
		}
	}
	var finalDbAItemMarginVal int32
	if pShopData.IsCbShop() {
		finalDbAItemMarginVal = int32(aItemMargin / 100)
	}

	err = impl.aItemDataDM.SetAItemMargin(ctx, pShopId, affiItemId, finalDbAItemMarginVal)
	if err != nil {
		return err
	}
	return nil
}

func (impl *CommonSIPLogicImpl) SetAItemRealWeight(ctx context.Context, affiShopId, affiItemId uint64, aItemRealWeight int64) error {
	pShopId, err := impl.shopCoreService.GetPShopIdByAShopId(ctx, affiShopId)
	if err != nil {
		return err
	}

	err = impl.aItemDataDM.SetAItemRealWeight(ctx, pShopId, affiItemId, int32(aItemRealWeight))
	if err != nil {
		return err
	}
	return nil
}

func (impl *CommonSIPLogicImpl) CreateCBSIPAShopSellerDiscountPromotion(ctx context.Context, affiShopId uint64) error {
	aShopRegion, err := impl.shopCoreService.GetShopRegionByShopId(ctx, affiShopId)
	if err != nil {
		return err
	}
	aShopDetail, err := impl.shopCoreService.GetShopDetail(ctx, affiShopId, aShopRegion)
	if err != nil {
		return err
	}
	if aShopDetail == nil {
		return cerr.Wrap(fmt.Errorf("shop detail not found for shopId=%d", affiShopId), "", uint32(pb.Constant_ERROR_NOT_FOUND))
	}

	pShopId, err := impl.shopCoreService.GetPShopIdByAShopId(ctx, affiShopId)
	if err != nil {
		return err
	}

	pShopData, err := impl.mstShopDM.GetPShopInfoWithCache(ctx, pShopId)
	if err != nil {
		return err
	}
	if !pShopData.IsCbShop() {
		return cerr.Wrap(fmt.Errorf("p shop is not CB, pShopId=%d", pShopId), "", uint32(pb.Constant_ERROR_PARAMS))
	}
	shopData, err := impl.ashopDataDM.GetAShopData(ctx, affiShopId)
	if err != nil {
		return err
	}
	if shopData.GetPromotionId() > 0 {
		currentSellerDiscountInfo, err := impl.itemDiscountService.GetShopSellerDiscountByShopIdPromoId(ctx, affiShopId, shopData.GetPromotionId(), aShopRegion)
		if err != nil {
			return err
		}
		// seller discout is already created, and not ended yet, not need to create.
		if impl.itemDiscountService.IsSellerDiscountPromotionValid(currentSellerDiscountInfo) {
			return cerr.Wrap(fmt.Errorf("A shop seller discount already created and still valid, promoId=%d", currentSellerDiscountInfo.GetPromotionId()), "", uint32(pb.Constant_ERROR_PARAMS))
		}
	}
	currTime := time.Now().Unix()
	// 15465600 = 179*24*3600, discount duration can not be longer than 180 days, use 179 for safety
	newSellerDiscountPromoID, err := impl.itemDiscountService.AddShopSellerDiscount(ctx, affiShopId, uint64(aShopDetail.UserId), aShopRegion, "Price Sync SIP Product Promotion", currTime, currTime+constant.CST179Days)
	if err != nil {
		return err
	}
	err = impl.ashopDataDM.SetAShopDataPromoId(ctx, affiShopId, pShopId, newSellerDiscountPromoID)
	if err != nil {
		return err
	}
	ulog.DefaultLoggerFromContext(ctx).Info("A shop seller discount successfully created", ulog.Uint64("a_shop_id", affiShopId), ulog.Uint64("promotion_id", newSellerDiscountPromoID))
	return nil
}

func (impl *CommonSIPLogicImpl) GetCBSIPAShopSellerDiscountPromotion(ctx context.Context, affiShopId uint64) (uint64, error) {
	aShopSellerDiscountPromoId, err := impl.ashopDataDM.GetAShopPromoId(ctx, affiShopId)
	if err != nil {
		return 0, err
	}
	return aShopSellerDiscountPromoId, nil
}
