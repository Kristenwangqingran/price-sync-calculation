package config

import (
	"context"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/common/uniconfig"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

const (
	defaultMaxBatchSizeForCalcGlobalDiscountInfoByItemIds = 10
	defaultMaxBatchSizeForShopGetMerchantList             = 10
	defaultMaxBatchSizeForIBSGetProductInfoByItemIds      = 50
	defaultMaxBatchSizeForSlsBatchCalcHiddenFee           = 20
	defaultMaxBatchSizeForSlsBatchGetSlsLocationInfo      = 10
	defaultMaxBatchSizeForCnscCheckWhiteList              = 800
	defaultMaxBatchSizeForGetAccount                      = 50
	defaultMaxBatchSizeForGetUserIdByShopId               = 50
)

// BatchConfig contains configures that is used for batch api
type BatchConfig struct {
	MaxBatchSizeForCalcGlobalDiscountInfoByItemIds uint32 `json:"max_batch_size_for_calc_global_discount_info_by_item_ids"`
	MaxBatchSizeForShopGetMerchantList             uint32 `json:"max_batch_size_for_shop_get_merchant_list"`
	MaxBatchSizeForIBSGetProductInfoByItemIds      uint32 `json:"max_batch_size_for_ibs_get_product_info_by_item_ids"`
	MaxBatchSizeForSlsBatchCalcHiddenFee           uint32 `json:"max_batch_size_for_sls_batch_calc_hidden_fee"`
	MaxBatchSizeForSlsBatchGetSlsLocationInfo      uint32 `json:"max_batch_size_for_sls_batch_get_sls_location_info"`
	MaxBatchSizeForCnscCheckWhiteList              uint32 `json:"max_batch_size_for_cnsc_check_white_list"`
	MaxBatchSizeForGetAccount                      uint32 `json:"max_batch_size_for_get_account"`
	MaxBatchSizeForGetUserIdByShopId               uint32 `json:"max_batch_size_for_get_user_id_by_shop_id"`
}

func onBatchConfigUpdate(e uniconfig.Event) {
	rawNewConfig, err := e.New()
	if err != nil {
		logging.GetLogger(context.Background()).Warn("error getting updated BatchConfig value from uniconfig.Event", ulog.Error(err))
		return
	}

	newConfig, ok := rawNewConfig.(*BatchConfig)
	if !ok {
		logging.GetLogger(context.Background()).Warn("new config is not a BatchConfig",
			ulog.String("newVal", cutil.JSONEncode(rawNewConfig)))
		return
	}

	confVal.BatchCfg = newConfig
	applyBatchConfig()

	logging.GetLogger(context.Background()).Info("BatchConfig is updated", ulog.String("batch_config", cutil.JSONEncode(newConfig)))
}

func applyBatchConfig() {
	if confVal == nil || confVal.BatchCfg == nil {
		logging.GetLogger(context.Background()).Warn("BatchConfig is empty")
	}

	applyDefaultValForBatchConfig()
}

func applyDefaultValForBatchConfig() {
	if confVal == nil {
		confVal = &Config{}
	}

	if confVal.BatchCfg == nil {
		confVal.BatchCfg = &BatchConfig{}
	}

	batchCfg := confVal.BatchCfg
	if batchCfg.MaxBatchSizeForCalcGlobalDiscountInfoByItemIds == 0 {
		batchCfg.MaxBatchSizeForCalcGlobalDiscountInfoByItemIds = defaultMaxBatchSizeForCalcGlobalDiscountInfoByItemIds
	}

	if batchCfg.MaxBatchSizeForShopGetMerchantList == 0 {
		batchCfg.MaxBatchSizeForShopGetMerchantList = defaultMaxBatchSizeForShopGetMerchantList
	}

	if batchCfg.MaxBatchSizeForIBSGetProductInfoByItemIds == 0 {
		batchCfg.MaxBatchSizeForIBSGetProductInfoByItemIds = defaultMaxBatchSizeForIBSGetProductInfoByItemIds
	}

	if batchCfg.MaxBatchSizeForSlsBatchCalcHiddenFee == 0 {
		batchCfg.MaxBatchSizeForSlsBatchCalcHiddenFee = defaultMaxBatchSizeForSlsBatchCalcHiddenFee
	}

	if batchCfg.MaxBatchSizeForSlsBatchGetSlsLocationInfo == 0 {
		batchCfg.MaxBatchSizeForSlsBatchGetSlsLocationInfo = defaultMaxBatchSizeForSlsBatchGetSlsLocationInfo
	}

	if batchCfg.MaxBatchSizeForCnscCheckWhiteList == 0 {
		batchCfg.MaxBatchSizeForCnscCheckWhiteList = defaultMaxBatchSizeForCnscCheckWhiteList
	}

	if batchCfg.MaxBatchSizeForGetAccount == 0 {
		batchCfg.MaxBatchSizeForGetAccount = defaultMaxBatchSizeForGetAccount
	}

	if batchCfg.MaxBatchSizeForGetUserIdByShopId == 0 {
		batchCfg.MaxBatchSizeForGetUserIdByShopId = defaultMaxBatchSizeForGetUserIdByShopId
	}
}

func GetBatchConfig() *BatchConfig {
	if confVal == nil {
		return nil
	}
	return confVal.BatchCfg
}
