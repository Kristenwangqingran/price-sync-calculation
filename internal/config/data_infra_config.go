package config

import (
	"context"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/common/uniconfig"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type DataInfraConfig struct {
	ClientId              string             `json:"client_id"`
	ClientSecret          string             `json:"client_secret"`
	OrderMartExchangeRate *DataServiceConfig `json:"order_mart_exchange_rate"`
}

type DataServiceConfig struct {
	API     string `json:"api"`
	Version string `json:"version"`
}

func onDataInfraConfigUpdate(e uniconfig.Event) {
	rawNewConfig, err := e.New()
	if err != nil {
		logging.GetLogger(context.Background()).Warn("error getting updated DataInfraConfig value from uniconfig.Event", ulog.Error(err))
		return
	}

	newConfig, ok := rawNewConfig.(*DataInfraConfig)
	if !ok {
		logging.GetLogger(context.Background()).Warn("new config is not DataInfraConfig",
			ulog.String("newVal", cutil.JSONEncode(rawNewConfig)))
		return
	}

	confVal.DataInfraConfig = newConfig

	logging.GetLogger(context.Background()).Info("DataInfraConfig updated", ulog.String("updated", cutil.JSONEncode(newConfig)))
}

func GetDataInfraConfig() *DataInfraConfig {
	if confVal == nil {
		return nil
	}
	return confVal.DataInfraConfig
}
