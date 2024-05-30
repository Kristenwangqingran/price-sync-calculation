package cb_sip_logic

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

func (c *CbSipLogicImpl) GetCbSipRegionLevelConfig(ctx context.Context, req model.CbSipGetRegionLevelConfigRequest) (model.CbSipGetRegionLevelConfigResult, error) {
	if req.InfoType == uint32(pb.Constant_EXCHANGE_RATE) {
		exchangeRates, err := c.getExchangeRates(ctx, req.Queries)
		if err != nil {
			return model.CbSipGetRegionLevelConfigResult{}, err
		}
		return model.CbSipGetRegionLevelConfigResult{
			ExchangeRateList: exchangeRates,
		}, nil
	} else if req.InfoType == uint32(pb.Constant_COUNTRY_MARGIN) {
		countryMargins, err := c.getCountryMargins(ctx, req.Queries)
		if err != nil {
			return model.CbSipGetRegionLevelConfigResult{}, err
		}
		return model.CbSipGetRegionLevelConfigResult{
			CountryMarginList: countryMargins,
		}, nil
	}
	return model.CbSipGetRegionLevelConfigResult{}, cerr.New(fmt.Sprintf("unknown infoType: %v", req.InfoType), uint32(pb.Constant_ERROR_PARAMS))
}

func (c *CbSipLogicImpl) getExchangeRates(ctx context.Context, queries []model.RegionPair) ([]model.ExchangeRate, error) {
	exchangeRateMap, err := c.factorsRepo.GetAllExchangeRateMapForCbSip(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]model.ExchangeRate, 0)

	// get all
	if len(queries) == 0 {
		for srcCurrency, exchangeRates := range exchangeRateMap {
			for dstCurrency, exchangeRate := range exchangeRates {
				res = append(res, model.ExchangeRate{
					SrcCurrency:  srcCurrency,
					DstCurrency:  dstCurrency,
					ExchangeRate: exchangeRate,
				})
			}
		}
		return res, nil
	}

	for _, query := range queries {
		srcCurrency, err := config.GetCurrencyByRegion(query.SrcRegion)
		if err != nil {
			return nil, err
		}
		dstCurrency, err := config.GetCurrencyByRegion(query.DstRegion)
		if err != nil {
			return nil, err
		}
		if exchangeRateMap[srcCurrency] == nil {
			return nil, cerr.New(fmt.Sprintf("failed to get exchange rate for currency=%v", srcCurrency), uint32(pb.Constant_ERROR_NOT_FOUND))
		}
		res = append(res, model.ExchangeRate{
			SrcCurrency:  srcCurrency,
			DstCurrency:  dstCurrency,
			ExchangeRate: exchangeRateMap[srcCurrency][dstCurrency],
		})
	}
	return res, nil
}

func (c *CbSipLogicImpl) getCountryMargins(ctx context.Context, queries []model.RegionPair) ([]model.CountryMargin, error) {
	countryMarginMap, err := c.factorsRepo.GetAllCountryMarginMapForCbSip(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]model.CountryMargin, 0)

	// return all
	if len(queries) == 0 {
		for srcRegion, margins := range countryMarginMap {
			for dstRegion, m := range margins {
				res = append(res, model.CountryMargin{
					SrcRegion:     srcRegion,
					DstRegion:     dstRegion,
					CountryMargin: m,
				})
			}
		}
		return res, nil
	}

	for _, query := range queries {
		if countryMarginMap[query.SrcRegion] == nil {
			return nil, cerr.New(fmt.Sprintf("failed to get exchange rate for currency=%v", query.SrcRegion), uint32(pb.Constant_ERROR_NOT_FOUND))
		}
		res = append(res, model.CountryMargin{
			SrcRegion:     query.SrcRegion,
			DstRegion:     query.DstRegion,
			CountryMargin: countryMarginMap[query.SrcRegion][query.DstRegion],
		})
	}
	return res, nil
}
