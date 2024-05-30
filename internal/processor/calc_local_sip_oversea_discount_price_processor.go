package processor

import (
	"context"
	"fmt"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/calculate"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/data"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"

	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

const (
	batchSizeLimit = 20

	maxDiscountRate = 100_000 // TODO combine with DbPectInflationFactor if discount side finalized the TD
)

func (s *CalculationServiceImpl) CalcLocalSipOverseaDiscountPrice(ctx context.Context,
	req *priceSyncPriceCalculationPb.CalcLocalSipOverseaDiscountPriceRequest,
	resp *priceSyncPriceCalculationPb.CalcLocalSipOverseaDiscountPriceResponse) uint32 {

	p := &calcLocalSipOverseaDiscountPriceProcessor{
		ctx:  ctx,
		req:  req,
		resp: resp,
	}
	p.fetchCalcFactorDataDm = s.FetchCalcFactorForAffiMpskuDm
	p.calculateDm = s.CalculateAffiMpskuDm

	err := p.process()
	if err != nil {
		resp.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}

	return uint32(spCommon.Constant_SUCCESS)
}

type calcLocalSipOverseaDiscountPriceProcessor struct {
	ctx  context.Context
	req  *priceSyncPriceCalculationPb.CalcLocalSipOverseaDiscountPriceRequest
	resp *priceSyncPriceCalculationPb.CalcLocalSipOverseaDiscountPriceResponse

	fetchCalcFactorDataDm data.FetchCalcFactorForAffiMpskuDm
	calculateDm           calculate.CalculateAffiMpskuDm
}

func (p *calcLocalSipOverseaDiscountPriceProcessor) process() error {
	if err := p.validateRequest(); err != nil {
		return err
	}

	calcFactorData, err := p.fetchCalcFactorDataDm.FetchCalcFactorDataForLocalSipOverseaDiscount(p.ctx, p.req)
	if err != nil {
		logging.GetLogger(p.ctx).Error("FetchCalcFactorDataForLocalSipOverseaDiscount failed",
			ulog.Reflect("req", p.req), ulog.Error(err))
		return err
	}

	result, err := p.calculateDm.CalcLocalSipOverseaDiscount(p.ctx, p.req, calcFactorData)
	if err != nil {
		logging.GetLogger(p.ctx).Error("CalcLocalSipOverseaDiscount failed",
			ulog.Reflect("req", p.req), ulog.Error(err))
		return err
	}
	p.resp.Results = result

	return nil
}

func (p *calcLocalSipOverseaDiscountPriceProcessor) validateRequest() error {
	if len(p.req.GetAffiItemModelIds()) == 0 || len(p.req.GetAffiItemModelIds()) > batchSizeLimit {
		return cerr.New(fmt.Sprintf("invalid req|len(req.AffiItemModelIds) = %d", len(p.req.GetAffiItemModelIds())),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if p.req.GetDiscountRate() > maxDiscountRate || p.req.GetDiscountRate() <= 0 {
		return cerr.New(fmt.Sprintf("invalid req|discountRate = %d", p.req.GetDiscountRate()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	return nil
}
