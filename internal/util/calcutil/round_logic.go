package calcutil

import (
	"fmt"
	"math"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"

	"github.com/shopspring/decimal"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/maths"
)

// copy from sip-goservice, need refactor later

func GetMiniPriceValueBasedOnCurrency(roundPlace int32) int64 {
	return int64(math.Pow10(-1*int(roundPlace)) * constant.PricePrecision)
}

func DBPriceRoundNearest(currency string, dbPrice int64) int64 {
	realPrice := ToRealPrice(dbPrice)
	return ToDBPrice(PriceRoundNearest(currency, realPrice))
}

func PriceRoundNearest(currency string, price float64) float64 {
	currencySetting := config.GetSIPCurrencyCommonConf().GetByCurrency(currency)
	return RoundWithPrecision(price, currencySetting.Precision)
}

func RoundWithPrecision(value float64, decimalPlaces int) float64 {
	if decimalPlaces >= 0 {
		return float64(int(math.Round(value/(math.Pow10(decimalPlaces))) * (math.Pow10(decimalPlaces))))
	}
	return math.Round(value*(math.Pow10(maths.IntAbs(decimalPlaces)))) / (math.Pow10(maths.IntAbs(decimalPlaces)))
}

func RoundUintToFloat(input uint64, precision int32, roundPlace int32) float64 {
	inputNum := decimal.NewFromFloat(float64(input))
	precisionNum := decimal.NewFromFloat(float64(precision))
	decimalValue := inputNum.Div(precisionNum)
	result, _ := decimalValue.Round(roundPlace).Float64()
	return result
}

func RoundFloatToInt(input float64, precision int32, roundPlace int32) int64 {
	inputNum := decimal.NewFromFloat(input).Round(roundPlace)
	precisionNum := decimal.NewFromFloat(float64(precision))
	decimalValue := inputNum.Mul(precisionNum)
	return decimalValue.Round(roundPlace).IntPart()
}

func RoundIntToFloat(input int64, precision int32, roundPlace int32) float64 {
	inputNum := decimal.NewFromFloat(float64(input))
	precisionNum := decimal.NewFromFloat(float64(precision))
	decimalValue := inputNum.Div(precisionNum)
	result, _ := decimalValue.Round(roundPlace).Float64()
	return result
}

func RoundUp(minValue int, roundSize int, actualValue float64) int {
	if actualValue <= float64(minValue) {
		return minValue
	} else {
		normalizeActualValue := int(actualValue)
		if normalizeActualValue%roundSize == 0 {
			return (normalizeActualValue / roundSize) * roundSize
		} else {
			return (normalizeActualValue/roundSize + 1) * roundSize
		}
	}
}

func PriceRoundUpByCountry(country string, price float64) float64 {
	currency := config.GetSIPRegionCommonConf().GetByRegion(country).Currency
	if len(currency) == 0 {
		ulog.DefaultLogger().Error("currency for region not exist", ulog.String("region", country))
		return price
	}
	currencySetting := config.GetSIPCurrencyCommonConf().GetByCurrency(currency)

	result := PriceRoundUp(currencySetting, price)
	ulog.DefaultLogger().Info(fmt.Sprintf("PriceRoundUpByCountry: %f => %f", price, result),
		ulog.String("region", country), ulog.String("currency", currency))

	return result
}

func PriceRoundUp(currencySetting config.CurrencyCommonSetting, price float64) float64 {
	if currencySetting.UseSpecialRoundup {
		return SpecialPriceRoundUp(currencySetting, price)
	}
	return ceilWithPrecision(price, currencySetting.Precision)
}

func ceilWithPrecision(value float64, decimalPlaces int) float64 {
	if decimalPlaces >= 0 {
		return float64(int(math.Ceil(value/(math.Pow10(decimalPlaces))) * (math.Pow10(decimalPlaces))))
	}
	return math.Ceil(value*(math.Pow10(maths.IntAbs(decimalPlaces)))) / (math.Pow10(maths.IntAbs(decimalPlaces)))
}

func SpecialPriceRoundUp(currencySetting config.CurrencyCommonSetting, price float64) float64 {
	decimalPlaces := currencySetting.Precision
	epsilon := math.Pow10(decimalPlaces - 2)
	priceWithoutDecimalPoint := math.Floor(price)
	priceDecimalPart := price - priceWithoutDecimalPoint
	if priceDecimalPart < epsilon {
		return priceWithoutDecimalPoint
	} else if priceDecimalPart < 0.5+epsilon {
		return priceWithoutDecimalPoint + 0.5
	} else if priceDecimalPart < 0.9+epsilon {
		return priceWithoutDecimalPoint + 0.9
	} else if priceDecimalPart < 0.99+epsilon {
		return priceWithoutDecimalPoint + 0.99
	} else {
		return priceWithoutDecimalPoint + 1
	}
}

type cmpResult int

const (
	AmtLt cmpResult = 1 // <, LessThan
	AmtEq cmpResult = 2 // =, Equal
	AmtGt cmpResult = 3 // >, GreaterThan
)

// Works for price, weight etc.
// https://jira.shopee.io/browse/SPSM-12236
func RoundCompare(x1, x2 float64) cmpResult {
	if x1 == math.MaxFloat64 || x2 == math.MaxFloat64 {
		switch {
		case x1 < x2:
			return AmtLt
		case x1 == x2:
			return AmtEq
		default:
			return AmtGt
		}
	}
	i1 := ToDBPrice(x1)
	i2 := ToDBPrice(x2)
	switch {
	case i1 < i2:
		return AmtLt
	case i1 == i2:
		return AmtEq
	default:
		return AmtGt
	}
}

func Gt(x1, x2 float64) bool {
	res := RoundCompare(x1, x2)
	return res == AmtGt
}
