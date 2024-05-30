package logic

import (
	"context"

	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
)

type CbscLogic interface {
	CalculatePriceForCbsc(ctx context.Context, merchantId uint64, isMtskuToMpsku bool, queries []model.MtskuMpskuPriceQuery) ([]model.MtskuMpskuPriceCalcResult, error)
	GetCbscPriceFactor(ctx context.Context, query *pb.GetCbscPriceFactorRequest) (*pb.CbscPriceFactor, error)
	SetCbscPriceFactor(ctx context.Context, query model.SetCbscPriceFactorQuery) error
	GetCbscPriceFactorLimit(ctx context.Context, merchantId uint64) (*pb.CbscServiceFeeRateLimit, map[string]*pb.CbscProfitRateLimit, error)
	UpdateProfitRateLimit(ctx context.Context, region, merchantRegion string, minProfitRateLimit, maxProfitRateLimit *float64, operator string) error
	GetProfitRateLimitListOfMerchantRegion(ctx context.Context, merchantRegion string) ([]*pb.ProfitRateLimit, error)
}
