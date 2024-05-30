package gdbcutil

import (
	"context"

	"git.garena.com/shopee/common/gdbc/hardy"
)

func ContextWithMasterCtrl(ctx context.Context) context.Context {
	ctrl := hardy.Ctrl{Role: hardy.Master}
	return hardy.ContextWithCtrl(ctx, ctrl)
}

func ContextWithSlaveCtrl(ctx context.Context) context.Context {
	ctrl := hardy.Ctrl{Role: hardy.Slave}
	return hardy.ContextWithCtrl(ctx, ctrl)
}
