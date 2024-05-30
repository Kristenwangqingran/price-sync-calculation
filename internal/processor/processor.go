package processor

import (
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/calculate"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/data"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/servicesetup"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewCalculationServiceImpl,
)

type CalculationServiceImpl struct {
	FetchCalcFactorForMtskuAndMpskuDm data.FetchCalcFactorForMtskuAndMpskuDm
	CalculateMtskuAndMpskuDm          calculate.CalculateMtskuAndMpskuDm
	FetchCalcFactorForAffiMpskuDm     data.FetchCalcFactorForAffiMpskuDm
	CalculateAffiMpskuDm              calculate.CalculateAffiMpskuDm
	cbscLogic                         logic.CbscLogic
	localSipLogic                     logic.LocalSipLogic
	cbsipLogic                        logic.CbSipLogic
	commonSipLogic                    logic.CommonSIPLogic
	currencyConvertLogic              logic.CurrencyConvertLogic
	AsyncDataLogic                    servicesetup.AsyncData
}

// NewCalculationServiceImpl returns an implementation of CalculationService
func NewCalculationServiceImpl(
	fetchCalcFactorForMtskuAndMpskuDm data.FetchCalcFactorForMtskuAndMpskuDm,
	calculateMtskuAndMpskuDm calculate.CalculateMtskuAndMpskuDm,
	fetchCalcFactorForAffiMpskuDm data.FetchCalcFactorForAffiMpskuDm,
	calculateAffiMpskuDm calculate.CalculateAffiMpskuDm,
	cbscLogic logic.CbscLogic,
	localSipLogic logic.LocalSipLogic,
	cbsipLogic logic.CbSipLogic,
	commonSipLogic logic.CommonSIPLogic,
	currencyConvertLogic logic.CurrencyConvertLogic,
	AsyncDataLogic servicesetup.AsyncData,
) *CalculationServiceImpl {
	return &CalculationServiceImpl{
		FetchCalcFactorForMtskuAndMpskuDm: fetchCalcFactorForMtskuAndMpskuDm,
		CalculateMtskuAndMpskuDm:          calculateMtskuAndMpskuDm,
		FetchCalcFactorForAffiMpskuDm:     fetchCalcFactorForAffiMpskuDm,
		CalculateAffiMpskuDm:              calculateAffiMpskuDm,
		cbscLogic:                         cbscLogic,
		localSipLogic:                     localSipLogic,
		cbsipLogic:                        cbsipLogic,
		commonSipLogic:                    commonSipLogic,
		currencyConvertLogic:              currencyConvertLogic,
		AsyncDataLogic:                    AsyncDataLogic,
	}
}
