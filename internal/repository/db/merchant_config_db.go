package db

import (
	"context"
	"fmt"
	"time"

	"git.garena.com/shopee/common/gdbc/gdbc"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	internalMerchantConfigSettingPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_config_setting.pb"
	internalMerchantConstraintsPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_constraints.pb"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/price_sync_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"

	"github.com/golang/protobuf/proto"
)

type MerchantConfigDB struct {
	merchantConfigSettingRepo price_sync_db.MerchantConfigSettingRepo
	merchantConstraintsRepo   price_sync_db.MerchantConstraintsRepo
}

func NewMerchantConfigDB(merchantConfigSettingRepo price_sync_db.MerchantConfigSettingRepo, merchantConstraintsRepo price_sync_db.MerchantConstraintsRepo) *MerchantConfigDB {
	return &MerchantConfigDB{
		merchantConfigSettingRepo: merchantConfigSettingRepo,
		merchantConstraintsRepo:   merchantConstraintsRepo,
	}
}

// GetMerchantConfigSettingList get merchant config setting by merchantId.
// If not found, then return empty and no error.
func (d *MerchantConfigDB) GetMerchantConfigSettingList(ctx context.Context, merchantId uint64) ([]*internalMerchantConfigSettingPb.MerchantConfigSetting, error) {
	session := d.merchantConfigSettingRepo.DbSession()
	return d.merchantConfigSettingRepo.GetMerchantConfigSettingList(ctx, session, merchantId)
}

func (d *MerchantConfigDB) SetMerchantConfigSettingList(ctx context.Context,
	merchantId uint64, settings []*internalMerchantConfigSettingPb.MerchantConfigSetting) error {
	db, ok := d.merchantConfigSettingRepo.DbSession().(*gdbc.DB)
	if !ok {
		return cerr.New(fmt.Sprintf("type insert for db failed|session=%T", d.merchantConfigSettingRepo.DbSession()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_DATABASE))
	}

	tx, err := db.Tx(ctx)
	if err != nil {
		return cerr.Wrap(err, "begin tx failed", uint32(priceSyncPriceCalculationPb.Constant_ERROR_DATABASE))
	}

	err = d.doSetMerchantConfigSettingList(ctx, tx, merchantId, settings)
	if err != nil {
		rErr := tx.Rollback()
		if rErr != nil {
			logging.GetLogger(ctx).Error(fmt.Sprintf("could not rollback transaction|err=%v", rErr))
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf("could not commit transaction|err=%v", err))
		return cerr.Wrap(err, "commit tx failed", uint32(priceSyncPriceCalculationPb.Constant_ERROR_DATABASE))
	}

	return nil
}

// GetProfitRateLimitList get merchant constraints by region and merchantRegion
// if region is empty, it fetches all merchant constraints using merchantRegion given.
// if not found, return empty and no error.
func (d *MerchantConfigDB) GetProfitRateLimitList(ctx context.Context, region, merchantRegion string) ([]*internalMerchantConstraintsPb.MerchantConstraints, error) {
	session := d.merchantConstraintsRepo.DbSession()
	return d.merchantConstraintsRepo.GetProfitRateLimitListByMerchantRegion(ctx, session, region, merchantRegion)
}

func (d *MerchantConfigDB) UpdateProfitRateLimit(ctx context.Context, region, merchantRegion string, profitRateMin, profitRateMax *float64, operator string) error {
	session := d.merchantConstraintsRepo.DbSession()
	return d.merchantConstraintsRepo.UpdateProfitRateLimitByRegionMerchantRegion(ctx, session, region, merchantRegion, profitRateMin, profitRateMax, operator)
}

func (d *MerchantConfigDB) doSetMerchantConfigSettingList(ctx context.Context,
	session gdbc.Session, merchantId uint64, settings []*internalMerchantConfigSettingPb.MerchantConfigSetting) error {
	records, err := d.merchantConfigSettingRepo.GetMerchantConfigSettingList(ctx, session, merchantId)
	if err != nil {
		return err
	}

	shopMap := make(map[uint64]*internalMerchantConfigSettingPb.MerchantConfigSetting)
	for _, record := range records {
		shopMap[record.GetShopId()] = record
	}

	targetSettings := make([]*internalMerchantConfigSettingPb.MerchantConfigSetting, 0)

	now := time.Now().Unix()
	for _, setting := range settings {
		record, ok := shopMap[setting.GetShopId()]
		if ok {
			record.ProfitRate = setting.ProfitRate
			record.ServiceFeeRate = setting.ServiceFeeRate
			record.Mtime = proto.Uint32(uint32(now))

			targetSettings = append(targetSettings, record)
		} else {
			targetSettings = append(targetSettings, &internalMerchantConfigSettingPb.MerchantConfigSetting{
				MerchantId:     proto.Uint64(merchantId),
				ShopId:         proto.Uint64(setting.GetShopId()),
				Region:         proto.String(setting.GetRegion()),
				ProfitRate:     setting.ProfitRate,
				ServiceFeeRate: setting.ServiceFeeRate,
				Ctime:          proto.Uint32(uint32(now)),
				Mtime:          proto.Uint32(uint32(now)),
			})
		}
	}

	return d.merchantConfigSettingRepo.UpsertMerchantConfigSettingList(ctx, session, targetSettings)
}
