package cb_sip_logic

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

func (c *CbSipLogicImpl) GetCbSipAHiddenFeeConfig(ctx context.Context, req model.CbSipGetAHiddenPriceConfigRequest) (*model.CbSipGetAHiddenPriceConfigResult, error) {
	switch req.InfoType {
	case model.RulesListWithPagination:
		return c.getHpfnRateTables(ctx, req.PageIndex, req.PageSize)
	case model.RulesListAll:
		return c.getHpfnAll(ctx)
	case model.RuleDetail:
		return c.getHpfnRateTableDetail(ctx, req.RuleKey)
	case model.RuleRegionSetting:
		return c.getRegionRateTableSetting(ctx)
	default:
		return nil, cerr.New(fmt.Sprintf("invalid infoType:%v", req.InfoType), uint32(pb.Constant_ERROR_PARAMS))
	}
}

func (c *CbSipLogicImpl) getHpfnAll(ctx context.Context) (*model.CbSipGetAHiddenPriceConfigResult, error) {
	allCfg, err := c.hpfnConfigRepo.GetAllHpfnConfigFromDb(ctx)
	if err != nil {
		return nil, err
	}

	data := map[string][]*sip_db.HpfnConfig{}
	for _, config := range allCfg {
		data[config.HpfnKey] = append(data[config.HpfnKey], config)
	}

	rules := make([]model.AHiddenFeeRuleInfo, 0)
	for key, configs := range data {
		rows := make([]model.AHiddenFeeRuleRow, 0)
		desc := ""
		for _, cfg := range configs {
			if len(desc) == 0 && len(cfg.DescInfo) != 0 {
				desc = cfg.DescInfo
			}
			rows = append(rows, model.AHiddenFeeRuleRow{
				WeightRange: cfg.WeightRange,
				StartPrice:  cfg.StartPrice,
				StartWeight: cfg.StartWeight,
				RoundSize:   cfg.RoundSize,
				Price:       cfg.Price,
				WeightStep:  cfg.WeightStep,
				Adjustment:  cfg.Adjustment,
				DescInfo:    cfg.DescInfo,
			})
		}
		rules = append(rules, model.AHiddenFeeRuleInfo{
			RuleKey:  key,
			DescInfo: desc,
			Details:  rows,
		})
	}
	return &model.CbSipGetAHiddenPriceConfigResult{
		Rules: rules,
	}, nil
}

func (c *CbSipLogicImpl) getHpfnRateTableDetail(ctx context.Context, ruleKey string) (*model.CbSipGetAHiddenPriceConfigResult, error) {
	allCfg, err := c.hpfnConfigRepo.GetAllHpfnConfig(ctx)
	if err != nil {
		return nil, err
	}
	cfgs, ok := allCfg[ruleKey]
	if !ok {
		return nil, cerr.New(fmt.Sprintf("failed to get hpfn rate detail, key=%v", ruleKey), uint32(pb.Constant_ERROR_NOT_FOUND))
	}
	rows := []model.AHiddenFeeRuleRow{}
	desc := ""
	for _, cfg := range cfgs {
		if len(cfg.DescInfo) > 0 && len(desc) == 0 {
			desc = cfg.DescInfo
		}

		x := model.AHiddenFeeRuleRow{
			WeightRange: cfg.WeightRange,
			StartPrice:  cfg.StartPrice,
			StartWeight: cfg.StartWeight,
			RoundSize:   cfg.RoundSize,
			Price:       cfg.Price,
			WeightStep:  cfg.WeightStep,
			Adjustment:  cfg.Adjustment,
			DescInfo:    cfg.DescInfo,
		}
		rows = append(rows, x)
	}
	return &model.CbSipGetAHiddenPriceConfigResult{
		Rules: []model.AHiddenFeeRuleInfo{
			{
				RuleKey:  ruleKey,
				DescInfo: desc,
				Details:  rows,
			},
		},
	}, nil
}

func (c *CbSipLogicImpl) getRegionRateTableSetting(ctx context.Context) (*model.CbSipGetAHiddenPriceConfigResult, error) {
	session := c.sipRepo.DbSession()
	config, err := c.sipRepo.GetSystemConfigRecordByType(ctx, session, sip_db.SystemConfigTypeRegionRateTableConfig)
	if err != nil {
		return nil, err
	}
	if config == nil || len(config.ConfigData) == 0 {
		return nil, cerr.New("failed to get region rate table config", uint32(pb.Constant_ERROR_NOT_FOUND))
	}
	return &model.CbSipGetAHiddenPriceConfigResult{
		RuleRegionSettingsStr: config.ConfigData,
	}, nil
}

func (c *CbSipLogicImpl) getHpfnRateTables(ctx context.Context, pageIndex uint32, pageSize uint32) (*model.CbSipGetAHiddenPriceConfigResult, error) {
	all, err := c.hpfnConfigRepo.GetAllHpfnConfig(ctx)
	if err != nil {
		return nil, err
	}
	briefList := []model.AHiddenFeeRuleInfo{}
	for key, cfgs := range all {
		if len(cfgs) == 0 {
			// bad cfg
			logging.GetLogger(ctx).Error(fmt.Sprintf("hpfn rate key=%s has no configs", key))
			continue
		}
		desc := ""
		for _, cfg := range cfgs {
			if len(cfg.DescInfo) > 0 {
				desc = cfg.DescInfo
				break
			}
		}
		briefList = append(briefList, model.AHiddenFeeRuleInfo{
			RuleKey:  key,
			DescInfo: desc,
		})
	}
	if len(briefList) > 0 {
		sort.Slice(briefList, func(i, j int) bool {
			return strings.Compare(briefList[i].RuleKey, briefList[j].RuleKey) < 0
		})
		briefLen := uint32(len(briefList))
		startIdx := pageIndex * pageSize
		endIdx := (pageIndex + 1) * pageSize
		if startIdx > briefLen {
			startIdx = briefLen
		}
		if endIdx > briefLen {
			endIdx = briefLen
		}
		briefList = briefList[startIdx:endIdx]
	}

	return &model.CbSipGetAHiddenPriceConfigResult{
		Total: uint32(len(all)),
		Rules: briefList,
	}, nil
}
