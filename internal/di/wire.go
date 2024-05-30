//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/cache"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/calculate"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/data"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/processor"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository"
	db2 "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/servicesetup"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/spex"
)

var providerSet = wire.NewSet(
	cache.ProviderSet,
	config.ProviderSet,
	db2.ProviderSet,
	calculate.ProvicerSet,
	data.ProviderSet,
	dm.ProviderSet,
	service.ProviderSet,
	servicesetup.ProviderSet,
	spex.ProviderSet,
	repository.ProviderSet,
	processor.ProviderSet,
	logic.ProviderSet,
)

func CreateApplication() (*processor.CalculationServiceImpl, error) {
	panic(wire.Build(providerSet))
}
