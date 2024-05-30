package local_sip_logic

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

func (l *LocalSipLogicImpl) GetLocalSipPriceFactors(ctx context.Context, infoType model.LocalSipPriceFactorInfoType, queries []model.GetLocalSipPriceFactorQuery) ([]model.LocalSipPriceFactorInfo, error) {
	switch infoType {
	case model.BasicInfo:
		return l.getLocalSipPriceFactorBasicInfo(ctx, queries)
	case model.HiddenFeeInfo:
		return l.getLocalSipPriceFactorHiddenFeeInfo(ctx, queries)
	case model.ShippingFeeInfo:
		return l.getLocalSipPriceFactorShippingFeeInfo(ctx, queries)
	default:
		return nil, cerr.New(fmt.Sprintf("invalid infoType=%v", infoType), uint32(pb.Constant_ERROR_PARAMS))
	}
}

func (l *LocalSipLogicImpl) getLocalSipPriceFactorShippingFeeInfo(ctx context.Context, queries []model.GetLocalSipPriceFactorQuery) ([]model.LocalSipPriceFactorInfo, error) {
	finalResults := make([]model.LocalSipPriceFactorInfo, len(queries))

	session := l.sipRepo.DbSession()
	for i, query := range queries {
		dbRes, err := l.sipRepo.GetShippingFeeConfigList(ctx, session, query.PRegion, query.ARegion)
		if err != nil {
			return nil, err
		}
		shippingFeeInfos := make([]*model.LocalSipFactorShippingFeeInfo, 0)
		for _, dbInfo := range dbRes {
			shippingFeeInfos = append(shippingFeeInfos, &model.LocalSipFactorShippingFeeInfo{
				Id:               dbInfo.Id,
				Ctime:            dbInfo.Ctime,
				PRegion:          dbInfo.MstRegion,
				ARegion:          dbInfo.AffiRegion,
				Weight:           dbInfo.Weight,
				ShippingFeePrice: dbInfo.ShippingFeePrice,
			})
		}
		finalResults[i] = model.LocalSipPriceFactorInfo{
			LocalShippingFeeInfo: shippingFeeInfos,
		}
	}
	return finalResults, nil
}

func (l *LocalSipLogicImpl) getLocalSipPriceFactorHiddenFeeInfo(ctx context.Context, queries []model.GetLocalSipPriceFactorQuery) ([]model.LocalSipPriceFactorInfo, error) {
	finalResults := make([]model.LocalSipPriceFactorInfo, len(queries))

	session := l.sipRepo.DbSession()
	for i, query := range queries {
		dbRes, err := l.sipRepo.GetHiddenPriceConfigList(ctx, session, query.PRegion, query.ARegion)
		if err != nil {
			return nil, err
		}
		hiddenFeeInfos := make([]*model.LocalSipFactorHiddenFeeInfo, 0)
		for _, dbInfo := range dbRes {
			hiddenFeeInfos = append(hiddenFeeInfos, &model.LocalSipFactorHiddenFeeInfo{
				Id:          dbInfo.Id,
				Ctime:       dbInfo.Ctime,
				PRegion:     dbInfo.MstRegion,
				ARegion:     dbInfo.AffiRegion,
				Weight:      dbInfo.Weight,
				HiddenPrice: dbInfo.HiddenPrice,
			})
		}
		finalResults[i] = model.LocalSipPriceFactorInfo{
			LocalHiddenFeeInfo: hiddenFeeInfos,
		}
	}
	return finalResults, nil
}

func (l *LocalSipLogicImpl) getLocalSipPriceFactorBasicInfo(ctx context.Context, queries []model.GetLocalSipPriceFactorQuery) ([]model.LocalSipPriceFactorInfo, error) {
	allLocalSipPriceConfig, err := l.factors.GetAllLocalSipPriceConfig(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]model.LocalSipPriceFactorInfo, len(queries))
	for i, query := range queries {
		pRegion, aRegion := query.PRegion, query.ARegion

		// fill limit first
		limitCfg, err := config.GetLocalSipSettingLimitConfigByRegions(pRegion, aRegion)
		if err != nil {
			return nil, err
		}
		res[i] = model.LocalSipPriceFactorInfo{
			BasicInfo: &model.LocalSipFactorBasicInfo{
				MinCountryMargin:   limitCfg.BufferMin,
				MaxCountryMargin:   limitCfg.BufferMax,
				MinExchangeRate:    limitCfg.ExchangeRateMin,
				MaxExchangeRate:    limitCfg.ExchangeRateMax,
				MinInitHiddenPrice: limitCfg.InitHiddenPriceMin,
				MaxInitHiddenPrice: limitCfg.InitHiddenPriceMax,
			},
		}

		// fill basic info if exists
		if allLocalSipPriceConfig[pRegion] == nil {
			logging.GetLogger(ctx).Info(fmt.Sprintf("cannot find local sip price config for pRegion=%v", pRegion))
			continue
		}
		localSipCfg := allLocalSipPriceConfig[pRegion][aRegion]
		if localSipCfg == nil {
			logging.GetLogger(ctx).Info(fmt.Sprintf("cannot find localSipConfig for pRegion=%v, aRegion=%v", pRegion, aRegion))
			continue
		}
		res[i].BasicInfo.CountryMargin = localSipCfg.Buffer
		res[i].BasicInfo.ExchangeRate = localSipCfg.ExchangeRate
		res[i].BasicInfo.InitHiddenPrice = localSipCfg.InitHiddenPrice
		res[i].BasicInfo.InitialHiddenFeeToggle = localSipCfg.HiddenPriceToggle
		res[i].BasicInfo.ShippingFeeToggle = localSipCfg.ShippingFeeToggle
	}

	return res, nil
}
