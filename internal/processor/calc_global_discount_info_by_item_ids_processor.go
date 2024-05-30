package processor

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/calculate"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/data"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

func (s *CalculationServiceImpl) CalcGlobalDiscountInfoByItemIds(ctx context.Context, req *priceSyncPriceCalculationPb.CalcGlobalDiscountInfoByItemIdsRequest, resp *priceSyncPriceCalculationPb.CalcGlobalDiscountInfoByItemIdsResponse) uint32 {
	p := &calcGlobalDiscountInfoByItemIdsProcessor{
		ctx:  ctx,
		req:  req,
		resp: resp,
	}
	p.fetchCalcFactorDataDm = s.FetchCalcFactorForMtskuAndMpskuDm
	p.calculateDm = s.CalculateMtskuAndMpskuDm

	err := p.process()
	if err != nil {
		resp.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}

	return uint32(spCommon.Constant_SUCCESS)
}

type calcGlobalDiscountInfoByItemIdsProcessor struct {
	ctx  context.Context
	req  *priceSyncPriceCalculationPb.CalcGlobalDiscountInfoByItemIdsRequest
	resp *priceSyncPriceCalculationPb.CalcGlobalDiscountInfoByItemIdsResponse

	fetchCalcFactorDataDm data.FetchCalcFactorForMtskuAndMpskuDm
	calculateDm           calculate.CalculateMtskuAndMpskuDm
}

func (p *calcGlobalDiscountInfoByItemIdsProcessor) process() error {
	// validate request param
	if err := p.validateRequest(); err != nil {
		return err
	}

	// fetch factor data for calculation
	calcFactorData := p.fetchCalcFactorDataDm.FetchCalcFactorDataForGlobalDiscount(p.ctx, p.req.GetQueries())

	// calculate
	globalDiscountInfoList, err := p.calculateDm.CalcGlobalDiscount(p.ctx, p.req.GetQueries(), calcFactorData)
	if err != nil {
		return err
	}

	p.resp.GlobalDiscountInfoList = globalDiscountInfoList
	return nil
}

func (p *calcGlobalDiscountInfoByItemIdsProcessor) validateRequest() error {
	batchSize := config.GetBatchConfig().MaxBatchSizeForCalcGlobalDiscountInfoByItemIds
	if len(p.req.GetQueries()) == 0 || len(p.req.GetQueries()) > int(batchSize) {
		return cerr.New(fmt.Sprintf("query size should be in (0, %d]", batchSize),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	for _, query := range p.req.GetQueries() {
		if query.GetMerchantId() == 0 {
			return cerr.New("merchantId cannot be empty or 0",
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if query.GetMpskuShopId() == 0 {
			return cerr.New("MpskuShopId cannot be empty or 0",
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if query.GetMpskuItemId() == 0 {
			return cerr.New("MpskuItemId cannot be empty or 0",
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if !cutil.IsValidCountry(query.GetMpskuRegion()) {
			return cerr.New("MpskuRegion should be valid",
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if query.GetMtskuOriginalPrice() == 0 {
			return cerr.New("MtskuOriginalPrice cannot be empty or 0",
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if query.GlobalDiscountInputType == nil {
			return cerr.New("GlobalDiscountInputType cannot be empty",
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if _, exist := priceSyncPriceCalculationPb.Constant_GlobalDiscountInputType_name[int32(query.GetGlobalDiscountInputType())]; !exist {
			return cerr.New("GlobalDiscountInputType is not valid",
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}

		if query.GetGlobalDiscountQueryData() == 0 {
			return cerr.New("GlobalDiscountQueryData cannot be empty or 0",
				uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
	}

	return nil
}
