package health

import (
	"context"
	"fmt"

	"git.garena.com/shopee/common/gdbc/gdbc"
	"git.garena.com/shopee/core-server/internal-tools/depck"
)

type gdbcHealthCheck struct {
	dbList []*gdbc.DB
}

func NewGDBCHealthCheck(dbList ...*gdbc.DB) depck.Checker {
	return &gdbcHealthCheck{
		dbList: dbList,
	}
}

func (c *gdbcHealthCheck) Check() error {
	for _, db := range c.dbList {
		err := db.Ping(context.Background())
		if err != nil {
			return fmt.Errorf("DB health check failed, err=%v", err)
		}
	}
	return nil
}
