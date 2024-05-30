package config

import (
	"context"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/common/uniconfig"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type OPLConfig struct {
	CustomizedOplRegionBlackList map[string]bool `json:"customized_opl_region_black_list"`
}

func onOPLConfigUpdate(e uniconfig.Event) {
	rawNewConfig, err := e.New()
	if err != nil {
		logging.GetLogger(context.Background()).Warn("error getting updated OPLConfig value from uniconfig.Event", ulog.Error(err))
		return
	}

	newConfig, ok := rawNewConfig.(*OPLConfig)
	if !ok {
		logging.GetLogger(context.Background()).Warn("new config is not a OPLConfig",
			ulog.String("newVal", cutil.JSONEncode(rawNewConfig)))
		return
	}

	confVal.OPLConfig = newConfig

	logging.GetLogger(context.Background()).Info("OPLConfig is updated", ulog.String(KeyOPLConfig, cutil.JSONEncode(newConfig)))
}

func GetOPLConfig() *OPLConfig {
	return confVal.OPLConfig
}
