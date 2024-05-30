package factors

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	internalExchangeRatePb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_exchange_rate.pb"
	internalMerchantConfigSettingPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
	internalMerchantConstraintsPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_constraints.pb"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	ib "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/item_business.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_v2_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
)

type CalculationFactorsRepo interface {
	// CBSC
	GetHidePriceForCbsc(ctx context.Context, queries []model.GetHidePriceForCbscRequest) ([]model.GetHidePriceForCbscResult, error)
	GetExchangeRateMapForCbsc(ctx context.Context, merchantId uint64) (string, map[string]float64, error)
	GetCommissionRateBatchForCbsc(ctx context.Context, queries []model.GetCommissionRateRequest) []model.GetCommissionRateResult
	GetCbscPriceRateBatchForCbsc(ctx context.Context, merchantRegion string, merchantConfigMap map[uint64]*internalMerchantConfigSettingPb.MerchantConfigSetting, queries []model.GetCbscPriceRateRequest) ([]model.GetCbscPriceRateResult, error)
	GetProfitRateBatchForCbsc(ctx context.Context, merchantConfigMap map[uint64]*internalMerchantConfigSettingPb.MerchantConfigSetting, queries []model.GetCbscProfitRateRequest) ([]model.GetCbscProfitRateResult, error)
	GetProfitRateLimit(ctx context.Context, region string, merchantRegion string) ([]*internalMerchantConstraintsPb.MerchantConstraints, error)
	UpdateProfitRateLimit(ctx context.Context, region, merchantRegion string, profitRateMin, profitRateMax *float64, operator string) error
	GetReferenceServiceFeeRateListForCbsc(ctx context.Context, shopIdList []model.ShopIdRegion) map[uint64]uint64
	GetCbscShopPriceCommonConfig(ctx context.Context, merchantRegion string) (*config.CBSCPriceFeeConfig, error)
	GetCbscExchangeRate(ctx context.Context, merchantId uint64) (string, []*pb.CbscExchangeRate, error)
	GetCbscFeeRateLimit(ctx context.Context, merchantId uint64, shopRegion *string) (*pb.CbscFeeRateLimit, error)
	GetCbscShopLevelFeeRate(ctx context.Context, merchantId uint64, mainAccountId *uint64, shopIds []uint64) ([]*pb.CbscShopLevelFeeRate, error)
	SetCbscPriceFactors(ctx context.Context, query model.SetCbscPriceFactorQuery) error

	// LocalSIP
	GetAllLocalSipPriceConfig(ctx context.Context) (map[string]map[string]*model.CommonPriceConfig, error)
	GetPItemDataForLocalSip(ctx context.Context, pShopId uint64, pItemId uint64) (service.PrimaryItemData, error)
	GetLocalSipConfigByRegionBatch(ctx context.Context, pRegion string, aRegions []string) (map[string]*model.CommonPriceConfig, error)
	GetAShopDataForLocalSip(ctx context.Context, aShopIds []uint64) (map[uint64]*internal.AShopData, error)
	GetAItemDataBatchForLocalSip(ctx context.Context, pShopId uint64, pItemId uint64, aShopIds []uint64) (map[uint64]*internal.AItemData, error)
	GetShippingFeeForLocalSip(ctx context.Context, pRegion string, queries []model.LocalSipShippingFeeQuery, calcForCreate bool) ([]model.LocalSipShippingFeeResult, error)
	GetInitialHiddenPriceForLocalSip(ctx context.Context, pItemId uint64, pShopId uint64, pRegion string, queries []model.LocalSipHiddenPriceQuery) ([]model.LocalSipHiddenPriceResult, error)

	// CB SIP
	GetHiddenPriceForCbSip(ctx context.Context, pItem *sip_v2_db.MstItemRecord, pShop *sip_db.MstShop, merchantRegion string, pRegion string, aRegion string, weight float64, psite bool) float64
	GetShopServiceFeeForCbSip(ctx context.Context, region string, shopId int64) (float64, error)
	GetShopCommissionFeeForCbSip(ctx context.Context, region string, shopId int64) (float64, error)
	GetHandlingFeeForCbSip(ctx context.Context) (float64, error)
	GetExchangeRateForCbSip(ctx context.Context, sourceCurrency, targetCurrency string, needProcessPrecision bool) (float64, error)
	GetCurrencyForCbSip(ctx context.Context, merchantId uint64, aRegion string, pItemInfo *ib.ProductInfo) (srcCurrency, dstCurrency string, err error)
	GetCountryMarginForCbSip(ctx context.Context, pRegion string, aRegion string) (float64, error)

	GetAllExchangeRateMapForCbSip(ctx context.Context) (map[string]map[string]string, error)
	GetAllCountryMarginMapForCbSip(ctx context.Context) (map[string]map[string]float64, error)
	GetShopSipRateConfigByPShopIdForCbSip(ctx context.Context, pShopId uint64) ([]model.AShopConfigInfo, error)

	// price factor
	GetMerchantExchangeRateInfo(ctx context.Context, merchantId uint64) (*internalExchangeRatePb.ExchangeRateInfo, error)
	GetOrderMartExchangeRate(ctx context.Context, srcCurrency string, dstCurrency string) (float64, error)
}

func GetCalcWeightByAItemAndPItemWeight(aRealWeight int64, pWeight int64) int64 {
	if aRealWeight > 0 {
		return aRealWeight
	}

	return pWeight
}
