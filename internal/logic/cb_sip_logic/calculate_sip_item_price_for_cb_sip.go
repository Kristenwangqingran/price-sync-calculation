package cb_sip_logic

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

func (c *CbSipLogicImpl) CalculateSipItemPriceForCbSip(ctx context.Context, req model.CbSipCalculateSipItemPriceRequest) ([]model.CbSipCalculateSipItemPriceResult, error) {
	isCbscSipPShop, err := c.isCbscSipPShop(ctx, req.ShopId)
	if err != nil {
		return nil, err
	}
	if !isCbscSipPShop {
		logging.GetLogger(ctx).Info(fmt.Sprintf("shopId=%v is not cbsc sip p_shop, skip calculation", req.ShopId))
		return nil, nil
	}
	isAllowedEditItemPriceShop, err := c.getShopAllowEditItemPrice(ctx, req.ShopId)
	if err != nil {
		return nil, err
	}

	if isAllowedEditItemPriceShop && req.ItemId > 0 {
		stopAutoSyncMap, err := c.getSipProductStopSyncModelMap(ctx, req.Region, req.ShopId, req.ItemId)
		if err != nil {
			return nil, err
		}

		filteredQuery := make([]model.CbSipCalculateSipItemPriceSingleQuery, 0)
		for _, query := range req.Queries {
			if stopAutoSyncMap[query.ModelId] {
				continue
			}
			filteredQuery = append(filteredQuery, query)
		}

		logging.GetLogger(ctx).Info(fmt.Sprintf("filter out queries whose stop_auto_sync = true, newQueries=%+v", filteredQuery))
		req.Queries = filteredQuery

		if len(req.Queries) == 0 {
			return nil, nil
		}
	}

	var decimalReferenceServiceFee, decimalTransactionFee, decimalCommissionFee decimal.Decimal
	var decimalExchangeRate decimal.Decimal

	merchantInfo, err := c.shopMerchantService.GetMerchantInfoByShopId(ctx, req.ShopId)
	if err != nil {
		return nil, err
	}

	decimalPricePrecision := decimal.NewFromInt32(int32(constant.PricePrecision))
	cbFees, err := c.factorsRepo.GetCbscShopLevelFeeRate(ctx, uint64(merchantInfo.GetMerchantId()), nil, []uint64{req.ShopId})
	if err != nil {
		return nil, err
	}

	if len(cbFees) > 0 {
		refFee, commissionFee, transactionFee := cbFees[0].GetReferenceServiceFeeRate(), cbFees[0].GetCommissionRate(), cbFees[0].GetTransactionFeeRate()
		decimalTransactionFee = decimal.NewFromInt(transactionFee * 10).Div(decimalPricePrecision)
		decimalReferenceServiceFee = decimal.NewFromInt(refFee * 10).Div(decimalPricePrecision) // //nolint:lll
		decimalCommissionFee = decimal.NewFromInt(commissionFee * 10).Div(decimalPricePrecision)
	}

	exchangeRateInfo, err := c.exchangeRateService.GetMerchantExchangeRateInfo(ctx, uint64(merchantInfo.GetMerchantId()))
	if err != nil {
		return nil, err
	}
	currency := exchangeRateInfo.GetCurrency()
	if currency == "" {
		return nil, cerr.New("get merchant currency is empty when processing item price", uint32(pb.Constant_ERROR_EXTERNAL))
	}

	targetCurrency, err := config.GetCurrencyByRegion(req.Region)
	if err != nil {
		return nil, err
	}
	rateFloat, err := c.factorsRepo.GetExchangeRateForCbSip(ctx, currency, targetCurrency, true)
	if err != nil {
		return nil, err
	}
	decimalExchangeRate = decimal.NewFromFloat(rateFloat) // rateFloat is original exchange_rate

	deductFee := decimal.NewFromInt(1).Sub(decimalReferenceServiceFee).Sub(decimalTransactionFee).Sub(decimalCommissionFee)

	hiddenFee, err := c.logisticService.CalcHiddenFeeForCbSip(ctx, req.ShopId, req.Region, req.ItemId, req.LeafCategoryId, req.Weight, req.ChannelIdList)
	if err != nil {
		return nil, err
	}
	hiddenFeeDb := calcutil.RoundFloatToInt(hiddenFee, constant.PricePrecision, 2)
	decimalHiddenPrice := decimal.NewFromInt(hiddenFeeDb).Div(decimalPricePrecision)

	results := make([]model.CbSipCalculateSipItemPriceResult, len(req.Queries))
	for i, info := range req.Queries {
		price := info.Price // already magnify 100000 times
		decimalPrice := decimal.NewFromInt(price).Div(decimalPricePrecision)

		var decimalItemPrice decimal.Decimal
		if decimalPrice.Mul(deductFee).Sub(decimalHiddenPrice).IsPositive() {
			// psku_price*(1 - commission_fee - service_fee - transaction_fee) - hidden_price > 0
			decimalItemPrice = decimalPrice.Mul(deductFee).Sub(decimalHiddenPrice).DivRound(decimalExchangeRate, 2)
		} else {
			decimalItemPrice = decimalPrice.Mul(deductFee).DivRound(decimalExchangeRate, 2)
		}
		if decimalItemPrice.Sub(decimal.NewFromFloat(0.01)).IsNegative() {
			// item_price < 0.01, then item_price=0.01
			decimalItemPrice = decimal.NewFromFloat(0.01)
		}
		magnifyDecimalItemPrice := decimalItemPrice.Mul(decimalPricePrecision)
		itemPrice := uint64(magnifyDecimalItemPrice.IntPart()) // itemPrice magnify 100000 times

		logging.GetLogger(ctx).Info(fmt.Sprintf("[CB SIP] Calc item price for CBSIP, query=%+v, referenceServiceFee=%v, transactionFee=%v, commissionFee=%v, deductFee=%v, hiddenFee=%v, exchangeRate=%v | result=%v", info, decimalReferenceServiceFee, decimalTransactionFee, decimalCommissionFee, deductFee, decimalHiddenPrice, decimalExchangeRate, itemPrice))

		results[i] = model.CbSipCalculateSipItemPriceResult{
			ModelId:        info.ModelId,
			CbSipItemPrice: int64(itemPrice),
			Currency:       currency,
		}
	}

	return results, nil
}

func (c *CbSipLogicImpl) isShopUserBannedOrDeleted(ctx context.Context, shopId uint64, region string) (bool, error) {
	shopUserStatusMap, err := c.accountServiceRepo.GetUserStatusMap(ctx, []model.ShopIdRegion{
		{
			ShopId: shopId,
			Region: region,
		},
	})
	if err != nil {
		return false, err
	}

	status, ok := shopUserStatusMap[shopId]
	if !ok {
		return false, cerr.New(fmt.Sprintf("failed to get shop user status map for shopId=%v", shopId), uint32(pb.Constant_ERROR_NOT_FOUND))
	}

	return status == model.StatusAccountDelete || status == model.StatusAccountBanned, nil
}

func (c *CbSipLogicImpl) isCbscSipPShop(ctx context.Context, shopId uint64) (bool, error) {
	isCbscShop, err := c.isCbscShop(ctx, shopId)
	if err != nil {
		return false, err
	}
	if !isCbscShop {
		return false, nil
	}
	isSipPShop, err := c.isSipPShop(ctx, shopId)
	if err != nil {
		return false, err
	}
	return isSipPShop, nil
}

func (c *CbSipLogicImpl) getPShopInfoWithCache(ctx context.Context, pShopId uint64) (*sip_db.MstShop, error) {
	return c.mstShopDM.GetPShopInfoWithCache(ctx, pShopId)
}

func (c *CbSipLogicImpl) getShopAllowEditItemPrice(ctx context.Context, pShopId uint64) (bool, error) {
	isAllowedEditItemPriceShop := c.allowEditItemPriceShopRepo.Exist(ctx, pShopId)
	if !isAllowedEditItemPriceShop {
		return false, nil
	}

	pShop, err := c.getPShopInfoWithCache(ctx, pShopId)
	if err != nil {
		return false, err
	}
	if !pShop.IsCbShop() {
		return false, nil
	}

	return true, nil
}

func (c *CbSipLogicImpl) getSipProductStopSyncModelMap(ctx context.Context, region string, shopId, itemId uint64) (map[uint64]bool, error) {
	info, err := c.listingUploadService.GetProductInfo(ctx, region, shopId, itemId)
	if err != nil {
		return nil, err
	}
	if len(info.ProductInfoList) == 0 {
		return nil, cerr.New(fmt.Sprintf("product info not found, itemId=%v, shopId=%v, region=%v", itemId, shopId, region), uint32(pb.Constant_ERROR_NOT_FOUND))
	}

	result := make(map[uint64]bool)
	for _, modelInfo := range info.GetProductInfoList()[0].GetSalesInfo().GetModelList() {
		result[modelInfo.GetModelId()] = modelInfo.GetPriceInfo().GetSipItemPriceInfo().GetStopSipPriceAutoSync()
	}
	return result, nil
}
