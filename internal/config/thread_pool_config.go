package config

import (
	"context"
	"time"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/common/uniconfig"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/threadpool"
)

type ThreadPoolConfig struct {
	Concurrent     int    `json:"concurrent" yaml:"concurrent"`
	IdleTimeout    int    `json:"idle_timeout" yaml:"idle_timeout"`
	MaxWaitTimeout int    `json:"max_wait_timeout" yaml:"max_wait_timeout"`
	PoolName       string `json:"pool_name" yaml:"pool_name"`
}

func applyThreadPoolConfig() {
	if confVal == nil || confVal.ThreadPoolConfig == nil {
		panic("failed to applyThreadPoolConfig, since ThreadPoolConfig is nil")
	}

	threadpool.InitThreadPool(combineCfg(confVal.ThreadPoolConfig))
}

func onThreadPoolConfigUpdate(e uniconfig.Event) {
	rawNewConfig, err := e.New()
	if err != nil {
		logging.GetLogger(context.Background()).Warn(
			"error getting updated ThreadPoolConfig value from uniconfig.Event", ulog.Error(err))
		return
	}

	newConfig, ok := rawNewConfig.(*ThreadPoolConfig)
	if !ok {
		logging.GetLogger(context.Background()).Warn("new config is not ThreadPoolConfig",
			ulog.String("newVal", cutil.JSONEncode(rawNewConfig)))
		return
	}

	confVal.ThreadPoolConfig = newConfig

	logging.GetLogger(context.Background()).Info("ThreadPoolConfig updated",
		ulog.String("updated", cutil.JSONEncode(newConfig)))

	threadpool.GetThreadPool().OnConfigUpdate(combineCfg(newConfig))
}

func combineCfg(cfg *ThreadPoolConfig) *threadpool.Config {
	confInfo := &threadpool.Config{
		Concurrent:     100,
		IdleTimeout:    15 * time.Second,
		MaxWaitTimeout: 10 * time.Minute,
		PoolName:       "default_pool_name",
	}

	if cfg.Concurrent > 0 {
		confInfo.Concurrent = cfg.Concurrent
	}

	if cfg.IdleTimeout > 0 {
		confInfo.IdleTimeout = time.Duration(cfg.IdleTimeout) * time.Millisecond
	}

	if cfg.MaxWaitTimeout > 0 {
		confInfo.MaxWaitTimeout = time.Duration(cfg.MaxWaitTimeout) * time.Millisecond
	}

	if len(cfg.PoolName) > 0 {
		confInfo.PoolName = cfg.PoolName
	}

	return confInfo
}
