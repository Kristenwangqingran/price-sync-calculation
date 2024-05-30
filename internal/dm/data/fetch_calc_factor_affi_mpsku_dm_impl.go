package data

import (
	"context"
	"strings"

	"github.com/google/wire"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/a_item"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/a_shop"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
)

var fetchCalcFactorAffiMpskuProviderSet = wire.NewSet(
	wire.Struct(new(FetchCalcFactorAffiMpskuOpts), "*"),
	NewFetchCalcFactorForAffiMpskuDm,
)

type FetchCalcFactorAffiMpskuOpts struct {
	SIPItemDataService  service.SipItemDataService
	ItemPriceService    service.ItemPriceService
	AItemDataDM         a_item.AItemDataDM
	AShopDataService    a_shop.AShopDataDM
	ShopCoreService     service.ShopCoreService
	SystemConfigService service.SystemConfigService
	ItemService         service.ItemService
	ListingSIPService   service.ListingSIPService
}

type fetchCalcFactorForAffiMpskuDm struct {
	sipItemDataService  service.SipItemDataService
	itemPriceService    service.ItemPriceService
	aItemDataDM         a_item.AItemDataDM
	aShopDataDM         a_shop.AShopDataDM
	shopCoreService     service.ShopCoreService
	systemConfigService service.SystemConfigService
	itemService         service.ItemService
	listingSIPService   service.ListingSIPService
}

func NewFetchCalcFactorForAffiMpskuDm(opts *FetchCalcFactorAffiMpskuOpts) FetchCalcFactorForAffiMpskuDm {
	return &fetchCalcFactorForAffiMpskuDm{
		sipItemDataService:  opts.SIPItemDataService,
		itemPriceService:    opts.ItemPriceService,
		aItemDataDM:         opts.AItemDataDM,
		aShopDataDM:         opts.AShopDataService,
		shopCoreService:     opts.ShopCoreService,
		systemConfigService: opts.SystemConfigService,
		itemService:         opts.ItemService,
		listingSIPService:   opts.ListingSIPService,
	}
}

func (dm *fetchCalcFactorForAffiMpskuDm) FetchCalcFactorDataForLocalSipOverseaDiscount(ctx context.Context,
	req *pb.CalcLocalSipOverseaDiscountPriceRequest) (*CalcFactorDataForAffiMpsku, error) {
	shopMappingData, err := dm.aShopDataDM.GetAShopData(ctx, uint64(req.GetAffiShopId()))
	if err != nil {
		return nil, err
	}

	affiRegion := strings.ToUpper(req.GetAffiRegion())
	primaryShopId, err := dm.shopCoreService.GetPShopIdByAShopId(ctx, uint64(req.GetAffiShopId()))
	primaryShopRegion, err := dm.shopCoreService.GetShopRegionByShopId(ctx, primaryShopId)
	if err != nil {
		return nil, err
	}

	primaryShopDetail, err := dm.shopCoreService.GetShopDetail(ctx, primaryShopId, primaryShopRegion)
	if err != nil {
		return nil, err
	}

	localPriceConfig, err := dm.systemConfigService.GetLocalPriceConfigByRegion(ctx, primaryShopRegion, affiRegion)
	if err != nil {
		return nil, err
	}

	affiItemModelIds := make([]model.ItemModelId, 0)
	for _, itemModel := range req.GetAffiItemModelIds() {
		affiItemModelIds = append(affiItemModelIds, model.ItemModelId{
			ItemId:  itemModel.GetItemId(),
			ModelId: itemModel.GetModelId(),
		})
	}

	itemModelMapping, err := dm.sipItemDataService.GetPrimaryItemModelIdsByAffiItemModelIds(ctx, primaryShopId, affiItemModelIds)
	primaryItemModelIds := make([]model.ItemModelId, 0)
	for _, itemModel := range itemModelMapping {
		primaryItemModelIds = append(primaryItemModelIds, itemModel)
	}

	affiItemIds := make([]uint64, 0)
	for _, itemModelId := range affiItemModelIds {
		affiItemIds = append(affiItemIds, itemModelId.ItemId)}

	aItemData, err := dm.aItemDataDM.GetAItemDataBatch(ctx, primaryShopId, affiItemIds)
	if err != nil {
		return nil, err
	}

	pShopItemInfoList, err := dm.listingSIPService.GetPrimaryItemIdByAffiItemIds(ctx, uint64(req.GetAffiShopId()), affiItemIds)
	if err != nil {
		return nil, err
	}

	aItemIdToPItemIdMap := make(map[uint64]uint64)
	primaryItemIds := make([]uint64, 0)
    for _, pShopItemInfo := range pShopItemInfoList {
    	primaryItemIds = append(primaryItemIds, pShopItemInfo.GetPitemId())
    	aItemIdToPItemIdMap[pShopItemInfo.GetAitemId()] = pShopItemInfo.GetPitemId()
	}

	primaryItemData, err := dm.sipItemDataService.GetPrimaryItemDataBatch(ctx, primaryShopId, primaryItemIds)
	if err != nil {
		return nil, err
	}

	primaryOriginPrices, err := dm.itemPriceService.GetOriginPriceBatch(ctx, primaryShopRegion, primaryItemModelIds)
	if err != nil {
		return nil, err
	}
	affiOriginPrices, err := dm.itemPriceService.GetOriginPriceBatch(ctx, affiRegion, affiItemModelIds)
	if err != nil {
		return nil, err
	}

	primaryItemProductInfos := dm.itemService.GetProductInfoMapForSameRegion(ctx, primaryItemIds, primaryShopRegion)
	primaryItemEnabledChannelIds := dm.itemService.GetItemEnableChannelIdsMap(ctx, primaryItemIds, primaryItemProductInfos)

	affiItemProductInfos := dm.itemService.GetProductInfoMapForSameRegion(ctx, affiItemIds, affiRegion)
	affiItemEnabledChannelIds := dm.itemService.GetItemEnableChannelIdsMap(ctx, affiItemIds, affiItemProductInfos)

	primaryItemLeafCategoryIDMap := dm.itemService.GetItemLeafCatIdMap(ctx, primaryItemIds, primaryItemProductInfos)
	affiItemLeafCategoryIDMap := dm.itemService.GetItemLeafCatIdMap(ctx, affiItemIds, affiItemProductInfos)

	channelWhitelist, err := dm.systemConfigService.GetChannelWhitelist(ctx)
	if err != nil {
		return nil, err
	}

	return &CalcFactorDataForAffiMpsku{
		PrimaryShopId: primaryShopId,
		PrimaryRegion: primaryShopRegion,
		AffiRegion:    affiRegion,
		AffiShopId:    req.GetAffiShopId(),

		LocalPriceConfig:  localPriceConfig,
		ShopMargin:        int64(shopMappingData.GetShopMargin()),
		ChannelWhitelist:  channelWhitelist,
		PrimaryShopDetail: primaryShopDetail,

		PrimaryOriginPrices:     primaryOriginPrices,
		AffiOriginPrices:        affiOriginPrices,
		AItemData:               aItemData,
		AItemIdToPItemIdMapping: aItemIdToPItemIdMap,
		PrimaryItemData:         primaryItemData,
		ItemModelMapping:        itemModelMapping,

		AffiItemEnabledChannelIds:    affiItemEnabledChannelIds,
		PrimaryItemEnabledChannelIds: primaryItemEnabledChannelIds,
		PrimaryItemLeafCategoryIDMap: primaryItemLeafCategoryIDMap,
		AffiItemLeafCategoryIDMap:    affiItemLeafCategoryIDMap,
	}, nil
}
