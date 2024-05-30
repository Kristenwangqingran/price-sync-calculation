package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	commonCache "git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/http"
	internalExchangeRatePb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_exchange_rate.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/httpcliutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type ExchangeRateServiceDm struct {
	exchangeRateCacheManager cache.ExchangeRateCacheManager
	httpCli                  httpcliutil.HTTPCli
}

func NewExchangeRateService(exchangeRateCacheManager cache.ExchangeRateCacheManager, httpCli httpcliutil.HTTPCli) ExchangeRateService {
	return &ExchangeRateServiceDm{
		exchangeRateCacheManager: exchangeRateCacheManager,
		httpCli:                  httpCli,
	}
}

func (dm *ExchangeRateServiceDm) GetMerchantExchangeRateMap(ctx context.Context, merchantIdList []uint64) map[uint64]*MerchantExchangeRateInfo {
	merchantExchangeRateInfoMap := make(map[uint64]*MerchantExchangeRateInfo)

	for _, merchantId := range merchantIdList {
		merchantCurrency, exchangeRateMap, err := dm.GetMerchantExchangeRate(ctx, merchantId)
		if err != nil {
			errMsg := fmt.Sprintf("failed to get exchange rate, merchantId=%d, err=%s",
				merchantId, err.Error())
			logging.GetLogger(ctx).Error(errMsg)

			merchantExchangeRateInfoMap[merchantId] = &MerchantExchangeRateInfo{
				Err: cerr.New(errMsg,
					uint32(priceSyncPriceCalculationPb.Constant_ERROR_GET_MERCHANT_EXCHANGE_RATE)),
			}

			continue
		}
		merchantExchangeRateInfoMap[merchantId] = &MerchantExchangeRateInfo{
			MerchantCurrency:        merchantCurrency,
			MerchantExchangeRateMap: exchangeRateMap,
		}
	}

	return merchantExchangeRateInfoMap
}

func (dm *ExchangeRateServiceDm) GetMerchantExchangeRateInfo(ctx context.Context, merchantId uint64) (*internalExchangeRatePb.ExchangeRateInfo, error) {
	// fetch from cache first
	key := dm.exchangeRateCacheManager.Key(merchantId)
	exchangeRateInfo, err := dm.exchangeRateCacheManager.Get(ctx, key)
	if err == nil && exchangeRateInfo != nil {
		return exchangeRateInfo, nil
	}
	if err != nil && err != commonCache.ErrCacheMiss {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to get exchange rate from cache, key=%s, err=%s", key, err.Error()))
	}
	// fetch from http api
	param := map[string]string{
		"merchant_id": strconv.FormatUint(merchantId, 10),
	}

	ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, cidutil.GlobalCID)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to fill CID, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	resp, err := http.GetSellerAdminHttp(dm.httpCli, ctxWithCID, "", constant.GetCurrencyExchangeRate, param)
	if err != nil || resp == nil || !resp.IsOk() {
		errMsg := "failed to GetMerchantExchangeRate"
		if err != nil {
			errMsg = fmt.Sprintf("%s, err=%s", errMsg, err.Error())
		}

		return nil, cerr.New(errMsg,
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
	}

	body, err := resp.Unmarshal()
	if err != nil {
		return nil, cerr.New(fmt.Sprintf(
			"failed to unmarshal for SellerAdminHttp response, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	var res *internalExchangeRatePb.ExchangeRateResult
	err = json.Unmarshal(body.([]byte), &res)
	if err != nil || res == nil {
		return nil, cerr.New("failed to unmarshal ExchangeRateResult",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}
	if res.Data == nil {
		return nil, cerr.New(fmt.Sprintf("exchange rate info is nil for merchantId=%v", merchantId), uint32(priceSyncPriceCalculationPb.Constant_ERROR_NOT_FOUND))
	}
	exchangeRateInfo = res.Data
	err = dm.exchangeRateCacheManager.Set(ctx, key, exchangeRateInfo)
	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to set exchange rate into cache, err=%s", err.Error()))
	}
	return exchangeRateInfo, nil
}

func (dm *ExchangeRateServiceDm) GetMerchantExchangeRate(ctx context.Context, merchantId uint64) (string, map[string]float64, error) {
	exchangeRateMap := map[string]float64{}

	exchangeRateInfo, err := dm.GetMerchantExchangeRateInfo(ctx, merchantId)
	if err != nil {
		return "", nil, err
	}
	for _, exchangeRateData := range exchangeRateInfo.ExchangeRateList {
		exchangeRateMap[exchangeRateData.GetRegion()] = exchangeRateData.GetExchangeRate()
	}
	return exchangeRateInfo.GetCurrency(), exchangeRateMap, nil
}
