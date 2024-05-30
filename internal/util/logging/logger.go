package logging

import (
	"context"

	"git.garena.com/shopee/common/ulog"
)

func GetLogger(ctx context.Context) *ulog.Logger {
	logger := ulog.DefaultLoggerFromContext(ctx)
	if logger == nil {
		logger = ulog.DefaultLogger()
	}

	return logger
}
