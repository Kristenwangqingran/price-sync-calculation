package service

import (
	"context"

	"git.garena.com/shopee/common/ulog"
	listing_sip "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/marketplace_listing_upload_crossupload_api.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
)

type ListingSIPService interface {
	GetPrimaryItemIdByAffiItemIds(ctx context.Context, affiShopId uint64, affiItemIds []uint64) ([]*listing_sip.SIPShopItem, error)
	GetAffiItemIdsByPrimaryItemId(ctx context.Context, pShopId uint64, pItemId uint64, aShopIds []uint64) ([]*listing_sip.SIPShopItem, error)
}

type ListingSIPServiceImpl struct {
	proxy spex.ListingSIP
}

func NewListingSIPService(proxy spex.ListingSIP) ListingSIPService {
	return &ListingSIPServiceImpl{proxy: proxy}
}

func (l *ListingSIPServiceImpl) GetPrimaryItemIdByAffiItemIds(ctx context.Context, affiShopId uint64, affiItemIds []uint64) ([]*listing_sip.SIPShopItem, error) {
	resp, err := l.proxy.GetPItemByAItemIds(ctx, affiShopId, affiItemIds)
	if err != nil {
		ulog.DefaultLoggerFromContext(ctx).Error("error when get primary item by a itemids", ulog.Error(err))
		return nil, err
	}
	return resp.GetPitemList(), nil
}

func (l *ListingSIPServiceImpl) GetAffiItemIdsByPrimaryItemId(ctx context.Context, pShopId uint64, pItemId uint64, aShopIds []uint64) ([]*listing_sip.SIPShopItem, error) {
	resp, err := l.proxy.GetAItemByPItemIds(ctx, pShopId, pItemId, aShopIds)
	if err != nil {
		ulog.DefaultLoggerFromContext(ctx).Error("error when get p item by a item id")
	}
	return resp.GetAitemList(), nil
}