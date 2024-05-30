package data

import (
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
)

type CalcFactorDataForAffiMpsku struct {
	PrimaryShopId    uint64
	PrimaryRegion    string
	AffiRegion       string
	AffiShopId       int64
	ItemModelMapping map[model.ItemModelId]model.ItemModelId

	PrimaryOriginPrices map[model.ItemModelId]int64
	AffiOriginPrices    map[model.ItemModelId]int64
	LocalPriceConfig    *model.CommonPriceConfig
	ShopMargin          int64
	PrimaryItemData     map[uint64]service.PrimaryItemData
	AItemData           map[uint64]*internal.AItemData
	AItemIdToPItemIdMapping map[uint64]uint64

	// used for calc hiddenPrice/ShippingFee from SLS
	PrimaryShopDetail            *model.ShopDetail
	ChannelWhitelist             []int64
	PrimaryItemEnabledChannelIds map[uint64]*service.EnabledChannelIdsInfo
	AffiItemEnabledChannelIds    map[uint64]*service.EnabledChannelIdsInfo
	PrimaryItemLeafCategoryIDMap map[uint64]*service.ItemLeafCategoryInfo
	AffiItemLeafCategoryIDMap    map[uint64]*service.ItemLeafCategoryInfo
}
