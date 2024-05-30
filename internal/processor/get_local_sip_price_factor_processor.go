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

func (s *CalculationServiceImpl) GetLocalSipPriceFactor(ctx context.Context, request *priceSyncPriceCalculationPb.GetLocalSipPriceFactorRequest, response *priceSyncPriceCalculationPb.GetLocalSipPriceFactorResponse) uint32 {
	p := &getLocalSipPriceFactorProcessor{
		ctx:           ctx,
		request:       request,
		response:      response,
		localSipLogic: s.localSipLogic,
	}

	err := p.process()
	if err != nil {
		response.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type getLocalSipPriceFactorProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetLocalSipPriceFactorRequest
	response *priceSyncPriceCalculationPb.GetLocalSipPriceFactorResponse

	localSipLogic logic.LocalSipLogic
}

func (g *getLocalSipPriceFactorProcessor) process() error {
	if err := g.validateRequest(); err != nil {
		return err
	}

	queries := make([]model.GetLocalSipPriceFactorQuery, 0)
	for i, pair := range g.request.GetRegionPairList() {
		queries = append(queries, model.GetLocalSipPriceFactorQuery{
			QueryId: i,
			PRegion: pair.GetSrcRegion(),
			ARegion: pair.GetDstRegion(),
		})
	}
	factors, err := g.localSipLogic.GetLocalSipPriceFactors(g.ctx, model.LocalSipPriceFactorInfoType(g.request.GetInfoType()), queries)

	if err != nil {
		return err
	}

	respResults := make([]*priceSyncPriceCalculationPb.LocalSipPriceFactorInfo, 0)
	for _, factor := range factors {
		hiddenFeeInfos := make([]*priceSyncPriceCalculationPb.LocalShippingFeeRule, 0)
		shippingFeeInfos := make([]*priceSyncPriceCalculationPb.LocalShippingFeeRule, 0)
		for _, info := range factor.LocalHiddenFeeInfo {
			hiddenFeeInfos = append(hiddenFeeInfos, &priceSyncPriceCalculationPb.LocalShippingFeeRule{
				MstRegion:  proto.String(info.PRegion),
				AffiRegion: proto.String(info.ARegion),
				Weight:     proto.Int64(info.Weight),
				Fee:        proto.Int64(info.HiddenPrice),
				Id:         proto.Int64(info.Id),
				Ctime:      proto.Int64(info.Ctime),
			})
		}
		for _, info := range factor.LocalShippingFeeInfo {
			shippingFeeInfos = append(shippingFeeInfos, &priceSyncPriceCalculationPb.LocalShippingFeeRule{
				Id:         proto.Int64(info.Id),
				Ctime:      proto.Int64(info.Ctime),
				MstRegion:  proto.String(info.PRegion),
				AffiRegion: proto.String(info.ARegion),
				Weight:     proto.Int64(info.Weight),
				Fee:        proto.Int64(info.ShippingFeePrice),
			})
		}

		var basicInfo *priceSyncPriceCalculationPb.LocalSipPriceFactorBasicInfo
		if factor.BasicInfo != nil {
			var shippingFeeToggle, hiddenFeeToggle *int32
			if factor.BasicInfo.ShippingFeeToggle != nil {
				if _, ok := priceSyncPriceCalculationPb.Constant_PriceSyncToggle_name[*factor.BasicInfo.ShippingFeeToggle]; ok {
					shippingFeeToggle = factor.BasicInfo.ShippingFeeToggle
				}
			}
			if factor.BasicInfo.InitialHiddenFeeToggle != nil {
				if _, ok := priceSyncPriceCalculationPb.Constant_PriceSyncToggle_name[*factor.BasicInfo.InitialHiddenFeeToggle]; ok {
					hiddenFeeToggle = factor.BasicInfo.InitialHiddenFeeToggle
				}
			}

			basicInfo = &priceSyncPriceCalculationPb.LocalSipPriceFactorBasicInfo{
				MinCountryMargin:       proto.Float64(factor.BasicInfo.MinCountryMargin),
				MaxCountryMargin:       proto.Float64(factor.BasicInfo.MaxCountryMargin),
				MinExchangeRate:        proto.Float64(factor.BasicInfo.MinExchangeRate),
				MaxExchangeRate:        proto.Float64(factor.BasicInfo.MaxExchangeRate),
				MinInitHiddenPrice:     proto.Float64(factor.BasicInfo.MinInitHiddenPrice),
				MaxInitHiddenPrice:     proto.Float64(factor.BasicInfo.MaxInitHiddenPrice),
				CountryMargin:          factor.BasicInfo.CountryMargin,
				ExchangeRate:           factor.BasicInfo.ExchangeRate,
				InitHiddenPrice:        factor.BasicInfo.InitHiddenPrice,
				ShippingFeeToggle:      shippingFeeToggle,
				InitialHiddenFeeToggle: hiddenFeeToggle,
			}
		}

		// fill based on info type in the request
		respRes := &priceSyncPriceCalculationPb.LocalSipPriceFactorInfo{}
		switch g.request.GetInfoType() {
		case uint32(priceSyncPriceCalculationPb.Constant_BASIC_INFO):
			respRes.BasicInfo = basicInfo
		case uint32(priceSyncPriceCalculationPb.Constant_HIDDEN_FEE):
			respRes.HiddenFeeInfo = &priceSyncPriceCalculationPb.LocalSipPriceFactorHiddenFeeInfo{
				LocalHiddenFeeRules: hiddenFeeInfos,
			}
		case uint32(priceSyncPriceCalculationPb.Constant_SHIPPING_FEE):
			respRes.ShippingFeeInfo = &priceSyncPriceCalculationPb.LocalSipPriceFactorShippingFeeInfo{
				LocalShippingFeeRules: shippingFeeInfos,
			}
		}
		respResults = append(respResults, respRes)
	}
	g.response.Results = respResults

	return nil
}

func (g *getLocalSipPriceFactorProcessor) validateRequest() error {
	req := g.request

	if req.InfoType == nil {
		return cerr.New("infoType is nil", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if _, ok := priceSyncPriceCalculationPb.Constant_LocalSipInfoType_name[int32(req.GetInfoType())]; !ok {
		return cerr.New("invalid infoType", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if len(req.GetRegionPairList()) == 0 {
		return cerr.New("empty region pair list", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	for _, pair := range req.GetRegionPairList() {
		if len(pair.GetSrcRegion()) == 0 || len(pair.GetDstRegion()) == 0 {
			return cerr.New("empty region in region pair list", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
		if pair.GetSrcRegion() == pair.GetDstRegion() {
			return cerr.New("srcRegion and dstRegion should be different", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
		if !cutil.IsValidCountry(pair.GetSrcRegion()) {
			return cerr.New(fmt.Sprintf("srcRegion %v is not a invalid", pair.GetSrcRegion()), uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
		if !cutil.IsValidCountry(pair.GetDstRegion()) {
			return cerr.New(fmt.Sprintf("dstRegion %v is not a invalid", pair.GetDstRegion()), uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
	}

	return nil
}
