package system_config

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
)

const (
	sipRateKey = "sipRate"
)

type SystemConfigDM interface {
	GetDefaultCBSIPPriceRatio(ctx context.Context, mstRegion, affiRegion string) (float64, error)
	GetDefaultCBSIPPriceRatioBatch(ctx context.Context, mstRegion string, affiRegionList []string) ([]float64, error)
}

type systemConfigDMImpl struct {
	db db.SystemConfigDB
}

func NewSystemConfigDM(db db.SystemConfigDB) SystemConfigDM {
	return &systemConfigDMImpl{
		db: db,
	}
}

func (s *systemConfigDMImpl) getAllRegionDefaultSIPRateConfig(ctx context.Context) (*sip_db.SystemConfigRecord, error) {
	record, err := s.db.GetSystemConfigRecordByType(ctx, sip_db.SystemConfigTypeDefaultSipRate)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s *systemConfigDMImpl) GetDefaultCBSIPPriceRatio(ctx context.Context, mstRegion, affiRegion string) (float64, error) {
	record, err := s.getAllRegionDefaultSIPRateConfig(ctx)
	if err != nil {
		return 0, err
	}
	return s.getDefaultSIPPriceRatioFromConfigData(ctx, record.ConfigData, mstRegion, affiRegion)
}

func (s *systemConfigDMImpl) getDefaultSIPPriceRatioFromConfigData(ctx context.Context, configDataStr string, mstRegion, affiRegion string) (float64, error) {
	configMap := make(map[string]map[string]float64)
	err := json.Unmarshal([]byte(configDataStr), &configMap)
	if err != nil {
		ulog.DefaultLoggerFromContext(ctx).Error("unmarshal default sip rate system config fail", ulog.String("config data str", configDataStr), ulog.Error(err))
		return 0, cerr.Wrap(err, "unmarshal default sip rate system config fail", uint32(pb.Constant_ERROR_INTERNAL))
	}

	if len(affiRegion) > 0 {
		key := fmt.Sprintf("%s-%s", strings.ToUpper(mstRegion), strings.ToUpper(affiRegion))
		if valueMap, ok := configMap[key]; ok {
			return valueMap[sipRateKey], nil
		}
	}
	key := fmt.Sprintf("%s-ALL", strings.ToUpper(mstRegion))
	if valueMap, ok := configMap[key]; !ok {
		return valueMap[sipRateKey], nil
	}

	//TODO is error_not_found ok here?
	return 0, cerr.Wrap(fmt.Errorf("PriceRatio/defaultSipRate is not config, mstRegion=%s", mstRegion), "", uint32(pb.Constant_ERROR_NOT_FOUND))
}

func (s *systemConfigDMImpl) GetDefaultCBSIPPriceRatioBatch(ctx context.Context, mstRegion string, affiRegionList []string) ([]float64, error) {
	record, err := s.getAllRegionDefaultSIPRateConfig(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]float64, 0, len(affiRegionList))
	for _, affiRegion := range affiRegionList {
		config, err := s.getDefaultSIPPriceRatioFromConfigData(ctx, record.ConfigData, mstRegion, affiRegion)
		if err != nil {
			return nil, err
		}
		res = append(res, config)
	}
	return res, nil
}
