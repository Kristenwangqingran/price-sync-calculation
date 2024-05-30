package config

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/common/uniconfig"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type LocalSipPriceConfig struct {
	WhitelistForSlsHiddenPriceMap            map[string][]uint64                  `json:"whitelist_for_sls_hidden_price_map,omitempty"`
	SlsTransitWarehouseDeliveryAddressIdList []*TransitWarehouseDeliveryAddressId `json:"sls_transit_warehouse_delivery_address_id_list" json:"sls_transit_warehouse_delivery_address_id_list,omitempty"`
}

type TransitWarehouseDeliveryAddressId struct {
	PRegion           string `json:"p_region"`
	PChannelId        uint32 `json:"p_channel_id"`
	DeliveryAddressId uint64 `json:"delivery_address_id"`
}

func onLocalSipPriceConfigUpdate(e uniconfig.Event) {
	rawNewConfig, err := e.New()
	if err != nil {
		logging.GetLogger(context.Background()).Warn("error getting updated LocalSipPriceConfig value from uniconfig.Event", ulog.Error(err))
		return
	}

	newConfig, ok := rawNewConfig.(*LocalSipPriceConfig)
	if !ok {
		logging.GetLogger(context.Background()).Warn("new config is not a LocalSipPriceConfig",
			ulog.String("newVal", cutil.JSONEncode(rawNewConfig)))
		return
	}

	confVal.LocalSipPriceConfig = newConfig

	logging.GetLogger(context.Background()).Info("LocalSipPriceConfig is updated", ulog.String("LocalSipPriceConfig", cutil.JSONEncode(newConfig)))
}

func GetSlsTransitWarehouseDeliveryAddressId(pChannelId uint32, pRegion string) *uint64 {
	if confVal == nil || confVal.LocalSipPriceConfig == nil {
		return nil
	}

	for _, r := range confVal.LocalSipPriceConfig.SlsTransitWarehouseDeliveryAddressIdList {
		if r.PChannelId == pChannelId && r.PRegion == pRegion {
			return proto.Uint64(r.DeliveryAddressId)
		}
	}
	return nil
}

// UseSlsHiddenPriceOnSlsMode we add whitelist toggle here for live testing
// if exists whitelist, then when SLS toggle on, only use SLS calculate when a shop is in the whitelist.
// if not exist whitelist, then depend on SLS toggle directly.
func UseSlsHiddenPriceOnSlsMode(aShopId uint64, pRegion string, aRegion string) bool {
	if confVal == nil || confVal.LocalSipPriceConfig == nil {
		return false
	}

	key := fmt.Sprintf("%s-%s", pRegion, aRegion)
	if len(confVal.LocalSipPriceConfig.WhitelistForSlsHiddenPriceMap) == 0 ||
		len(confVal.LocalSipPriceConfig.WhitelistForSlsHiddenPriceMap[key]) == 0 {
		return true
	}

	for _, shopId := range confVal.LocalSipPriceConfig.WhitelistForSlsHiddenPriceMap[key] {
		if shopId == aShopId {
			return true
		}
	}

	return false
}
