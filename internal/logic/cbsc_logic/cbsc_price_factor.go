package cbsc_logic

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/convutil"
)

func (c *CbscLogicImpl) GetCbscPriceFactor(ctx context.Context, query *pb.GetCbscPriceFactorRequest) (*pb.CbscPriceFactor, error) {
	switch query.GetInfoType() {
	case uint32(pb.Constant_CBSC_SHOP_LEVEL_FEE_RATE):
		shopFeeRateList, err := c.factorsRepo.GetCbscShopLevelFeeRate(ctx, query.GetMerchantId(), query.MainAccountId, query.GetShopIdList())
		if err != nil {
			return nil, err
		}
		return &pb.CbscPriceFactor{
			ShopFeeRateList: shopFeeRateList,
		}, nil
	case uint32(pb.Constant_CBSC_FEE_RATE_LIMIT):
		feeRateLimit, err := c.factorsRepo.GetCbscFeeRateLimit(ctx, query.GetMerchantId(), query.ShopRegion)
		if err != nil {
			return nil, err
		}
		return &pb.CbscPriceFactor{
			FeeRateLimit: feeRateLimit,
		}, nil
	case uint32(pb.Constant_CBSC_EXCHANGE_RATE):
		_, exchangeRateList, err := c.factorsRepo.GetCbscExchangeRate(ctx, query.GetMerchantId())
		if err != nil {
			return nil, err
		}
		return &pb.CbscPriceFactor{
			ExchangeRateList: exchangeRateList,
		}, nil
	default:
		return nil, cerr.New(fmt.Sprintf("unsupported input type =%d", query.InfoType), uint32(pb.Constant_ERROR_PARAMS))
	}
}

func (c *CbscLogicImpl) SetCbscPriceFactor(ctx context.Context, query model.SetCbscPriceFactorQuery) error {
	return c.factorsRepo.SetCbscPriceFactors(ctx, query)
}

func (c *CbscLogicImpl) GetCbscPriceFactorLimit(ctx context.Context, merchantId uint64) (*pb.CbscServiceFeeRateLimit, map[string]*pb.CbscProfitRateLimit, error) {
	limit, err := c.factorsRepo.GetCbscFeeRateLimit(ctx, merchantId, nil)
	if err != nil {
		return nil, nil, err
	}

	if limit == nil {
		return nil, nil, nil
	}

	serviceFeeLimit := limit.GetServiceFeeLimit()

	profitRateLimitMap := make(map[string]*pb.CbscProfitRateLimit, 0) // based on shop region
	for _, p := range limit.GetProfitRateLimit() {
		profitRateLimitMap[p.GetRegion()] = p
	}

	return serviceFeeLimit, profitRateLimitMap, nil
}

func (c *CbscLogicImpl) UpdateProfitRateLimit(ctx context.Context, region, merchantRegion string, minProfitRateLimit, maxProfitRateLimit *float64, operator string) error {
	err := c.factorsRepo.UpdateProfitRateLimit(ctx, region, merchantRegion, minProfitRateLimit, maxProfitRateLimit, operator)
	if err != nil {
		return err
	}
	return nil
}

func (c *CbscLogicImpl) GetProfitRateLimitListOfMerchantRegion(ctx context.Context, merchantRegion string) ([]*pb.ProfitRateLimit, error) {
	internalProfitRateLimitList, err := c.factorsRepo.GetProfitRateLimit(ctx, "", merchantRegion)
	if err != nil {
		return nil, err
	}
	return convutil.ConvertMerchantConstraintListToProfitRateLimitList(internalProfitRateLimitList), nil
}
