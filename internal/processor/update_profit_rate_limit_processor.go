package processor

import (
	"context"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"

	"github.com/golang/protobuf/proto"
)

func (s *CalculationServiceImpl) UpdateProfitRateLimit(
	ctx context.Context,
	req *pb.UpdateProfitRateLimitRequest, resp *pb.UpdateProfitRateLimitResponse,
) uint32 {
	p := &updateProfitRateLimitProcessor{
		ctx:       ctx,
		request:   req,
		response:  resp,
		cbscLogic: s.cbscLogic,
	}
	err := p.process()
	if err != nil {
		resp.DebugMsg = proto.String(err.Error())
		logging.GetLogger(ctx).Error("response error", ulog.Error(err))
		return GetErrorCode(err)
	}
	return uint32(spCommon.Constant_SUCCESS)
}

type updateProfitRateLimitProcessor struct {
	ctx      context.Context
	request  *pb.UpdateProfitRateLimitRequest
	response *pb.UpdateProfitRateLimitResponse

	cbscLogic logic.CbscLogic
}

func (p *updateProfitRateLimitProcessor) process() error {
	if err := p.validateRequest(); err != nil {
		return err
	}
	err := p.cbscLogic.UpdateProfitRateLimit(p.ctx, p.request.GetRegion(), p.request.GetMerchantRegion(), p.request.ProfitRateMin, p.request.ProfitRateMax, p.request.GetOperator())
	if err != nil {
		return err
	}
	return nil
}

func (p *updateProfitRateLimitProcessor) validateRequest() error {
	if len(p.request.GetRegion()) == 0 {
		return cerr.New("region is empty", uint32(pb.Constant_ERROR_PARAMS))
	}
	if len(p.request.GetMerchantRegion()) == 0 {
		return cerr.New("merchant region is empty", uint32(pb.Constant_ERROR_PARAMS))
	}
	if p.request.ProfitRateMin == nil && p.request.ProfitRateMax == nil {
		return cerr.New("profit rate min and profit rate max both nil", uint32(pb.Constant_ERROR_PARAMS))
	}
	if len(p.request.GetOperator()) == 0 {
		return cerr.New("operator is empty", uint32(pb.Constant_ERROR_PARAMS))
	}
	return nil
}
