package calcutil

import (
	"math"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
)

// copy from sip-goservice, need refactor later

func ToRealPrice(input int64) float64 {
	if input == -1 || input == 0 {
		return float64(input)
	}

	return float64(input) / constant.PricePrecision
}

func ToRealPect(dbPect int) float64 {
	return float64(dbPect) / float64(constant.DbPectInflationFactor)
}

func ToDBRatio(ratio float64) int32 {
	return int32(ratio * constant.DbPectInflationFactor)
}

func ToDBPrice(realPrice float64) int64 {
	if realPrice == 0 || realPrice == -1 {
		return int64(realPrice)
	}
	return int64(math.Round(realPrice * float64(constant.PricePrecision)))
}

func DbWeightToGram(weightDb int64) float64 {
	return float64(weightDb*1000) / float64(constant.WeightPrecision)
}

func GramToDbWeight(gram float64) int64 {
	return int64(gram * float64(constant.WeightPrecision) / 1000)
}

func ToRealRatio(ratio int64) float64 {
	return float64(ratio) / float64(constant.DbPectInflationFactor)
}
