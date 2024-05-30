package sip_v2_db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/common/gdbc/datum"
	"git.garena.com/shopee/common/gdbc/gdbc"
	"git.garena.com/shopee/common/gdbc/hardy/sharding"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/orm"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/gdbcutil"
)

func NewSipV2RepoImpl() *SipV2RepoImpl {
	return &SipV2RepoImpl{}
}

type SipV2RepoImpl struct {
}

func (s *SipV2RepoImpl) DbSession() orm.DbSession {
	return (*gdbc.DB)(config.GetSipV2DbClient())
}

func (s *SipV2RepoImpl) getShardedCtx(ctx context.Context, pShopId uint64, useMaster bool) context.Context {
	hint := &sharding.Hint{Values: map[string]datum.Datum{"mst_shopid": datum.Uint64(pShopId)}}
	ctx = sharding.ContextWithShardHint(ctx, hint)
	if useMaster {
		ctx = gdbcutil.ContextWithMasterCtrl(ctx)
	}
	return ctx
}

func (s *SipV2RepoImpl) GetMstItemRecordBatch(ctx context.Context, session orm.DbSession, primaryShopId uint64, primaryItemIds []uint64) ([]*MstItemRecord, error) {
	hint := &sharding.Hint{Values: map[string]datum.Datum{"mst_shopid": datum.Uint64(primaryShopId)}}
	ctx = sharding.ContextWithShardHint(ctx, hint)

	rows, err := session.Select(&MstItemRecord{}).
		Where(gdbc.P("itemid").INUint64(primaryItemIds...)).
		FetchAll(ctx)
	if err != nil {
		return nil, cerr.Wrap(err, "query from mst_item_tab failed", uint32(pb.Constant_ERROR_DATABASE))
	}

	result := make([]*MstItemRecord, 0)
	for _, row := range rows {
		result = append(result, row.(*MstItemRecord))
	}
	return result, nil
}

func (s *SipV2RepoImpl) GetAItemDataBatch(ctx context.Context, session orm.DbSession, primaryShopId uint64, affiItemIds []uint64) ([]*internal.AItemData, error) {
	hint := &sharding.Hint{Values: map[string]datum.Datum{"mst_shopid": datum.Uint64(primaryShopId)}}
	ctx = sharding.ContextWithShardHint(ctx, hint)

	rows, err := session.Select(&internal.AItemData{}).
		Where(gdbc.P("affi_itemid").INUint64(affiItemIds...)).
		FetchAll(ctx)
	if err != nil {
		return nil, cerr.Wrap(err, "query from item_map_tab failed", uint32(pb.Constant_ERROR_DATABASE))
	}

	result := make([]*internal.AItemData, 0)
	for _, row := range rows {
		result = append(result, row.(*internal.AItemData))
	}
	return result, nil
}

func (s *SipV2RepoImpl) GetAItemData(ctx context.Context, session orm.DbSession, primaryShopId uint64, affiItemId uint64) ([]*internal.AItemData, error) {
	hint := &sharding.Hint{Values: map[string]datum.Datum{"mst_shopid": datum.Uint64(primaryShopId)}}
	ctx = sharding.ContextWithShardHint(ctx, hint)

	rows, err := session.Select(&internal.AItemData{}).
		Where(gdbc.P("affi_itemid").EQUint64(affiItemId)).
		FetchAll(ctx)
	if err != nil {
		return nil, cerr.Wrap(err, "query from item_map_tab failed", uint32(pb.Constant_ERROR_DATABASE))
	}
	result := make([]*internal.AItemData, 0)
	for _, row := range rows {
		result = append(result, row.(*internal.AItemData))
	}

	return result, nil
}

func (s *SipV2RepoImpl) GetMtskuMapByAffiMskuIds(ctx context.Context, session orm.DbSession, mstShopId uint64, affiMskuIds []model.ItemModelId) ([]*MskuMapRecord, error) {
	hint := &sharding.Hint{Values: map[string]datum.Datum{"mst_shopid": datum.Uint64(mstShopId)}}
	ctx = sharding.ContextWithShardHint(ctx, hint)

	placeHolders := make([]string, 0)
	queryArgs := make([]interface{}, 0)
	for _, affiMskuId := range affiMskuIds {
		placeHolders = append(placeHolders, "(?, ?)")
		queryArgs = append(queryArgs, affiMskuId.ItemId, affiMskuId.ModelId)
	}

	sql := fmt.Sprintf("SELECT affi_itemid,affi_modelid,affi_shopid,mst_itemid,mst_modelid FROM `msku_map_tab` "+
		"WHERE (affi_itemid, affi_modelid) in (%s)", strings.Join(placeHolders, ","))

	rows, err := session.Query(ctx, sql, queryArgs...)
	if err != nil {
		return nil, cerr.Wrap(err, "query from msku_map_tab failed", uint32(pb.Constant_ERROR_DATABASE))
	}
	defer rows.Close()

	result := make([]*MskuMapRecord, 0)
	for rows.Next() {
		var record MskuMapRecord
		if err := rows.StructScan(&record); err != nil {
			return nil, cerr.Wrap(err, "GetByAffiMskuId failed", uint32(pb.Constant_ERROR_DATABASE))
		}

		result = append(result, &record)
	}

	return result, nil
}

func (s *SipV2RepoImpl) UpdateAItemMargin(ctx context.Context, session orm.DbSession, primaryShopId uint64, affiItemId uint64, aItemMargin int32) error {
	hint := &sharding.Hint{Values: map[string]datum.Datum{"mst_shopid": datum.Uint64(primaryShopId)}}
	ctx = sharding.ContextWithShardHint(ctx, hint)
	_, err := session.Update(&internal.AItemData{
		MstShopid:  proto.Uint64(primaryShopId), // TODO: to remove after DB split
		AffiItemid: proto.Uint64(affiItemId),
	}).
		Set(
			gdbc.Field("item_margin", aItemMargin),
			gdbc.Field("mtime", int32(time.Now().Unix())),
		).Where(
		gdbc.P("affi_itemid"),
		gdbc.P("mst_shopid"),
	).Do(ctx)
	if err != nil {
		return cerr.New(fmt.Sprintf("failed to update AItemData item_margin where affiItemId=%d and pShopId=%d", affiItemId, primaryShopId), uint32(pb.Constant_ERROR_DATABASE))
	}
	return nil
}

func (s *SipV2RepoImpl) UpdateAItemRealWeight(ctx context.Context, session orm.DbSession, primaryShopId uint64, affiItemId uint64, aItemRealWeight int32) error {
	hint := &sharding.Hint{Values: map[string]datum.Datum{"mst_shopid": datum.Uint64(primaryShopId)}}
	ctx = sharding.ContextWithShardHint(ctx, hint)
	_, err := session.Update(&internal.AItemData{
		MstShopid:  proto.Uint64(primaryShopId), // TODO: to remove after DB split
		AffiItemid: proto.Uint64(affiItemId),
	}).
		Set(
			gdbc.Field("affi_real_weight", aItemRealWeight),
			gdbc.Field("mtime", int32(time.Now().Unix())),
		).Where(
		gdbc.P("affi_itemid"),
		gdbc.P("mst_shopid"),
	).Do(ctx)
	if err != nil {
		return cerr.New(fmt.Sprintf("failed to update AItemData affi_real_weight where affiItemId=%d and pShopId=%d", affiItemId, primaryShopId), uint32(pb.Constant_ERROR_DATABASE))
	}
	return nil
}
