package calcutil

func CalcAffiDBPriceForCbSip(basePrice int64, aRegion string, exchangeRate,
	priceRatio, ratio, countryMargin, shopMargin, itemMargin, hiddenPrice, finalFee float64) int64 {
	realPrimPrice := ToRealPrice(basePrice)
	affiPrice := calcAffiPriceForCbSip(realPrimPrice, exchangeRate, priceRatio, ratio,
		countryMargin, shopMargin, itemMargin, hiddenPrice, finalFee)
	return ToDBPrice(PriceRoundUpByCountry(aRegion, affiPrice))
}

func calcAffiPriceForCbSip(primPrice, exchangeRate,
	priceRatio, ratio, countryMargin, shopMargin, itemMargin, hiddenPrice, finalFee float64) float64 {
	margin := calFinalMargin(countryMargin, shopMargin, itemMargin)
	return (primPrice*priceRatio*exchangeRate + hiddenPrice) * ratio * margin * finalFee
}

func calFinalMargin(countryMargin float64, shopMargin float64, itemMargin float64) float64 {
	margin := 1 + countryMargin + shopMargin + itemMargin
	if margin <= 0 {
		margin = 1
	}
	return margin
}
