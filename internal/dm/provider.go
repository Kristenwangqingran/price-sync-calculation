package dm

import (
	"github.com/google/wire"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/a_item"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/a_shop"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/mst_shop"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/system_config"
)

var ProviderSet = wire.NewSet(
	a_shop.NewAShopDataDM,
	a_item.NewAItemDataDM,
	mst_shop.NewMSTShopDM,
	system_config.NewSystemConfigDM,
)
