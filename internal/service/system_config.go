package service

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
)

const (
	channelWhitelistType    = 8
	localPriceConfigNewType = 30
)

type SystemConfigService interface {
	GetAllLocalPriceConfig(ctx context.Context) (map[string]map[string]*model.CommonPriceConfig, error)
	GetLocalPriceConfigByRegion(ctx context.Context, primaryRegion, affiRegion string) (*model.CommonPriceConfig, error)
	GetChannelWhitelist(ctx context.Context) ([]int64, error)
}
