package infra

import (
	"github.com/google/wire"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/infra/snowflake"
)

var ProviderSet = wire.NewSet(
	snowflake.NewGenIDWorker,
)
