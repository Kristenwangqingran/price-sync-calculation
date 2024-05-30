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

func (s *CalculationServiceImpl) GetProfitRateLimitList(
	ctx context.Context,
	req *pb.GetProfitRateLimitListRequest, resp *pb.GetProfitRateLimitListResponse,
) uint32 {
	p := &getProfitRateLimitListProcessor{
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

type getProfitRateLimitListProcessor struct {
	ctx      context.Context
	request  *pb.GetProfitRateLimitListRequest
	response *pb.GetProfitRateLimitListResponse

	cbscLogic logic.CbscLogic
}

func (p *getProfitRateLimitListProcessor) process() error {
	if err := p.validateRequest(); err != nil {
		return err
	}
	profitRateList, err := p.cbscLogic.GetProfitRateLimitListOfMerchantRegion(p.ctx, p.request.GetMerchantRegion())
	if err != nil {
		return err
	}
	p.response.Data = profitRateList
	return nil
}

func (p *getProfitRateLimitListProcessor) validateRequest() error {
	if len(p.request.GetMerchantRegion()) == 0 {
		return cerr.New("merchant region is empty", uint32(pb.Constant_ERROR_PARAMS))
	}
	return nil
}
