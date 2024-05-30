package spex

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewItemBusiness,
	GetAccountAddress,
	NewMarketplaceOrderAccountingIntegratedFee,
	GetOrderProcessing,
	GetPriceBasic,
	NewShopCore,
	NewShopMerchant,
	NewAccountCore,
	NewListinguploadServiceImpl,
	NewLogisticsShopChannels,
	NewPriceBusiness,
	NewListingSIPProxy,
	NewItemDiscount,
)
