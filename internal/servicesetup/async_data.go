package servicesetup

import (
	"context"
	"fmt"
	"time"

	"git.garena.com/shopee/core-server/core-logic/cutil"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/data_infra"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type AsyncData interface {
	AsyncSetOrderMartExchangeRate()
}

type AsyncDataImpl struct {
	Cache cache.CommonCache
}

type AsyncDataOpt struct {
	Cache cache.CommonCache
}

func NewAsyncData(opts *AsyncDataOpt) *AsyncDataImpl {
	return &AsyncDataImpl{
		Cache: opts.Cache,
	}
}

func (s *AsyncDataImpl) AsyncSetOrderMartExchangeRate() {
	ds := data_infra.NewDataService()

	// for initial data
	interval := s.aSyncQuery(ds)

	// for refreshing
	go func() {
		for {
			wait := time.After(interval)
			<-wait

			interval = s.aSyncQuery(ds)
		}
	}()
}

func (s *AsyncDataImpl) aSyncQuery(ds *data_infra.DataService) time.Duration {
	ctx := context.Background()

	var interval time.Duration

	data, success := ds.GetOrderMartExchangeRateList(ctx)
	if !success {
		interval = time.Duration(config.GetCommonConfig().OrderMartExchangeRateRetrySeconds) * time.Second

		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to get order mart exchange rate from data infra, interval=%v",
			interval.String()))
		return interval
	}

	logging.GetLogger(ctx).Info(fmt.Sprintf(
		"success to get order mart exchange rate from data infra, data=%v", cutil.JSONEncode(data)))

	success = s.Cache.SetOrderMartExchangeRateBatch(ctx, data)
	if success {
		interval = time.Duration(config.GetCommonConfig().OrderMartExchangeRateRefreshSeconds) * time.Second

		logging.GetLogger(ctx).Info(fmt.Sprintf(
			"success to get order mart exchange rate from data infra and set cache, next interval=%v", interval.String()))
	} else {
		interval = time.Duration(config.GetCommonConfig().OrderMartExchangeRateRetrySeconds) * time.Second

		logging.GetLogger(ctx).Error(fmt.Sprintf(
			"failed to set order mart exchange rate into cache, next interval=%v", interval.String()))
	}
	return interval
}
