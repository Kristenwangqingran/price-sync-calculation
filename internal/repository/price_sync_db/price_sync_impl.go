package price_sync_db

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/gdbc/gdbc"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	internalMerchantConfigSettingPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
	internalMerchantConstraintsPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_constraints.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/orm"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type MerchantConfigSettingRepoImpl struct {
}

func NewMerchantConfigSettingRepoImpl() *MerchantConfigSettingRepoImpl {
	return &MerchantConfigSettingRepoImpl{}
}

func (m *MerchantConfigSettingRepoImpl) DbSession() orm.DbSession {
	if config.GetCommonConfig().MerchantConfigSettingTabMigrationToggle.UseNewTable {
		logging.GetLogger(context.Background()).Info("[MerchantConfigSettingTabMigrationToggle.UseNewTable=on] returning price sync DB client")
		return (*gdbc.DB)(config.GetPriceSyncDBClient())
	} else {
		logging.GetLogger(context.Background()).Info("[MerchantConfigSettingTabMigrationToggle.UseNewTable=off] returning listing DB client")
		return (*gdbc.DB)(config.GetListingDbClient())
	}
}

func (m *MerchantConfigSettingRepoImpl) GetMerchantConfigSettingList(ctx context.Context, session orm.DbSession, merchantId uint64) ([]*internalMerchantConfigSettingPb.MerchantConfigSetting, error) {
	rows, err := session.Select(&internalMerchantConfigSettingPb.MerchantConfigSetting{}).
		Where(gdbc.P("merchant_id").EQ(merchantId)).
		FetchAll(ctx)
	if err != nil {
		errMsg := fmt.Sprintf(
			"failed to get merchant config setting from db, merchantId=%d, err=%s",
			merchantId, err.Error())
		logging.GetLogger(ctx).Error(errMsg)
		return nil, cerr.New(errMsg, uint32(priceSyncPriceCalculationPb.Constant_ERROR_DATABASE))
	}

	merchantConfigSettings := make([]*internalMerchantConfigSettingPb.MerchantConfigSetting, len(rows))
	for i, row := range rows {
		merchantConfigSettings[i] = row.(*internalMerchantConfigSettingPb.MerchantConfigSetting)
	}
	return merchantConfigSettings, nil
}

func (m *MerchantConfigSettingRepoImpl) UpsertMerchantConfigSettingList(ctx context.Context, session orm.DbSession, settings []*internalMerchantConfigSettingPb.MerchantConfigSetting) error {
	if config.GetCommonConfig().MerchantConfigSettingTabMigrationToggle.ReadOnly {
		msg := "no write to merchant_config_setting_tab is allowed at the moment"
		logging.GetLogger(ctx).Error(fmt.Sprintf("[MerchantConfigSettingTabMigrationToggle.ReadOnly=on], err=%v", msg))
		return cerr.New(msg, uint32(priceSyncPriceCalculationPb.Constant_ERROR_DATABASE))
	}

	for _, setting := range settings {
		setting.Mtime = proto.Uint32(uint32(time.Now().Unix()))
		setting.Ctime = setting.Mtime
		_, err := session.Create(setting).
			OnConflict().
			Update(gdbc.Fields("profit_rate", "service_fee_rate", "mtime")).
			Do(ctx)
		if err != nil {
			logging.GetLogger(ctx).Error(fmt.Sprintf("error when upsert merchant config settings, err=%v", err))
			return cerr.New(err.Error(), uint32(priceSyncPriceCalculationPb.Constant_ERROR_DATABASE))
		}
	}

	return nil
}

type MerchantConstraintsRepoImpl struct {
}

func NewMerchantConstraintsRepoImpl() *MerchantConstraintsRepoImpl {
	return &MerchantConstraintsRepoImpl{}
}

func (m *MerchantConstraintsRepoImpl) DbSession() orm.DbSession {
	if config.GetCommonConfig().MerchantConstraintsTabMigrationToggle.UseNewTable {
		logging.GetLogger(context.Background()).Info("[MerchantConstraintsTabMigrationToggle.UseNewTable=on] returning price sync DB client")
		return (*gdbc.DB)(config.GetPriceSyncDBClient())
	} else {
		logging.GetLogger(context.Background()).Info("[MerchantConstraintsTabMigrationToggle.UseNewTable=off] returning listing DB client")
		return (*gdbc.DB)(config.GetListingDbClient())
	}
}

func (m *MerchantConstraintsRepoImpl) GetProfitRateLimitListByMerchantRegion(ctx context.Context, session orm.DbSession, region string, merchantRegion string) ([]*internalMerchantConstraintsPb.MerchantConstraints, error) {
	var rows []gdbc.Entity
	var err error
	// better throw error, print warn log first to keep behaviour consistent with existing logic
	if len(merchantRegion) == 0 {
		logging.GetLogger(ctx).Warn(fmt.Sprintf("merchantRegion cannot be empty when getting profit rate limit"))
	}
	if region == "" {
		rows, err = session.
			Select(&internalMerchantConstraintsPb.MerchantConstraints{}).
			Where(gdbc.P("merchant_region").EQ(merchantRegion)).
			FetchAll(ctx)
	} else {
		rows, err = session.
			Select(&internalMerchantConstraintsPb.MerchantConstraints{}).
			Where(gdbc.P("region").EQ(region).And(gdbc.P("merchant_region").EQ(merchantRegion))).
			FetchAll(ctx)
	}

	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf("error getting profit rate limit, err=%v", err))
		return nil, cerr.New(fmt.Sprintf("failed to get profit limit from DB, merchantRegion=%s, err=%v", merchantRegion, err), uint32(priceSyncPriceCalculationPb.Constant_ERROR_DATABASE))
	}
	results := make([]*internalMerchantConstraintsPb.MerchantConstraints, len(rows))
	for i, row := range rows {
		results[i] = row.(*internalMerchantConstraintsPb.MerchantConstraints)
	}
	return results, nil
}

func (m *MerchantConstraintsRepoImpl) UpdateProfitRateLimitByRegionMerchantRegion(ctx context.Context, session orm.DbSession, region, merchantRegion string, profitRateMin, profitRateMax *float64, operator string) error {
	if config.GetCommonConfig().MerchantConstraintsTabMigrationToggle.ReadOnly {
		msg := "no write to merchant_constraints_tab is allowed at the moment"
		logging.GetLogger(ctx).Error(fmt.Sprintf("[MerchantConfigSettingTabMigrationToggle.ReadOnly=on], err=%v", msg))
		return cerr.New(msg, uint32(priceSyncPriceCalculationPb.Constant_ERROR_DATABASE))
	}

	if profitRateMax == nil && profitRateMin == nil {
		logging.GetLogger(ctx).Warn("update profit rate limit, both profitRateMin and profitRateMax are nil, skipped")
		return nil
	}

	updateTo := &ProfitRateLimit{
		Region:         region,         // identifier
		MerchantRegion: merchantRegion, //identifier
		Operator:       operator,
	}

	updatingFields := make([]string, 0)
	updatingFields = append(updatingFields, "operator")

	if profitRateMin != nil {
		updatingFields = append(updatingFields, "profit_rate_min")
		updateTo.ProfitRateMin = calcutil.RoundFloatToInt(*profitRateMin, constant.ProfitRateLimitPrecision, constant.ProfitRateLimitRoundPlace)
	}
	if profitRateMax != nil {
		updatingFields = append(updatingFields, "profit_rate_max")
		updateTo.ProfitRateMax = calcutil.RoundFloatToInt(*profitRateMax, constant.ProfitRateLimitPrecision, constant.ProfitRateLimitRoundPlace)
	}

	updatingFields = append(updatingFields, "mtime")
	updateTo.ModifyTime = uint64(time.Now().Unix())

	_, err := session.Update(updateTo).Set(gdbc.Fields(updatingFields...)).Where(gdbc.P("region"), gdbc.P("merchant_region")).Do(ctx)
	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf("error when updating constraints, region=%s, merchantRegion=%s, new struct=%v, err=%v", region, merchantRegion, *updateTo, err))
		return cerr.New(fmt.Sprintf("failed to update profit limit DB, region=%s, merchantRegion=%s, err=%v", region, merchantRegion, err), uint32(priceSyncPriceCalculationPb.Constant_ERROR_DATABASE))
	}

	return nil
}
