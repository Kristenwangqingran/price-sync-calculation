package cb_sip_logic

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/clog"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/calculate"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	ib "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/item_business.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/factors"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_v2_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

func (c *CbSipLogicImpl) CalculateAPriceByPItemForCbSip(ctx context.Context, request model.CbSipCalculateAPriceByPItemRequest) ([]model.CbSipCalculateAPriceByPItemResult, error) {
	pItemData, err := c.getPItemInfo(ctx, request.PShopId, request.PItemId)
	if err != nil {
		return nil, err
	}

	aItemData, err := c.aItemDataDM.GetAItemData(ctx, request.PShopId, request.AItemId)
	if err != nil {
		if !c.isNotFoundError(err) || !request.CalculateForCreate {
			return nil, err
		}
	}

	itemMargin := calcutil.ToRealPect(int(aItemData.GetItemMargin()))
	itemRealWeight := int64(aItemData.GetAffiRealWeight())

	var pItemWeight int64
	if pItemData != nil {
		pItemWeight = pItemData.Weight
	}

	weight := calcutil.DbWeightToGram(factors.GetCalcWeightByAItemAndPItemWeight(itemRealWeight, pItemWeight))

	isSipPShop, err := c.isSipPShop(ctx, request.PShopId)
	if err != nil {
		return nil, err
	}
	if !isSipPShop {
		logging.GetLogger(ctx).Error(fmt.Sprintf("shopId=%d is not primary shop", request.PShopId))
		return nil, nil
	}

	pShopData, err := c.getPShopInfo(ctx, request.PShopId, false)
	if err != nil {
		return nil, err
	}

	aHiddenPrice := c.factorsRepo.GetHiddenPriceForCbSip(ctx, pItemData, pShopData, request.MerchantRegion, request.PRegion, request.ARegion, weight, false)

	serviceFee, err := c.factorsRepo.GetShopServiceFeeForCbSip(ctx, request.ARegion, int64(request.AShopId))
	if err != nil {
		return nil, err
	}

	commissionFee, err := c.factorsRepo.GetShopCommissionFeeForCbSip(ctx, request.ARegion, int64(request.AShopId))
	if err != nil {
		return nil, err
	}

	handlingFee, err := c.factorsRepo.GetHandlingFeeForCbSip(ctx)
	if err != nil {
		return nil, err
	}

	finalFee := 1 + serviceFee + commissionFee + handlingFee
	if finalFee <= 0 {
		return nil, cerr.New(fmt.Sprintf("sync price biz exception: serviceFee %v + commissionFee %v + handlingFee %v <= 0", serviceFee, commissionFee, handlingFee), uint32(pb.Constant_ERROR_INTERNAL))
	}

	pProductInfo, err := c.getItemProductInfo(ctx, request.PShopId, request.PItemId, request.PRegion)
	if err != nil {
		return nil, err
	}

	srcCurrency, dstCurrency, err := c.factorsRepo.GetCurrencyForCbSip(ctx, request.MerchantId, request.ARegion, pProductInfo)
	if err != nil {
		return nil, err
	}

	exchangeRate, err := c.factorsRepo.GetExchangeRateForCbSip(ctx, srcCurrency, dstCurrency, false)
	if err != nil {
		return nil, err
	}

	aShopData, err := c.aShopDataDM.GetAShopData(ctx, request.AShopId)
	if err != nil {
		return nil, err
	}

	shopMargin := aShopData.GetShopMargin()

	aShopInfo, err := c.shopCoreService.GetAShopInfo(ctx, request.AShopId)
	if err != nil {
		return nil, err
	}

	if c.shopCoreService.IsAShopOffboarded(aShopInfo) {
		shopMargin = 0
	}

	countryMargin, err := c.factorsRepo.GetCountryMarginForCbSip(ctx, request.PRegion, request.ARegion)
	if err != nil {
		return nil, err
	}

	priceRatio := calcutil.ToRealPect(int(aShopData.GetPriceRatio()))

	res := make([]model.CbSipCalculateAPriceByPItemResult, len(request.Queries))
	for i, query := range request.Queries {
		var affiNormalPriceDB int64
		var affiSettlementPriceDB int64
		ratio := 1.0
		aPromotionPrice := int64(-1)
		pNormalPriceReal := calcutil.ToRealPrice(query.PNormalPrice)

		basePrice := query.PItemPrice

		if basePrice <= 0 {
			return nil, cerr.New(fmt.Sprintf("invalid p_item_price=%v", query.PItemPrice), uint32(pb.Constant_ERROR_PARAMS))
		}
		if query.PPromotionPrice != nil && *query.PPromotionPrice > 0 {
			pPromotionPriceReal := calcutil.ToRealPrice(*query.PPromotionPrice)
			ratio = pNormalPriceReal / pPromotionPriceReal
			if strings.ToUpper(request.ARegion) == "VN" && ratio > constant.VnPromoRatioLimit {
				ratio = constant.VnPromoRatioLimit
			}
			aPromotionPrice = calcutil.CalcAffiDBPriceForCbSip(basePrice, request.ARegion, exchangeRate, priceRatio, 1.0, countryMargin, float64(shopMargin), float64(itemMargin), aHiddenPrice, finalFee)
		}

		affiNormalPriceDB = calcutil.CalcAffiDBPriceForCbSip(basePrice, request.ARegion, exchangeRate, priceRatio, ratio, countryMargin, float64(shopMargin), float64(itemMargin), aHiddenPrice, finalFee)
		affiSettlementPriceDB = calcutil.DBPriceRoundNearest(srcCurrency, int64(float64(basePrice)*priceRatio)) // basePrice is db price value alr

		logging.GetLogger(ctx).Info(fmt.Sprintf("[CB SIP] Calc price for CBSIP, "+
			"query=%v, basePrice=%v, ratio=%v, aRegion=%v, exchangeRate=%v, priceRatio=%v, countryMargin=%v, shopMargin=%v, itemMargin=%v, aHiddenPrice=%v, serviceFee=%v, commissionFee=%v, handlingFee=%v "+
			"| result: affiNormalPriceDB=%v, affiPromotionPriceDB=%v, affiSettlementPriceDB=%v",
			cutil.JSONEncode(query), basePrice, ratio, request.ARegion, exchangeRate, priceRatio, countryMargin, shopMargin, itemMargin, aHiddenPrice, serviceFee, commissionFee, handlingFee,
			affiNormalPriceDB, aPromotionPrice, affiSettlementPriceDB))

		if affiSettlementPriceDB < 0 {
			return nil, cerr.New(fmt.Sprintf("a settlement price is invalid, query=%v, price=%v",
				cutil.JSONEncode(query), affiSettlementPriceDB), uint32(pb.Constant_ERROR_INTERNAL))
		}

		res[i] = model.CbSipCalculateAPriceByPItemResult{
			ANormalPrice:             affiNormalPriceDB,
			APromotionPrice:          aPromotionPrice,
			ASettlementPrice:         affiSettlementPriceDB,
			ASettlementPriceCurrency: srcCurrency,
			Snap: &pb.CbSipPriceFactorSnap{
				Weight:          proto.Float64(weight),
				CountryMargin:   proto.Float64(countryMargin),
				ShopMargin:      proto.Float64(float64(shopMargin)),
				ItemMargin:      proto.Float64(float64(itemMargin)),
				ExchangeRate:    proto.Float64(exchangeRate),
				PriceRatio:      proto.Float64(priceRatio),
				AffiHiddenPrice: proto.Float64(aHiddenPrice),
				SrcCurrency:     proto.String(srcCurrency),
				ServiceFee:      proto.Float64(serviceFee),
				CommissionFee:   proto.Float64(commissionFee),
				HandlingFee:     proto.Float64(handlingFee),
			},
		}
	}

	return res, nil
}

func (c *CbSipLogicImpl) CalculateAItemOPL(ctx context.Context, request *model.CbSipCalculateAOPLByPItemRequest) (*model.CbSipCalculateAOPLByPItemResult, error) {
	if config.GetOPLConfig().CustomizedOplRegionBlackList[request.ARegion] {
		clog.Infof(ctx, "A shop region in CustomizedOPLRegionBlacklist, aShopId=%d, aRegion=%s", request.AShopId, request.ARegion)
		return &model.CbSipCalculateAOPLByPItemResult{}, nil
	}

	oplPrices, purchaseLimit, err := c.priceBusinessService.GetOPLPrices(ctx, request.PItemId, request.PRegion)
	if err != nil {
		clog.Errorf(ctx, "error when getting OPL prices, pItemId=%d, pRegion=%s, err=%v", request.PItemId, request.PRegion, err)
		return nil, err
	}
	if len(oplPrices) == 0 {
		return &model.CbSipCalculateAOPLByPItemResult{}, nil
	}

	opl := calculate.GetAItemCustomizedOPLFromPItemOPLPrices(oplPrices, purchaseLimit)

	return &model.CbSipCalculateAOPLByPItemResult{
		Opl: opl,
	}, nil
}

func (c *CbSipLogicImpl) isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return cerr.Code(err) == uint32(pb.Constant_ERROR_NOT_FOUND)
}

func (c *CbSipLogicImpl) getItemProductInfo(ctx context.Context, pShopId, pItemId uint64, pRegion string) (*ib.ProductInfo, error) {
	infoTypes := []uint32{
		uint32(ib.Constant_PRICE), uint32(ib.Constant_ITEM_BASIC), uint32(ib.Constant_CATEGORY), uint32(ib.Constant_LOGISTICS),
	}
	ctx, _ = cidutil.FillCtxWithNewCID(ctx, pRegion)
	resp, err := c.ibsService.GetProductInfoForDisplay(ctx, &ib.GetProductInfoForDisplayRequest{
		ShopItemIds: []*ib.ShopItemId{
			{
				ShopId: proto.Uint32(uint32(pShopId)),
				ItemId: proto.Uint64(pItemId),
			},
		},
		InfoTypes:   infoTypes,
		Region:      proto.String(pRegion),
		NeedDeleted: proto.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.GetInfo()) == 0 {
		return nil, cerr.New(fmt.Sprintf("failed to get item product info, pItemId=%v", pItemId), uint32(pb.Constant_ERROR_NOT_FOUND))

	}
	return resp.GetInfo()[0], nil
}

func (c *CbSipLogicImpl) getPItemInfo(ctx context.Context, pShopId uint64, pItemId uint64) (*sip_v2_db.MstItemRecord, error) {
	session := c.sipV2Repo.DbSession()
	res, err := c.sipV2Repo.GetMstItemRecordBatch(ctx, session, pShopId, []uint64{pItemId})
	if err != nil {
		return nil, err
	}
	if len(res) < 1 {
		return nil, nil
	}

	return res[0], nil
}

func (c *CbSipLogicImpl) getPShopInfo(ctx context.Context, pShopId uint64, useCache bool) (*sip_db.MstShop, error) {
	var mstShop *sip_db.MstShop
	var err error
	if useCache {
		mstShop, err = c.mstShopDM.GetPShopInfoWithCache(ctx, pShopId)
	} else {
		mstShop, err = c.mstShopDM.GetPShopInfo(ctx, pShopId)
	}
	if err != nil {
		return nil, err
	}
	if mstShop == nil {
		ulog.DefaultLoggerFromContext(ctx).Warn("no MstShop record found in DB", ulog.Uint64("pShopId", pShopId))
		return nil, nil
	}
	return mstShop, nil
}

func (c *CbSipLogicImpl) getPShopInfoBatch(ctx context.Context, pShopIds []uint64) ([]*sip_db.MstShop, error) {
	mstShops, err := c.mstShopDM.GetPShopInfoBatch(ctx, pShopIds)
	if err != nil {
		return nil, err
	}
	if len(mstShops) == 0 {
		ulog.DefaultLoggerFromContext(ctx).Warn("no MstShop records found in DB", ulog.Reflect("pShopIds", pShopIds))
		return nil, nil
	}
	return mstShops, nil
}

func (c *CbSipLogicImpl) isSipPShop(ctx context.Context, pShopId uint64) (bool, error) {
	region, err := c.shopCoreService.GetShopRegionByShopId(ctx, pShopId)
	if err != nil {
		return false, err
	}
	if len(region) == 0 {
		return false, cerr.New(fmt.Sprintf("get_shop_region return emtpy, shopId=%v", pShopId), uint32(pb.Constant_ERROR_NOT_FOUND))
	}
	shopInfo, err := c.shopCoreService.GetShopDetail(ctx, pShopId, region)
	if err != nil {
		return false, err
	}

	if shopInfo == nil || !shopInfo.IsSipPrimary {
		return false, nil
	}

	return true, nil
}

func (c *CbSipLogicImpl) isCbscShop(ctx context.Context, pShopId uint64) (bool, error) {
	isCbsc, err := c.shopMerchantService.CheckMerchantShopCbsc(ctx, pShopId)
	if err != nil {
		return false, err
	}
	return isCbsc, nil
}
