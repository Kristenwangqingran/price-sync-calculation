package calculate

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/google/wire"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/data"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	db2 "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/factors"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

var calculateAffiMpskuDmProvider = wire.NewSet(
	wire.Struct(new(CalculateAffiMpskuDmOpts), "*"),
	NewCalculateAffiMpskuDm,
)

type CalculateAffiMpskuDmOpts struct {
	LocalShippingFeeConfigDB db2.LocalShippingFeeConfigDB
	LogisticService          service.LogisticService
	CalculationFactorsRepo   factors.CalculationFactorsRepo
}

type calculateAffiMpskuDm struct {
	localShippingFeeConfigDB db2.LocalShippingFeeConfigDB
	logisticService          service.LogisticService
	calculationFactorsRepo   factors.CalculationFactorsRepo
}

func NewCalculateAffiMpskuDm(opts *CalculateAffiMpskuDmOpts) CalculateAffiMpskuDm {
	return &calculateAffiMpskuDm{
		localShippingFeeConfigDB: opts.LocalShippingFeeConfigDB,
		logisticService:          opts.LogisticService,
		calculationFactorsRepo:   opts.CalculationFactorsRepo,
	}
}

func (dm *calculateAffiMpskuDm) CalcLocalSipOverseaDiscount(ctx context.Context,
	req *pb.CalcLocalSipOverseaDiscountPriceRequest, calcData *data.CalcFactorDataForAffiMpsku) ([]*pb.LocalSIPAffiPriceResult, error) {
	results := make([]*pb.LocalSIPAffiPriceResult, 0)
	for _, affiItemModelId := range req.GetAffiItemModelIds() {
		itemModelId := model.ItemModelId{
			ItemId:  affiItemModelId.GetItemId(),
			ModelId: affiItemModelId.GetModelId(),
		}
		result := dm.doCalcLocalSipOverseaDiscount(ctx,
			req.GetAffiShopId(), req.GetAffiRegion(), req.GetDiscountRate(),
			itemModelId, calcData)
		results = append(results, result)
	}
	return results, nil
}

func (dm *calculateAffiMpskuDm) doCalcLocalSipOverseaDiscount(ctx context.Context,
	affiShopId int64, affiRegion string, discountRate int64,
	affiItemModelId model.ItemModelId, calcData *data.CalcFactorDataForAffiMpsku) *pb.LocalSIPAffiPriceResult {
	result := &pb.LocalSIPAffiPriceResult{}

	primaryItemModel, ok := calcData.ItemModelMapping[affiItemModelId]
	if !ok {
		errMsg := fmt.Sprintf("primary itemModel not found|affiItemModelId=%v", affiItemModelId)
		result.ErrorDetail = proto.String(errMsg)
		result.CalcErr = proto.Uint32(uint32(pb.Constant_DISCOUNT_PRICE_FACTOR_NOT_FOUND))

		return result
	}

	primaryOriginPrice, ok := calcData.PrimaryOriginPrices[primaryItemModel]
	if !ok {
		errMsg := fmt.Sprintf("primary origin price not found|affiItemModelId=%v", affiItemModelId)
		result.ErrorDetail = proto.String(errMsg)
		result.CalcErr = proto.Uint32(uint32(pb.Constant_DISCOUNT_PRICE_FACTOR_NOT_FOUND))

		return result
	}

	affiOriginPrice, ok := calcData.AffiOriginPrices[affiItemModelId]
	if !ok {
		errMsg := fmt.Sprintf("affi origin price not found|affiItemModelId=%v", affiItemModelId)
		result.ErrorDetail = proto.String(errMsg)
		result.CalcErr = proto.Uint32(uint32(pb.Constant_DISCOUNT_PRICE_FACTOR_NOT_FOUND))

		return result
	}

	aItemData, ok := calcData.AItemData[affiItemModelId.ItemId]
	if !ok {
		errMsg := fmt.Sprintf("a item data not found|affiItemModelId=%v", affiItemModelId)
		result.ErrorDetail = proto.String(errMsg)
		result.CalcErr = proto.Uint32(uint32(pb.Constant_DISCOUNT_PRICE_FACTOR_NOT_FOUND))

		return result
	}

	pItemId, ok := calcData.AItemIdToPItemIdMapping[affiItemModelId.ItemId]
	if !ok {
		errMsg := fmt.Sprintf("p itemid mapped to a itemid not found|affiItemModelId=%v", affiItemModelId)
		result.ErrorDetail = proto.String(errMsg)
		result.CalcErr = proto.Uint32(uint32(pb.Constant_DISCOUNT_PRICE_FACTOR_NOT_FOUND))

		return result
	}
	primaryItemData, ok := calcData.PrimaryItemData[pItemId]
	if !ok {
		errMsg := fmt.Sprintf("primary item weight not found|affiItemModelId=%v", affiItemModelId)
		result.ErrorDetail = proto.String(errMsg)
		result.CalcErr = proto.Uint32(uint32(pb.Constant_DISCOUNT_PRICE_FACTOR_NOT_FOUND))

		return result
	}

	weight := primaryItemData.Weight
	if aItemData.GetAffiRealWeight() > 0 {
		weight = int64(aItemData.GetAffiRealWeight())
	}

	itemMargin := calcutil.ToRealRatio(int64(aItemData.GetItemMargin()))
	shopMargin := calcutil.ToRealRatio(calcData.ShopMargin)

	shippingFee, err := dm.calcShippingFee(ctx, affiShopId, affiItemModelId, weight, calcData)
	if err != nil {
		logging.GetLogger(ctx).Error("calcShippingFee failed", ulog.Error(err), ulog.Reflect("affiItemModelId", affiItemModelId))

		errMsg := err.Error()
		result.ErrorDetail = proto.String(errMsg)
		result.CalcErr = proto.Uint32(cerr.Code(err))

		return result
	}

	initHiddenPrice, err := dm.calcHiddenPrice(ctx, pItemId, weight, calcData)
	if err != nil {
		logging.GetLogger(ctx).Error("calcHiddenPrice failed", ulog.Error(err), ulog.Reflect("affiItemModelId", affiItemModelId))

		errMsg := err.Error()
		result.ErrorDetail = proto.String(errMsg)
		result.CalcErr = proto.Uint32(cerr.Code(err))

		return result
	}

	affiPrice := calcutil.CalculateAffiPriceForLocalSip(ctx,
		calcutil.ToRealPrice(primaryOriginPrice),
		calcutil.DbWeightToGram(weight), itemMargin, shopMargin,
		initHiddenPrice, shippingFee,
		calcData.LocalPriceConfig, proto.Float64(calcutil.ToRealRatio(discountRate)))

	if dm.hitVNPriceLimit(ctx, affiRegion, calcutil.ToRealPrice(affiOriginPrice), affiPrice) {
		result.ErrorDetail = proto.String(fmt.Sprintf("hit vn price limit|affiOriginPrice=%f|affiPrice=%f|affiItemModelId=%v",
			calcutil.ToRealPrice(affiOriginPrice), affiPrice, affiItemModelId))
		result.CalcErr = proto.Uint32(uint32(pb.Constant_DISCOUNT_PRICE_HIT_LIMIT))

		return result
	}

	result.AffiPrice = proto.Int64(calcutil.ToDBPrice(calcutil.PriceRoundUpByCountry(affiRegion, affiPrice)))

	return result
}

func (dm *calculateAffiMpskuDm) calcHiddenPrice(ctx context.Context,
	primaryItemId uint64, weight int64, calcData *data.CalcFactorDataForAffiMpsku) (float64, error) {

	hiddenPriceResults, err := dm.calculationFactorsRepo.GetInitialHiddenPriceForLocalSip(
		ctx, primaryItemId, uint64(calcData.PrimaryShopId), calcData.PrimaryRegion,
		[]model.LocalSipHiddenPriceQuery{
			{
				QueryId:      0,
				CommonConfig: calcData.LocalPriceConfig,
				ARegion:      calcData.AffiRegion,
				AShopId:      uint64(calcData.AffiShopId),
				Weight:       weight,
				PRegion:      calcData.PrimaryRegion,
			},
		})
	if err != nil {
		return 0, err
	}

	if len(hiddenPriceResults) != 1 {
		return 0, cerr.New("hiddenPriceResults length is not 1 as expected.",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	return hiddenPriceResults[0].HiddenPrice, nil
}

func (dm *calculateAffiMpskuDm) calcShippingFee(ctx context.Context,
	affiShopId int64, affiItemModelId model.ItemModelId, weight int64, calcData *data.CalcFactorDataForAffiMpsku) (float64, error) {
	if calcData.LocalPriceConfig.ShippingFeeToggle != nil && *calcData.LocalPriceConfig.ShippingFeeToggle == model.PriceSyncToggleReadFromSLS {
		return dm.calcShippingFeeFromSLS(ctx, affiShopId, affiItemModelId, weight, calcData)
	} else {
		return dm.calcShippingFeeFromDB(ctx, weight, calcData)
	}
}

func (dm *calculateAffiMpskuDm) calcShippingFeeFromDB(ctx context.Context,
	weight int64, calcData *data.CalcFactorDataForAffiMpsku) (float64, error) {
	record, err := dm.localShippingFeeConfigDB.GetLocalShippingFeeConfigRecordByWeight(ctx,
		calcData.PrimaryRegion, calcData.AffiRegion, weight)
	if err != nil {
		return 0, err
	}

	if record == nil {
		return 0, cerr.New(fmt.Sprintf("shipping fee record for weight not found|P-region=%s|A-region=%s|weight=%d",
			calcData.PrimaryRegion, calcData.AffiRegion, weight), uint32(pb.Constant_ERROR_NOT_FOUND))
	}

	return calcutil.ToRealPrice(record.ShippingFeePrice), nil
}

// TODO: @wang.zhong can migrate to use CalculationFactorsRepoImpl.GetShippingFeeForLocalSip
func (dm *calculateAffiMpskuDm) calcShippingFeeFromSLS(ctx context.Context,
	affiShopId int64, affItemModelId model.ItemModelId, weight int64, calcData *data.CalcFactorDataForAffiMpsku) (float64, error) {
	enabledChannelInfo, ok := calcData.AffiItemEnabledChannelIds[affItemModelId.ItemId]
	if !ok {
		logging.GetLogger(ctx).Warn("calcShippingFeeFromSLS: enabledChannelIDList not found", ulog.Uint64("affiItemId", affItemModelId.ItemId))
		return 0, cerr.New("enabledChannelIDList not found", uint32(pb.Constant_ERROR_GET_ENABLED_CHANNELS))
	}
	if enabledChannelInfo.Err != nil {
		return 0, enabledChannelInfo.Err
	}

	if len(enabledChannelInfo.EnabledChannelIds) < 1 {
		logging.GetLogger(ctx).Warn("calcShippingFeeFromSLS: EnabledChannelIds is empty", ulog.Uint64("affiItemId", affItemModelId.ItemId))
		return 0, cerr.New("EnabledChannelIds is empty", uint32(pb.Constant_ERROR_EMPTY_ENABLED_CHANNELS))
	}

	leafCategoryID, ok := calcData.AffiItemLeafCategoryIDMap[affItemModelId.ItemId]
	if !ok {
		logging.GetLogger(ctx).Warn("calcShippingFeeFromSLS: leafCategoryID not found", ulog.Uint64("affiItemId", affItemModelId.ItemId))
		return 0, cerr.New("leafCategoryID not found", uint32(pb.Constant_ERROR_CALCULATE_HIDDEN_FEE))
	}
	if leafCategoryID.Err != nil {
		return 0, leafCategoryID.Err
	}

	weightInGram := calcutil.DbWeightToGram(weight)

	shippingFee, err := dm.logisticService.CalcShippingFeeForLocalSIP(ctx,
		uint64(affiShopId), calcData.AffiRegion, affItemModelId.ItemId,
		uint64(leafCategoryID.LeafCatId), weightInGram, enabledChannelInfo.EnabledChannelIds)
	return shippingFee, err
}

func (dm *calculateAffiMpskuDm) hitVNPriceLimit(ctx context.Context, region string, originPrice, promotionPrice float64) bool {
	if strings.ToUpper(region) == "VN" && calcutil.Gt(originPrice/promotionPrice, constant.VnPromoRatioLimit) {
		logging.GetLogger(ctx).Warn("hitVNPriceLimit|originPrice=%f|promotionPrice=%f")
		return true
	}

	return false
}
