package health

import (
	"context"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/internal-tools/depck"
	redisck "git.garena.com/shopee/core-server/internal-tools/depck/redis"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

func HealthCheck(ctx context.Context) error {
	checkerList := make([]depck.Checker, 0)

	// check db
	checkerList = append(checkerList, NewGDBCHealthCheck(config.GetListingDbClient()))
	checkerList = append(checkerList, NewGDBCHealthCheck(config.GetPriceSyncDBClient()))

	// check cache
	checkerList = append(checkerList, redisck.New(
		config.GetRedisConfig().Address,
		[]redisck.Extra{
			redisck.Credential(config.GetRedisConfig().Password),
		}...))

	if err := depck.CheckAll(checkerList, false); err != nil {
		logging.GetLogger(ctx).Error("health check failed", ulog.Error(err))
		return err
	}

	return nil
}
