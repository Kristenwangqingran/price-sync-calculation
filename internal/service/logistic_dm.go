package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/wire"
	"google.golang.org/protobuf/proto"

	commonCache "git.garena.com/shopee/common/cache"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/http"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	internalFulfillmentChannelPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_fulfillment_channel.pb"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_lps.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/marketplace_order_processing_cb_collection_api.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/cidutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/convutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/httpcliutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

var logisticServiceProvider = wire.NewSet(
	wire.Struct(new(LogisticServiceOpts), "*"),
	NewLogisticService,
)

type LogisticServiceOpts struct {
	FulfillmentChannelDetailCacheManager cache.FulfillmentChannelDetailCacheManager
	HttpCli                              httpcliutil.HTTPCli
	OrderProcessing                      spex.OrderProcessing
	AccountAddress                       spex.AccountAddress
	LogisticsShopChannels                spex.LogisticsShopChannels
	LogisticCacheManager                 cache.LogisticCacheManager
	CommonCacheManager                   cache.CommonCache
}

type LogisticServiceDm struct {
	fulfillmentChannelDetailCacheManager cache.FulfillmentChannelDetailCacheManager
	httpCli                              httpcliutil.HTTPCli

	orderProcessing       spex.OrderProcessing
	accountAddress        spex.AccountAddress
	logisticsShopChannels spex.LogisticsShopChannels
	logisticCacheManager  cache.LogisticCacheManager
	commonCacheManager    cache.CommonCache
}

func NewLogisticService(opts *LogisticServiceOpts) LogisticService {
	return &LogisticServiceDm{
		fulfillmentChannelDetailCacheManager: opts.FulfillmentChannelDetailCacheManager,
		logisticCacheManager:                 opts.LogisticCacheManager,
		httpCli:                              opts.HttpCli,
		orderProcessing:                      opts.OrderProcessing,
		accountAddress:                       opts.AccountAddress,
		logisticsShopChannels:                opts.LogisticsShopChannels,
		commonCacheManager:                   opts.CommonCacheManager,
	}
}

func (dm *LogisticServiceDm) GetShopChannelDetailMap(ctx context.Context, shopId uint64, region string) (map[uint32]*internalFulfillmentChannelPb.FulfillmentChannelDetail, error) {
	shopChannelList, err := dm.getShopFulfillmentChannelDetailList(ctx, shopId, region)
	if err != nil {
		return nil, err
	}

	shopChannelMap := make(map[uint32]*internalFulfillmentChannelPb.FulfillmentChannelDetail)
	for _, c := range shopChannelList {
		shopChannelMap[c.GetChannelId()] = c
	}
	return shopChannelMap, nil
}

// getShopFulfillmentChannelDetailList get shop channel detail list via cache and api
func (dm *LogisticServiceDm) getShopFulfillmentChannelDetailList(ctx context.Context, shopId uint64, region string) ([]*internalFulfillmentChannelPb.FulfillmentChannelDetail, error) {
	fulfillmentChannelDetailCacheKey := dm.fulfillmentChannelDetailCacheManager.Key(shopId)
	fulfillmentChannelDetailList, err := dm.fulfillmentChannelDetailCacheManager.Get(ctx, fulfillmentChannelDetailCacheKey)
	if err == nil && len(fulfillmentChannelDetailList) > 0 { // cache hit, then use cache data directly
		return fulfillmentChannelDetailList, nil
	}
	if err != nil && err != commonCache.ErrCacheMiss {
		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to getShopFulfillmentChannelDetailList from cache, key=%s, err=%s",
			fulfillmentChannelDetailCacheKey, err.Error()))
	}

	fulfillmentChannelDetailList, err = dm.getShopFulfillmentChannelDetailListFromApi(ctx, shopId, region)
	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf("failed to getShopFulfillmentChannelDetailListFromApi, err=%s", err.Error()))
		return nil, err
	}

	err = dm.fulfillmentChannelDetailCacheManager.Set(ctx, fulfillmentChannelDetailCacheKey, fulfillmentChannelDetailList)
	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf("failed to set fulfillmentChannelDetailCache, err=%s", err.Error()))
		return nil, err
	}

	return fulfillmentChannelDetailList, nil
}

func (dm *LogisticServiceDm) getShopFulfillmentChannelDetailListFromApi(ctx context.Context, shopId uint64, region string) ([]*internalFulfillmentChannelPb.FulfillmentChannelDetail, error) {
	param := map[string]string{
		"shop_id":   strconv.FormatUint(shopId, 10),
		"region_id": region,
	}
	header := map[string]string{
		"shop-id":   strconv.FormatUint(shopId, 10),
		"region-id": region,
	}

	ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to fill CID, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	resp, err := http.GetFulfillmentHttp(dm.httpCli, ctxWithCID, constant.GetShopChannelListCmd, header, param)
	if err != nil {
		return nil, err
	}
	body, err := resp.Unmarshal()
	if err != nil {
		return nil, cerr.New(fmt.Sprintf(
			"failed to unmarshal GetFulfillmentHttp response, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	type Body struct {
		List []*internalFulfillmentChannelPb.FulfillmentChannelDetail `json:"list"`
	}
	type Response struct {
		Data    *Body   `json:"data"`
		Code    *int32  `json:"code"`
		Message *string `json:"message"`
	}
	rsp := Response{}
	err = json.Unmarshal(body.([]byte), &rsp)
	if err != nil || rsp.Code == nil || *rsp.Code != 0 {
		errMsg := "failed in GetFulfillmentHttp"
		if err != nil {
			errMsg += fmt.Sprintf(", err=%s", err.Error())
		}
		if rsp.Code != nil {
			errMsg += fmt.Sprintf(", code=%d", *rsp.Code)
		}
		if rsp.Message != nil {
			errMsg += fmt.Sprintf(", msg=%s", *rsp.Message)
		}
		return nil, cerr.New(errMsg, uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
	}

	return rsp.Data.List, nil
}

func (dm *LogisticServiceDm) GetSlsLocationInfoByAddressInfoBatch(ctx context.Context, region string, addressQueries []*internal.AddressQuery) (map[string]*internal.LocationInfo, error) {
	if len(addressQueries) == 0 {
		return nil, nil
	}

	locationInfoByUniqIdMap := make(map[string]*internal.LocationInfo)
	batchSize := config.GetBatchConfig().MaxBatchSizeForSlsBatchGetSlsLocationInfo

	for start := 0; start < len(addressQueries); {
		end := start + int(batchSize)
		if end > len(addressQueries) {
			end = len(addressQueries)
		}

		iRes, iErr := dm.getSlsLocationInfoByAddressInfo(ctx, region, addressQueries[start:end])
		if iErr != nil {
			return nil, iErr
		}

		for uniqId, r := range iRes {
			locationInfoByUniqIdMap[uniqId] = r
		}

		start = end
	}

	return locationInfoByUniqIdMap, nil
}

func (dm *LogisticServiceDm) getSlsLocationInfoByAddressInfo(ctx context.Context, region string, addressQueries []*internal.AddressQuery) (map[string]*internal.LocationInfo, error) {
	domainUrl := config.GetShopeeRegionDomainUrl(region)
	req := &internal.BatchGetSlsLocationInfoRequest{
		AddressList: addressQueries,
	}

	ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to fill CID, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	resp, err := http.PostSlsLpsHttp(dm.httpCli, ctxWithCID, domainUrl, constant.BatchGetSlsLocationInfoCmd, req, config.GetHTTPApiConfig().SlsLpsTimeoutMs)
	if err != nil {
		return nil, err
	}
	respUnmarshalled, err := resp.Unmarshal()
	if err != nil {
		return nil, cerr.New(fmt.Sprintf(
			"failed to unmarshal BatchGetSlsLocationInfoCmd response, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	var rsp *internal.BatchGetSlsLocationInfoResponse
	err = json.Unmarshal(respUnmarshalled.([]byte), &rsp)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to unmarshal for BatchGetSlsLocationInfoResponse, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}
	logging.GetLogger(ctx).Debug(fmt.Sprintf("getSlsLocationInfoByAddressInfo successfully, "+
		"cmd=%s, request=%v, response=%v",
		constant.BatchGetSlsLocationInfoCmd, cutil.JSONEncode(req), cutil.JSONEncode(rsp)))

	if rsp.GetRetcode() != 0 {
		return nil, cerr.New(fmt.Sprintf("failed to batch get sls location info, retCode=%d, msg=%s, detail=%s",
			rsp.GetRetcode(), rsp.GetMessage(), rsp.GetDetail()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
	}
	if len(req.GetAddressList()) != len(rsp.GetData().GetLocationInfo()) {
		return nil, cerr.New(fmt.Sprintf("the result length is not same like request, req=%v, request length=%d, result length=%d",
			cutil.JSONEncode(req), len(req.GetAddressList()), len(rsp.GetData().GetLocationInfo())),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
	}

	ret := make(map[string]*internal.LocationInfo)
	for _, info := range rsp.GetData().GetLocationInfo() {
		ret[info.GetUniqueId()] = info
	}

	return ret, nil
}

func (dm *LogisticServiceDm) GetShopDefaultChannel(ctx context.Context, shopId uint64, region string) (map[string]*internalFulfillmentChannelPb.ItemLogisticsInfo, error) {
	domainUrl := config.GetShopeeRegionDomainUrl(region)
	body := map[string]interface{}{
		"country": region,
		"shopid":  shopId,
	}

	ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to fill CID, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	resp, err := http.PostSlsOldUrlHttp(dm.httpCli, ctxWithCID, domainUrl, constant.DefaultItemLogisticInfoCmd, body, config.GetHTTPApiConfig().ShopLogisticsTimeoutMs)
	if err != nil {
		return nil, err
	}
	respUnmarshalled, err := resp.Unmarshal()
	if err != nil {
		return nil, cerr.New(fmt.Sprintf(
			"failed to unmarshal GetShopDefaultChannel response, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	type ResponseStructure struct {
		DefaultItemLogisticInfo map[string]*internalFulfillmentChannelPb.ItemLogisticsInfo `json:"default_item_logistic_info"`
	}
	rsp := ResponseStructure{}
	err = json.Unmarshal(respUnmarshalled.([]byte), &rsp)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to unmarshal in GetShopDefaultChannel, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}
	if rsp.DefaultItemLogisticInfo == nil {
		return nil, cerr.New("DefaultItemLogisticInfo is nil in GetShopDefaultChannel",
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
	}

	for channelId, v := range rsp.DefaultItemLogisticInfo {
		val, err := strconv.ParseUint(channelId, 10, 64)
		if err != nil {
			val = 0
		}
		v.Channelid = proto.Uint32(uint32(val))
	}
	return rsp.DefaultItemLogisticInfo, nil
}

func (dm *LogisticServiceDm) CalcHiddenFeeForCbSip(ctx context.Context, shopID uint64, region string, itemId uint64, leafCategoryID uint64, weight uint64, enabledChannelIDList []uint32) (float64, error) {
	// checked with business side,
	// for cb sip, we use 0 as hidden fee if no enabled channel
	if len(enabledChannelIDList) == 0 {
		return 0, nil
	}

	weightKg := calcutil.RoundUintToFloat(weight, constant.WeightPrecision, 3)
	weightInGram := weightKg * 1000
	req := model.BatchCalcHiddenFeeRequest{
		List: []*model.CalcHiddenFee{
			{
				ProductIDs: convutil.Uint32sToInt64s(enabledChannelIDList),
				Region:     region,
				ShopID:     shopID,
				SkuInfos: []*model.SkuInfo{
					{
						ItemID:     itemId,
						CategoryID: leafCategoryID,
						Weight:     weightInGram,
						Quantity:   1,
					},
				},
				Direction:    1,
				WmsFlag:      0,
				DgFlag:       0,
				FallbackFlag: 0,
			},
		},
	}
	regionDomainUrl := config.GetShopeeRegionDomainUrl(region)

	ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return 0, cerr.New(fmt.Sprintf("failed to fill CID, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	resp, err := http.BatchCalcHiddenFee(dm.httpCli, ctxWithCID, regionDomainUrl, &req)
	if err != nil {
		errMsg := fmt.Sprintf("failed to calc hidden fee, shopId=%d, itemId=%d, err=%s",
			shopID, itemId, err)
		logging.GetLogger(ctx).Error(errMsg)
		return 0, err
	}

	if resp.RetCode != 0 {
		logging.GetLogger(ctx).Error(fmt.Sprintf("calc hidden fee fail, code=%d, msg=%v", resp.RetCode, resp.Message))
		return 0, cerr.New(fmt.Sprintf("cal hidden fee fail, code=%d, msg=%s", resp.RetCode, resp.Message), uint32(priceSyncPriceCalculationPb.Constant_ERROR_EXTERNAL))
	}

	var maxHiddenFee float64
	for _, channelList := range resp.Data {
		for _, result := range channelList {
			if result.RetCode != 0 {
				errMsg := fmt.Sprintf("failed to calculate hidden fee, channelId=%d, code=%d, msg=%s",
					result.ProductID, result.RetCode, result.Message)
				logging.GetLogger(ctx).Error(errMsg)
				return 0, cerr.New(errMsg, uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
			}

			if result.HiddenFee > maxHiddenFee {
				maxHiddenFee = result.HiddenFee
			}
		}
	}

	return maxHiddenFee, nil
}

func (dm *LogisticServiceDm) CalcHiddenFeeForLocalSip(
	ctx context.Context, pShopId uint64, pRegion string, pItemId uint64, pItemLeafCatId uint64, pShopPickUpLocationIds []uint64,
	aItemQueries []*model.SlsHiddenPriceQuery) ([][]*model.CalcHiddenFeeResult, error) {

	slsQueries := make([]*model.CalcHiddenFee, 0)
	for _, query := range aItemQueries {
		slsQueries = append(slsQueries, &model.CalcHiddenFee{
			ProductIDs: convutil.Uint32sToInt64s(query.PItemEnabledChannelIds),
			Region:     pRegion,
			ShopID:     pShopId,
			ShopGroup:  nil,
			SkuInfos: []*model.SkuInfo{
				{
					ItemID:     pItemId,
					CategoryID: pItemLeafCatId,
					Weight:     query.WeightInGram,
					Quantity:   1,
				},
			},
			Direction:    1,
			WmsFlag:      0,
			DgFlag:       0,
			FallbackFlag: 0,
			PickupLocation: &model.Location{
				LocationIds: pShopPickUpLocationIds,
			},
			DeliverLocation: &model.Location{
				LocationIds: query.DeliveryLocationIds,
			},
		})
	}

	slsResList, err := dm.batchCalcHiddenFeeWithBatchSize(ctx, pRegion, pItemId, slsQueries)

	return slsResList, err
}

func (dm *LogisticServiceDm) batchCalcHiddenFeeWithBatchSize(ctx context.Context, pRegion string, pItemId uint64, slsQueries []*model.CalcHiddenFee) ([][]*model.CalcHiddenFeeResult, error) {
	slsResList := make([][]*model.CalcHiddenFeeResult, 0)
	batchSize := config.GetBatchConfig().MaxBatchSizeForSlsBatchCalcHiddenFee

	for start := 0; start < len(slsQueries); {
		end := start + int(batchSize)
		if end > len(slsQueries) {
			end = len(slsQueries)
		}

		slsRes, iErr := dm.batchCalcHiddenFee(ctx, pRegion, pItemId, slsQueries[start:end])
		if iErr != nil {
			return nil, iErr
		}

		slsResList = append(slsResList, slsRes...)

		start = end
	}

	return slsResList, nil
}

func (dm *LogisticServiceDm) batchCalcHiddenFee(ctx context.Context, pRegion string, pItemId uint64, slsQueries []*model.CalcHiddenFee) ([][]*model.CalcHiddenFeeResult, error) {
	req := model.BatchCalcHiddenFeeRequest{
		List: slsQueries,
	}

	regionDomainUrl := config.GetShopeeRegionDomainUrl(pRegion)

	ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, pRegion)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to fill CID, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	resp, err := http.BatchCalcHiddenFee(dm.httpCli, ctxWithCID, regionDomainUrl, &req)
	if err != nil {
		errMsg := fmt.Sprintf("failed to calc hidden fee, itemId=%d, err=%s", pItemId, err)
		logging.GetLogger(ctx).Error(errMsg)
		return nil, err
	}

	if len(resp.Data) != len(slsQueries) {
		return nil, cerr.New(fmt.Sprintf("the length of response is different with slsQueries, query=%d, response=%d", len(slsQueries),
			len(resp.Data)), uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
	}

	return resp.Data, nil
}

func (dm *LogisticServiceDm) CalcShippingFeeForLocalSIP(ctx context.Context,
	shopID uint64, region string, itemId uint64, leafCategoryID uint64, weightInGram float64, enabledChannelIDList []uint32) (float64, error) {
	req := model.BatchCalcHiddenFeeRequest{
		List: []*model.CalcHiddenFee{
			{
				ProductIDs: convutil.Uint32sToInt64s(enabledChannelIDList),
				Region:     region,
				ShopID:     shopID,
				SkuInfos: []*model.SkuInfo{
					{
						ItemID:     itemId,
						CategoryID: leafCategoryID,
						Weight:     weightInGram,
						Quantity:   1,
					},
				},
				Direction:    1,
				WmsFlag:      0,
				DgFlag:       0,
				FallbackFlag: 0,
			},
		},
	}
	regionDomainUrl := config.GetShopeeRegionDomainUrl(region)

	ctxWithCID, err := cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return 0, cerr.New(fmt.Sprintf("failed to fill CID, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	resp, err := http.BatchCalcHiddenFee(dm.httpCli, ctxWithCID, regionDomainUrl, &req)
	if err != nil {
		errMsg := fmt.Sprintf("failed to calc hidden fee, shopId=%d, itemId=%d, err=%s",
			shopID, itemId, err)
		logging.GetLogger(ctx).Error(errMsg)
		return 0, err
	}

	if resp.RetCode != 0 {
		errMsg := fmt.Sprintf("failed to calc hidden fee, shopId=%d, itemId=%d, ret_code=%d, detail=%s",
			shopID, itemId, resp.RetCode, resp.Detail)
		logging.GetLogger(ctx).Error(errMsg)
		return 0, cerr.New(fmt.Sprintf("failed to call hidden fee, ret_code=%d, detail=%s",
			resp.RetCode, resp.Detail), uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
	}

	var maxHiddenFee float64
	for _, channelList := range resp.Data {
		for _, result := range channelList {
			if result.RetCode != 0 {
				errMsg := fmt.Sprintf("failed to calculate shipping fee, channelId=%d, code=%d, msg=%s",
					result.ProductID, result.RetCode, result.Message)
				return 0, cerr.New(errMsg, uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
			}

			if result.HiddenFee > maxHiddenFee {
				maxHiddenFee = result.HiddenFee
			}
		}
	}

	return maxHiddenFee, nil
}

func (dm *LogisticServiceDm) GetDummyBuyerUserID(ctx context.Context, region string) (int64, error) {
	cacheKey := dm.logisticCacheManager.DummyBuyerUserIdKey(region)
	dummyBuyerUserID, err := dm.logisticCacheManager.GetDummyBuyerUserId(ctx, cacheKey)
	if err == nil {
		return dummyBuyerUserID, nil
	}

	ctx, err = cidutil.FillCtxWithNewCID(ctx, region)
	if err != nil {
		return 0, cerr.New(fmt.Sprintf("failed to fill CID, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_INTERNAL))
	}

	req := &marketplace_order_processing_cb_collection_api.GetDummyBuyerIdRequest{}
	resp, err := dm.orderProcessing.GetDummyBuyerId(ctx, req, region)
	if err != nil {
		return 0, err
	}

	dummyBuyerUserID = resp.GetUserId()
	_ = dm.logisticCacheManager.SetDummyBuyerUserId(ctx, cacheKey, dummyBuyerUserID)

	return dummyBuyerUserID, nil
}

func (dm *LogisticServiceDm) GetChannelInfoMapForRegions(ctx context.Context, regionList []string) (map[string]map[uint64]*model.ChannelInfo, error) {
	if len(regionList) == 0 {
		return nil, nil
	}

	// remove duplicate before querying
	regions := make([]string, 0, len(regionList))
	regionMap := make(map[string]bool)
	for _, region := range regionList {
		regionMap[region] = true
	}
	for region := range regionMap {
		regions = append(regions, region)
	}

	// region -> (channel id -> channel info)
	regionChannelMap := make(map[string]map[uint64]*model.ChannelInfo)

	// fetch from cache first
	missRegionList := make([]string, 0)
	for _, region := range regions {
		cacheKey := constant.GetRegionChannelInfoMapCacheKey(region)
		currRegionChannelInfoMap, err := dm.commonCacheManager.GetRegionLevelChannelInfoMap(ctx, cacheKey)
		if err != nil {
			missRegionList = append(missRegionList, region)
			continue
		}
		regionChannelMap[region] = currRegionChannelInfoMap
	}

	// fetch from api for missed regions
	for _, region := range missRegionList {
		regionChannels, err := dm.logisticsShopChannels.GetChannelsRequest(ctx, region)
		if err != nil { // if has any error, return error directly
			return nil, err
		}

		regionChannelMap[region] = make(map[uint64]*model.ChannelInfo)
		for _, channel := range regionChannels {
			regionChannelMap[region][channel.GetChannelId()] = &model.ChannelInfo{
				ChannelId: channel.GetChannelId(),
				Tag:       channel.GetTag(),
			}
		}

		// set back to cache
		cacheKey := constant.GetRegionChannelInfoMapCacheKey(region)
		_ = dm.commonCacheManager.Set(ctx, cacheKey, regionChannelMap[region], time.Duration(config.GetCommonConfig().RegionChannelInfoCacheExpireSeconds)*time.Second)
	}

	return regionChannelMap, nil
}
