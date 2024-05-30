package sip_db

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/gdbc/gdbc"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/orm"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/gdbcutil"
)

const (
	aSipShopStatusOffboard = 3
)

func NewSipRepoImpl() *SipRepoImpl {
	return &SipRepoImpl{}
}

type SipRepoImpl struct {
}

func (s *SipRepoImpl) DbSession() orm.DbSession {
	return (*gdbc.DB)(config.GetSipDbClient())
}

func (s *SipRepoImpl) GetAllEditItemPriceAllowList(ctx context.Context, session orm.DbSession) ([]*EditItemPriceAllowList, error) {
	var allRecords []gdbc.Entity
	done := false
	batchSize := 50
	lastId := int64(0)

	for !done {
		records, err := session.Select(&EditItemPriceAllowList{}).Where(gdbc.P("id").GTEQ(lastId)).OrderBy(gdbc.Asc("id")).Limit(batchSize).FetchAll(ctx)
		if err != nil {
			return nil, cerr.Wrap(err, "failed to fetch EditItemPriceAllowList", uint32(pb.Constant_ERROR_DATABASE))
		}
		for _, record := range records {
			allRecords = append(allRecords, record)
			lastId = record.(*EditItemPriceAllowList).Id
		}
		if len(records) < batchSize {
			done = true
		}
	}

	results := make([]*EditItemPriceAllowList, len(allRecords))
	for i, record := range allRecords {
		results[i] = record.(*EditItemPriceAllowList)
	}
	return results, nil
}

func (s *SipRepoImpl) GetMstShopRecordByShopId(ctx context.Context, session orm.DbSession, pShopId uint64) (*MstShop, error) {
	res := &MstShop{}
	err := session.Select(res).Where(gdbc.P("shopid").EQ(pShopId)).Fetch(ctx)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to GetMstShopRecordByShopId for pShopId=%d", pShopId), uint32(pb.Constant_ERROR_DATABASE))
	}
	return res, nil
}

func (s *SipRepoImpl) GetMstShopRecordByShopIdBatch(ctx context.Context, session orm.DbSession, pShopIds []uint64) ([]*MstShop, error) {
	entities, err := session.Select(&MstShop{}).Where(gdbc.P("shopid").IN(pShopIds)).FetchAll(ctx)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to GetMstShopRecordByShopId for pShopIds=%d", pShopIds), uint32(pb.Constant_ERROR_DATABASE))
	}

	mstShops := make([]*MstShop, 0, len(entities))
	for _, e := range entities {
		mstShops = append(mstShops, e.(*MstShop))
	}
	return mstShops, nil
}

func (s *SipRepoImpl) GetAShopDataByAffiShopId(ctx context.Context, session orm.DbSession, affiShopId uint64) (*internal.AShopData, error) {
	records, err := s.getShopMapByAShopIdBatch(ctx, session, []uint64{affiShopId})
	if err != nil {
		return nil, err
	}

	if records == nil || len(records) == 0 {
		return nil, cerr.New("AShopDataRecord not found", uint32(pb.Constant_ERROR_NOT_FOUND))
	}

	return records[0], nil
}

func (s *SipRepoImpl) GetShopMapWithoutOffboardByAShopIdsAndPShopId(ctx context.Context, session orm.DbSession, pShopId uint64, aShopIds []uint64) ([]*internal.AShopData, error) {
	rows, err := session.Select(&internal.AShopData{}).
		Where(gdbc.P("affi_shopid").INUint64(aShopIds...).And(gdbc.P("mst_shopid").EQ(pShopId).And(gdbc.P("sip_shop_status").NEQ(aSipShopStatusOffboard)))).
		FetchAll(ctx)
	if err != nil {
		return nil, cerr.Wrap(err, "query from shop_map_tab failed", uint32(pb.Constant_ERROR_DATABASE))
	}

	result := make([]*internal.AShopData, 0)
	for _, row := range rows {
		result = append(result, row.(*internal.AShopData))
	}
	return result, nil
}

func (s *SipRepoImpl) GetAShopDataByAffiShopIdBatch(ctx context.Context, session orm.DbSession, affiShopIds []uint64) ([]*internal.AShopData, error) {
	return s.getShopMapByAShopIdBatch(ctx, session, affiShopIds)
}

func (s *SipRepoImpl) getShopMapByAShopIdBatch(ctx context.Context, session orm.DbSession, affiShopIds []uint64) ([]*internal.AShopData, error) {
	rows, err := session.Select(&internal.AShopData{}).
		Where(gdbc.P("affi_shopid").INUint64(affiShopIds...)).
		FetchAll(ctx)
	if err != nil {
		return nil, cerr.Wrap(err, "query from shop_map_tab failed", uint32(pb.Constant_ERROR_DATABASE))
	}

	result := make([]*internal.AShopData, 0)
	for _, row := range rows {
		result = append(result, row.(*internal.AShopData))
	}
	return result, nil
}

func (s *SipRepoImpl) GetExchangeRateByCurrency(ctx context.Context, session orm.DbSession, currencyPair string) (*ExchangeRate, error) {
	res := &ExchangeRate{}
	err := session.Select(res).Where(gdbc.P("currency_pair").EQ(currencyPair)).Fetch(ctx)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to GetExchangeRateByCurrency for currencyPair=%s", currencyPair), uint32(pb.Constant_ERROR_DATABASE))
	}
	return res, nil
}

func (s *SipRepoImpl) GetAllExchangeRate(ctx context.Context, session orm.DbSession) ([]*ExchangeRate, error) {
	allResults := make([]gdbc.Entity, 0)
	done := false
	batchSize := 50
	lastId := int64(0)

	for !done {
		data, err := session.Select(&ExchangeRate{}).Where(gdbc.P("id").GTEQ(lastId)).OrderBy(gdbc.Asc("id")).Limit(batchSize).FetchAll(ctx)
		if err != nil {
			return nil, cerr.New(fmt.Sprintf("failed to GetAllExchangeRate"), uint32(pb.Constant_ERROR_DATABASE))
		}
		for _, datum := range data {
			allResults = append(allResults, datum)
			lastId = datum.(*ExchangeRate).ID
		}
		if len(data) < batchSize {
			done = true
		}
	}

	res := make([]*ExchangeRate, len(allResults))
	for i, datum := range allResults {
		res[i] = datum.(*ExchangeRate)
	}
	return res, nil
}

func (s *SipRepoImpl) GetHiddenPriceConfigList(ctx context.Context, session orm.DbSession, pRegion, aRegion string) ([]*LocalHiddenPriceConfigRecord, error) {
	slaveCtx := gdbcutil.ContextWithSlaveCtrl(ctx)
	records, err := session.Select(&LocalHiddenPriceConfigRecord{}).Where(gdbc.P("mst_region").EQ(pRegion).And(gdbc.P("affi_region").EQ(aRegion))).FetchAll(slaveCtx)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to get hidden price config list for pRegion=%s and aRegion=%s", pRegion, aRegion), uint32(pb.Constant_ERROR_DATABASE))
	}

	res := make([]*LocalHiddenPriceConfigRecord, 0)
	for _, record := range records {
		res = append(res, record.(*LocalHiddenPriceConfigRecord))
	}
	return res, nil
}

func (s *SipRepoImpl) GetShippingFeeConfigList(ctx context.Context, session orm.DbSession, pRegion, aRegion string) ([]*LocalShippingFeeConfigRecord, error) {
	slaveCtx := gdbcutil.ContextWithSlaveCtrl(ctx)
	records, err := session.Select(&LocalShippingFeeConfigRecord{}).Where(gdbc.P("mst_region").EQ(pRegion).And(gdbc.P("affi_region").EQ(aRegion))).FetchAll(slaveCtx)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to get shipping price config list for pRegion=%s and aRegion=%s", pRegion, aRegion), uint32(pb.Constant_ERROR_DATABASE))
	}

	res := make([]*LocalShippingFeeConfigRecord, 0)
	for _, record := range records {
		res = append(res, record.(*LocalShippingFeeConfigRecord))
	}
	return res, nil
}

func (s *SipRepoImpl) GetHiddenPriceConfigRecordByWeight(ctx context.Context, session orm.DbSession, mstRegion, affiRegion string, weight int64) (*LocalHiddenPriceConfigRecord, error) {
	var record LocalHiddenPriceConfigRecord

	sql := "SELECT hidden_price FROM local_hidden_price_config_tab " +
		"WHERE mst_region = ? AND affi_region = ? AND weight >= ? ORDER BY weight LIMIT 1"

	row := session.QueryRow(ctx, sql, mstRegion, affiRegion, weight)
	if err := row.StructScan(&record); err != nil {
		if err == gdbc.ErrNoRows {
			return nil, nil
		}
		return nil, cerr.Wrap(err, "query from local_hidden_price_config_tab failed", uint32(pb.Constant_ERROR_DATABASE))
	}

	return &record, nil
}

func (s *SipRepoImpl) GetLocalShippingFeeConfigRecordByWeight(ctx context.Context, session orm.DbSession, mstRegion, affiRegion string, weight int64) (*LocalShippingFeeConfigRecord, error) {
	var record LocalShippingFeeConfigRecord

	sql := "SELECT shipping_fee_price FROM local_shipping_fee_config_tab " +
		"WHERE mst_region = ? AND affi_region = ? AND weight >= ? ORDER BY weight LIMIT 1"

	row := session.QueryRow(ctx, sql, mstRegion, affiRegion, weight)
	if err := row.StructScan(&record); err != nil {
		if err == gdbc.ErrNoRows {
			return nil, nil
		}
		return nil, cerr.Wrap(err, "query from local_shipping_fee_config_tab failed", uint32(pb.Constant_ERROR_DATABASE))
	}

	return &record, nil
}

func (s *SipRepoImpl) GetSystemConfigRecordByType(ctx context.Context, session orm.DbSession, configType int) (*SystemConfigRecord, error) {
	var record SystemConfigRecord
	err := session.Select(&record).
		Where(gdbc.P("type").EQInt(configType)).
		Fetch(ctx)
	if err != nil {
		return nil, cerr.Wrap(err, fmt.Sprintf("query from system_config_tab failed, configType=%v", configType), uint32(pb.Constant_ERROR_DATABASE))
	}

	return &record, nil
}

func (s *SipRepoImpl) GetAllHpfnConfig(ctx context.Context, session orm.DbSession) ([]*HpfnConfig, error) {
	var allRecords []gdbc.Entity
	done := false
	batchSize := 50
	lastId := int64(0)
	for !done {
		records, err := session.Select(&HpfnConfig{}).Where(gdbc.P("id").GTEQ(lastId)).OrderBy(gdbc.Asc("id")).Limit(batchSize).FetchAll(ctx)
		if err != nil {
			return nil, cerr.New(fmt.Sprintf("failed to get all hpfn config, err=%v", err), uint32(pb.Constant_ERROR_DATABASE))
		}

		for _, record := range records {
			allRecords = append(allRecords, record)
			lastId = record.(*HpfnConfig).Id
		}

		if len(records) < batchSize {
			done = true
		}
	}

	results := make([]*HpfnConfig, len(allRecords))
	for i, record := range allRecords {
		results[i] = record.(*HpfnConfig)
	}
	return results, nil
}

func (s *SipRepoImpl) GetHpfnConfigByHpfnKey(ctx context.Context, session orm.DbSession, hpfnKey string) (*HpfnConfig, error) {
	res := &HpfnConfig{}
	err := session.Select(res).Where(gdbc.P("hpfn_key").EQ(hpfnKey)).Fetch(ctx)
	if err != nil {
		return nil, cerr.New(fmt.Sprintf("failed to get hpfn config, err=%v", err), uint32(pb.Constant_ERROR_DATABASE))
	}
	return res, nil
}

func (s *SipRepoImpl) SetAShopDataShopMargin(ctx context.Context, session orm.DbSession, aShopId, pShopId uint64, shopMargin int32) error {
	_, err := session.Update(&internal.AShopData{
		AffiShopid: proto.Uint64(aShopId),
		MstShopid:  proto.Uint64(pShopId),
	}).
		Set(
			gdbc.Field("shop_margin", shopMargin),
			gdbc.Field("mtime", int32(time.Now().Unix())),
		).Where(
		gdbc.P("affi_shopid"),
		gdbc.P("mst_shopid"),
	).Do(ctx)
	if err != nil {
		return cerr.New(fmt.Sprintf("failed to update AShopData shop_margin where aShopId=%d and pShopId=%d", aShopId, pShopId), uint32(pb.Constant_ERROR_DATABASE))
	}
	return nil
}

func (s *SipRepoImpl) SetAShopDataPromoId(ctx context.Context, session orm.DbSession, aShopId, pShopId uint64, promoId uint64) error {
	_, err := session.Update(&internal.AShopData{
		AffiShopid: proto.Uint64(aShopId),
		MstShopid:  proto.Uint64(pShopId),
	}).
		Set(
			gdbc.Field("promotion_id", promoId),
			gdbc.Field("mtime", int32(time.Now().Unix())),
		).Where(
		gdbc.P("affi_shopid"),
		gdbc.P("mst_shopid"),
	).Do(ctx)
	if err != nil {
		return cerr.New(fmt.Sprintf("failed to update AShopData promotion_id where aShopId=%d and pShopId=%d", aShopId, pShopId), uint32(pb.Constant_ERROR_DATABASE))
	}
	return nil
}
