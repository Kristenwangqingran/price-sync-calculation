package calcutil

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

func CalculateAffiPriceForLocalSip(ctx context.Context, pPrice float64, aRealWeight float64, itemMargin float64, shopMargin float64, initHiddenPrice float64, shippingFee float64, localPriceConfig *model.CommonPriceConfig, overseaDiscountRate *float64) float64 {
	if localPriceConfig == nil {
		return pPrice
	}

	if aRealWeight <= 0 {
		return 0
	}

	itemMargin = GetItemMargin(itemMargin)

	shopMargin = GetShopMargin(shopMargin)

	countryMargin := GetLocalSipCountryMargin(localPriceConfig)

	exchangeRate := GetLocalSipExchangeRate(localPriceConfig)

	overseaRate := float64(0)
	if overseaDiscountRate == nil || *overseaDiscountRate < 0 || *overseaDiscountRate > 1 {
		overseaRate = 0
	} else {
		overseaRate = *overseaDiscountRate
	}

	aPrice := (pPrice*(1-overseaRate)/exchangeRate + initHiddenPrice + shippingFee) * countryMargin * shopMargin * itemMargin

	logging.GetLogger(ctx).Info(fmt.Sprintf("%v = (%v * (1 - %v) / %v + %v + %v) * %v * %v * %v",
		aPrice, pPrice, overseaRate, exchangeRate, initHiddenPrice, shippingFee, countryMargin, shopMargin, itemMargin))

	return aPrice
}

func GetItemMargin(itemMargin float64) float64 {
	if itemMargin <= 0 {
		itemMargin = 1
	}
	return itemMargin
}

func GetShopMargin(shopMargin float64) float64 {
	if shopMargin <= 0 {
		shopMargin = 1
	}
	return shopMargin
}

func GetLocalSipCountryMargin(localPriceConfig *model.CommonPriceConfig) float64 {
	var countryMargin float64
	if localPriceConfig.Buffer == nil || *localPriceConfig.Buffer <= 0 {
		countryMargin = 1
	} else {
		countryMargin = *localPriceConfig.Buffer
	}

	return countryMargin
}

func GetLocalSipExchangeRate(localPriceConfig *model.CommonPriceConfig) float64 {
	var exchangeRate float64
	if localPriceConfig.ExchangeRate == nil || *localPriceConfig.ExchangeRate == 0 {
		exchangeRate = 1
	} else {
		exchangeRate = *localPriceConfig.ExchangeRate
	}
	return exchangeRate
}
