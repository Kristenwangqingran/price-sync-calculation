package processor

import (
	"context"
	"math"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	spCommon "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

const maxPageSizeForCbSipAHiddenFeeConfig = 100

func (s *CalculationServiceImpl) GetCbSipAHiddenFeeConfig(ctx context.Context, request *priceSyncPriceCalculationPb.GetCbSipAHiddenFeeConfigRequest, response *priceSyncPriceCalculationPb.GetCbSipAHiddenFeeConfigResponse) uint32 {
	p := &getCbSipAHiddenFeeConfigProcessor{
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

type getCbSipAHiddenFeeConfigProcessor struct {
	ctx      context.Context
	request  *priceSyncPriceCalculationPb.GetCbSipAHiddenFeeConfigRequest
	response *priceSyncPriceCalculationPb.GetCbSipAHiddenFeeConfigResponse

	cbsipLogic logic.CbSipLogic
}

func (g *getCbSipAHiddenFeeConfigProcessor) process() error {
	if err := g.validateRequest(); err != nil {
		return err
	}

	result, err := g.cbsipLogic.GetCbSipAHiddenFeeConfig(g.ctx, model.CbSipGetAHiddenPriceConfigRequest{
		InfoType:  g.request.GetInfoType(),
		PageIndex: g.request.GetPageIndex(),
		PageSize:  g.request.GetPageSize(),
		RuleKey:   g.request.GetRuleKey(),
	})

	if err != nil {
		return err
	}

	rules := make([]*priceSyncPriceCalculationPb.AHiddenFeeRuleInfo, 0)
	for _, rule := range result.Rules {
		rows := make([]*priceSyncPriceCalculationPb.AHiddenFeeRuleRow, 0)
		for _, detail := range rule.Details {
			rows = append(rows, &priceSyncPriceCalculationPb.AHiddenFeeRuleRow{
				WeightRange: proto.Int64(detail.WeightRange),
				StartPrice:  proto.Int64(detail.StartPrice),
				StartWeight: proto.Int64(detail.StartWeight),
				RoundSize:   proto.Int64(detail.RoundSize),
				Price:       proto.Int64(detail.Price),
				WeightStep:  proto.Int64(detail.WeightStep),
				Adjustment:  proto.Int64(detail.Adjustment),
				DescInfo:    proto.String(detail.DescInfo),
			})
		}

		rules = append(rules, &priceSyncPriceCalculationPb.AHiddenFeeRuleInfo{
			RuleKey:  proto.String(rule.RuleKey),
			DescInfo: proto.String(rule.DescInfo),
			Details:  rows,
		})
	}
	g.response.Rules = rules
	g.response.Total = proto.Uint32(result.Total)
	g.response.RuleRegionSettingsStr = proto.String(result.RuleRegionSettingsStr)
	return nil
}

func (g *getCbSipAHiddenFeeConfigProcessor) validateRequest() error {
	req := g.request
	if req.InfoType == nil {
		return cerr.New("infoType is nil", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if _, ok := priceSyncPriceCalculationPb.Constant_CbSipAHiddenFeeInfoType_name[int32(req.GetInfoType())]; !ok {
		return cerr.New("invalid infoType", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	if req.GetInfoType() == uint32(priceSyncPriceCalculationPb.Constant_RULES_HPFN_CONFIG_LIST_WITH_PAGINATION) {
		if req.PageIndex == nil || req.PageSize == nil {
			return cerr.New("PageIndex or PageSize is nil", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
	}

	if req.GetInfoType() == uint32(priceSyncPriceCalculationPb.Constant_RULE_HPFN_CONFIG_DETAIL) {
		if req.GetRuleKey() == "" {
			return cerr.New("invalid rule key", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
		}
	}

	if req.GetPageIndex() >= uint32(math.MaxUint16) || req.GetPageSize() >= maxPageSizeForCbSipAHiddenFeeConfig {
		return cerr.New("invalid page index or page size", uint32(priceSyncPriceCalculationPb.Constant_ERROR_PARAMS))
	}

	return nil
}
