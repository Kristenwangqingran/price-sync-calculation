package calcutil

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	internalMerchantConfigSettingPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

// GetCBSCDenominatorPriceRate get cbsc denominator price rate,
// = 1 - commissionRate - transactionFeeRate - ServiceFeeRate
func GetCBSCDenominatorPriceRate(ctx context.Context, cbscPriceFeeConfig *config.CBSCPriceFeeConfig, merchantConfig *internalMerchantConfigSettingPb.MerchantConfigSetting, commissionRate uint64) float64 {
	var result int64

	if cbscPriceFeeConfig == nil || !cbscPriceFeeConfig.UseServiceFeeRate {
		result = constant.PercentPrecisionBetweenMtskuAndMpsku
	} else {
		serviceFeeRate := uint64(0)
		if merchantConfig != nil && merchantConfig.ServiceFeeRate != nil {
			serviceFeeRate = merchantConfig.GetServiceFeeRate()
		}
		result = int64(constant.PercentPrecisionBetweenMtskuAndMpsku - commissionRate - cbscPriceFeeConfig.TransactionFeeRate - serviceFeeRate)
	}
	if result <= 0 {
		result = constant.PercentPrecisionBetweenMtskuAndMpsku
	}

	logging.GetLogger(ctx).Info(fmt.Sprintf(
		"calculation factor for inflated CBSCDenominatorPriceRate=%v. commissionRate=%v, cbscPriceFeeConfig=%v, merchantConfig=%v",
		result, commissionRate, cutil.JSONEncode(cbscPriceFeeConfig), cutil.JSONEncode(merchantConfig)))

	return RoundIntToFloat(result, constant.PercentPrecisionBetweenMtskuAndMpsku, 4)
}

// CalculateMpskuPrice calculate mpsku price from mtsku price
// MpskuPrice = ((MtskuPrice * exchangeRate * profitRate(marketAdjustmentRate)) + hiddenFee) / (1 - commissionRate - transactionFeeRate - ServiceFeeRate)
func CalculateMpskuPrice(mtskuPrice float64, exchangeRate float64, profitRate float64, hidePrice float64, denominatorPriceRate float64) float64 {
	decimalMtskuPrice := decimal.NewFromFloat(mtskuPrice)
	decimalExchangeRate := decimal.NewFromFloat(exchangeRate)
	decimalProfitRate := decimal.NewFromFloat(profitRate)
	decimalHidePrice := decimal.NewFromFloat(hidePrice)
	decimalDenominatorPriceRate := decimal.NewFromFloat(denominatorPriceRate)

	decimalValue := decimalMtskuPrice.Mul(decimalExchangeRate).Mul(decimalProfitRate).Add(decimalHidePrice).Div(decimalDenominatorPriceRate)
	result, _ := decimalValue.Float64()
	return result
}

// CalculateMtskuPrice calculate mtsku price from mpsku price
// MpskuPrice = ((MtskuPrice * exchangeRate * profitRate(marketAdjustmentRate)) + hiddenFee) / (1 - commissionRate - transactionFeeRate - ServiceFeeRate)
func CalculateMtskuPrice(mpskuPrice float64, exchangeRate float64, profitRate float64, hidePrice float64, denominatorPriceRate float64) float64 {
	decimalMpskuPrice := decimal.NewFromFloat(mpskuPrice)
	decimalExchangeRate := decimal.NewFromFloat(exchangeRate)
	decimalProfitRate := decimal.NewFromFloat(profitRate)
	decimalHidePrice := decimal.NewFromFloat(hidePrice)
	decimalOtherRate := decimal.NewFromFloat(denominatorPriceRate)

	decimalValue := decimalMpskuPrice.Mul(decimalOtherRate).Sub(decimalHidePrice)
	if decimalExchangeRate.IsZero() || decimalProfitRate.IsZero() {
		return 0
	}
	if decimalValue.IsNegative() {
		return 0
	}

	decimalValue = decimalValue.Div(decimalExchangeRate).Div(decimalProfitRate)
	result, _ := decimalValue.Float64()
	return result
}
