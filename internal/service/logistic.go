package service

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"

	internalFulfillmentChannelPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_fulfillment_channel.pb"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_lps.pb"
)

type LogisticService interface {
	// GetShopChannelDetailMap get shop channel detail map via cache and api.
	GetShopChannelDetailMap(ctx context.Context, shopId uint64, region string) (map[uint32]*internalFulfillmentChannelPb.FulfillmentChannelDetail, error)

	// GetShopDefaultChannel get shop default channel info.
	GetShopDefaultChannel(ctx context.Context, shopId uint64, region string) (map[string]*internalFulfillmentChannelPb.ItemLogisticsInfo, error)

	// CalcHiddenFeeForCbSip calculate hidden fee by SLS api for cb sip.
	CalcHiddenFeeForCbSip(ctx context.Context, shopID uint64, region string, itemId uint64, leafCategoryID uint64, weight uint64, enabledChannelIDList []uint32) (float64, error)

	// CalcHiddenFeeForLocalSip calculate hidden fee by SLS api for local sip.
	CalcHiddenFeeForLocalSip(
		ctx context.Context, pShopId uint64, pRegion string, pItemId uint64, pItemLeafCatId uint64, pShopPickUpLocationIds []uint64,
		aItemQueries []*model.SlsHiddenPriceQuery) ([][]*model.CalcHiddenFeeResult, error)

	// CalcShippingFeeForLocalSIP calculate shipping fee by SLS api for local sip.
	CalcShippingFeeForLocalSIP(ctx context.Context, shopID uint64, region string, itemId uint64, leafCategoryID uint64, weightInGram float64, enabledChannelIDList []uint32) (float64, error)

	// GetDummyBuyerUserID get transit warehouse user id for SLS api
	GetDummyBuyerUserID(ctx context.Context, region string) (int64, error)

	// GetSlsLocationInfoByAddressInfoBatch get sls location info map by address info
	GetSlsLocationInfoByAddressInfoBatch(ctx context.Context, region string, addressQueries []*internal.AddressQuery) (map[string]*internal.LocationInfo, error)

	// GetChannelInfoMapForRegions get channel info based on region
	GetChannelInfoMapForRegions(ctx context.Context, regionList []string) (map[string]map[uint64]*model.ChannelInfo, error)
}
