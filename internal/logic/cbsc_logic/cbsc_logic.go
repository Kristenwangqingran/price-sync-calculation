package cbsc_logic

import (
	"github.com/google/wire"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/factors"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
)

type CbscLogicImpl struct {
	shopMerchantService   service.ShopMerchantService
	merchantConfigService service.MerchantConfigService
	factorsRepo           factors.CalculationFactorsRepo
}

type CbscLogicOpts struct {
	ShopMerchantService   service.ShopMerchantService
	MerchantConfigService service.MerchantConfigService
	FactorsRepo           factors.CalculationFactorsRepo
}

func NewCbscLogicImpl(deps *CbscLogicOpts) *CbscLogicImpl {
	return &CbscLogicImpl{
		shopMerchantService:   deps.ShopMerchantService,
		merchantConfigService: deps.MerchantConfigService,
		factorsRepo:           deps.FactorsRepo,
	}
}

var ProviderSet = wire.NewSet(
	wire.Struct(new(CbscLogicOpts), "*"),
	NewCbscLogicImpl,
)
