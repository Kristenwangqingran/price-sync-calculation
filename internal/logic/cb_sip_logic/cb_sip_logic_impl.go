package cb_sip_logic

import (
	"github.com/google/wire"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/a_item"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/a_shop"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/mst_shop"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/system_config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/account_service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/edit_item_price_allow_list"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/factors"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/hpfn_config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_v2_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
)

type CbSipLogicImpl struct {
	factorsRepo          factors.CalculationFactorsRepo
	sipV2Repo            sip_v2_db.SipV2Repo
	sipRepo              sip_db.SipRepo
	shopCoreService      service.ShopCoreService
	aShopDataDM          a_shop.AShopDataDM
	aItemDataDM          a_item.AItemDataDM
	mstShopDM            mst_shop.MSTShopDM
	systemConfigDM       system_config.SystemConfigDM
	integratedFeeService service.OrderAccountIntegratedFeeService
	ibsService           spex.ItemBusiness

	hpfnConfigRepo             hpfn_config.HpfnConfigRepo
	allowEditItemPriceShopRepo edit_item_price_allow_list.PShopWhiteListToEditItemPriceRepo

	shopMerchantService  service.ShopMerchantService
	exchangeRateService  service.ExchangeRateService
	logisticService      service.LogisticService
	listingUploadService spex.ListingUploadService
	accountServiceRepo   account_service.AccountServiceRepo
	cache                cache.CommonCache

	priceBusinessService service.PriceBusinessService
}

type CbSipLogicOpts struct {
	FactorsRepo          factors.CalculationFactorsRepo
	SipV2Repo            sip_v2_db.SipV2Repo
	SipRepo              sip_db.SipRepo
	ShopCoreService      service.ShopCoreService
	AShopDataDM          a_shop.AShopDataDM
	AItemDataDM          a_item.AItemDataDM
	MSTShopDM            mst_shop.MSTShopDM
	SystemConfigDM       system_config.SystemConfigDM
	IntegratedFeeService service.OrderAccountIntegratedFeeService
	IbsService           spex.ItemBusiness

	HpfnConfigRepo             hpfn_config.HpfnConfigRepo
	AllowEditItemPriceShopRepo edit_item_price_allow_list.PShopWhiteListToEditItemPriceRepo

	ShopMerchantService  service.ShopMerchantService
	ExchangeRateService  service.ExchangeRateService
	LogisticService      service.LogisticService
	ListingUploadService spex.ListingUploadService
	AccountServiceRepo   account_service.AccountServiceRepo

	PriceBusinessService service.PriceBusinessService

	Cache cache.CommonCache
}

func NewCbSipLogicImpl(opts *CbSipLogicOpts) *CbSipLogicImpl {
	return &CbSipLogicImpl{
		factorsRepo:                opts.FactorsRepo,
		sipV2Repo:                  opts.SipV2Repo,
		sipRepo:                    opts.SipRepo,
		shopCoreService:            opts.ShopCoreService,
		aShopDataDM:                opts.AShopDataDM,
		aItemDataDM:                opts.AItemDataDM,
		mstShopDM:                  opts.MSTShopDM,
		systemConfigDM:             opts.SystemConfigDM,
		integratedFeeService:       opts.IntegratedFeeService,
		ibsService:                 opts.IbsService,
		hpfnConfigRepo:             opts.HpfnConfigRepo,
		shopMerchantService:        opts.ShopMerchantService,
		exchangeRateService:        opts.ExchangeRateService,
		logisticService:            opts.LogisticService,
		cache:                      opts.Cache,
		allowEditItemPriceShopRepo: opts.AllowEditItemPriceShopRepo,
		listingUploadService:       opts.ListingUploadService,
		accountServiceRepo:         opts.AccountServiceRepo,
		priceBusinessService:       opts.PriceBusinessService,
	}
}

var ProviderSet = wire.NewSet(
	wire.Struct(new(CbSipLogicOpts), "*"),
	NewCbSipLogicImpl,
)
