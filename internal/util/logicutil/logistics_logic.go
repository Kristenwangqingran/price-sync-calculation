package logicutil

import (
	marketplaceLogisticsShopChannelsPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/marketplace_logistics_shop_channels.pb"
)

func IsChannelLocal(channelTag int32) bool {
	if channelTag == int32(marketplaceLogisticsShopChannelsPb.Channel_LOCAL_WMS) || channelTag == int32(marketplaceLogisticsShopChannelsPb.Channel_LOCAL_NO_WMS) {
		return true
	}

	return false
}
