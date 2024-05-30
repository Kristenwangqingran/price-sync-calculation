package service

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	logisticServiceProvider,
	NewShopMerchantService,
	NewExchangeRateService,
	NewItemService,
	NewItemPriceService,
	NewMerchantConfigService,
	NewOrderAccountIntegratedFeeService,
	NewShopCoreService,
	NewSipItemDataService,
	NewSystemConfigService,
	NewPriceBusinessService,
	NewListingSIPService,
	wire.Bind(new(PriceBusinessService), new(*PriceBusinessServiceDm)),
)
