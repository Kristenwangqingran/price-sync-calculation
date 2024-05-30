package db

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewItemMapDB,
	NewAItemDB,
	NewLocalHiddenPriceConfigDB,
	NewLocalShippingFeeConfigDB,
	NewMerchantConfigDB,
	NewMskuDB,
	NewShopMapDB,
	NewMSTShopDB,
	NewSystemConfigDB,
)
