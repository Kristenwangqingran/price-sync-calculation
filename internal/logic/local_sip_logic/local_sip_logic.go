package local_sip_logic

import (
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/factors"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"

	"github.com/google/wire"
)

type LocalSipLogicImpl struct {
	factors factors.CalculationFactorsRepo
	sipRepo sip_db.SipRepo

	priceBusinessService service.PriceBusinessService
}

func NewLocalSipLogicImpl(factorsRepo factors.CalculationFactorsRepo, sipRepo sip_db.SipRepo, priceBusinessService service.PriceBusinessService) *LocalSipLogicImpl {
	return &LocalSipLogicImpl{
		factors:              factorsRepo,
		sipRepo:              sipRepo,
		priceBusinessService: priceBusinessService,
	}
}

var ProviderSet = wire.NewSet(
	NewLocalSipLogicImpl,
)
