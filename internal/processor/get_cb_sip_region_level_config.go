package processor

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

func (s *CalculationServiceImpl) GetCbSipRegionLevelConfig(ctx context.Context, request *priceSyncPriceCalculationPb.GetCbSipRegionLevelConfigRequest, response *priceSyncPriceCalculationPb.GetCbSipRegionLevelConfigResponse) uint32 {
	p := &getCbSipRegionLevelConfigProcessor{
		ctx:        ctx,
		request:    request,
		response:   response,
		cbsipLogic: s.cbsipLogic,
	}

	err := p.process()
	if err != nil {
		response.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type getCbSipRegionLevelConfigProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetCbSipRegionLevelConfigRequest
	response *priceSyncPriceCalculationPb.GetCbSipRegionLevelConfigResponse

	cbsipLogic logic.CbSipLogic
}

func (g *getCbSipRegionLevelConfigProcessor) process() error {
	if err := g.validateRequest(); err != nil {
		return err
	}

	queries := make([]model.RegionPair, 0)
	for _, pair := range g.request.RegionPairList {
		queries = append(queries, model.RegionPair{
			SrcRegion: pair.GetSrcRegion(),
			DstRegion: pair.GetDstRegion(),
		})
	}
	config, err := g.cbsipLogic.GetCbSipRegionLevelConfig(g.ctx, model.CbSipGetRegionLevelConfigRequest{
		InfoType: g.request.GetInfoType(),
		Queries:  queries,
	})
	if err != nil {
		return err
	}

	exchangeRateConfig := make([]*priceSyncPriceCalculationPb.ExchangeRateData, 0)
	for _, rate := range config.ExchangeRateList {
		exchangeRateConfig = append(exchangeRateConfig, &priceSyncPriceCalculationPb.ExchangeRateData{
			SrcCurrency:  proto.String(rate.SrcCurrency),
			DstCurrency:  proto.String(rate.DstCurrency),
			ExchangeRate: proto.String(rate.ExchangeRate),
		})
	}
	g.response.ExchangeRateConfig = &priceSyncPriceCalculationPb.CbSipRegionLevelExchangeRateConfig{
		ExchangeRateList: exchangeRateConfig,
	}

	countryMarginConfig := make([]*priceSyncPriceCalculationPb.CountryMarginData, 0)
	for _, margin := range config.CountryMarginList {
		countryMarginConfig = append(countryMarginConfig, &priceSyncPriceCalculationPb.CountryMarginData{
			SrcRegion:     proto.String(margin.SrcRegion),
			DstRegion:     proto.String(margin.DstRegion),
			CountryMargin: proto.Float64(margin.CountryMargin),
		})
	}
	g.response.CountryMarginConfig = &priceSyncPriceCalculationPb.CbSipRegionLevelCountryMarginConfig{
		CountryMarginList: countryMarginConfig,
	}
	return nil
}

func (g *getCbSipRegionLevelConfigProcessor) validateRequest() error {
	req := g.request

	if req.InfoType == nil {
		return cerr.New("infoType is nil", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if _, ok := priceSyncPriceCalculationPb.Constant_CbSipRegionLevelInfoType_name[int32(req.GetInfoType())]; !ok {
		return cerr.New("invalid infoType", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	for _, pair := range req.GetRegionPairList() {
		if len(pair.GetSrcRegion()) == 0 || len(pair.GetDstRegion()) == 0 {
			return cerr.New("empty region in region pair list", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if !cutil.IsValidCountry(pair.GetSrcRegion()) {
			return cerr.New(fmt.Sprintf("srcRegion %v is invalid", pair.GetSrcRegion()), uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
		if !cutil.IsValidCountry(pair.GetDstRegion()) {
			return cerr.New(fmt.Sprintf("dstRegion %v is invalid", pair.GetDstRegion()), uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
	}

	return nil
}
