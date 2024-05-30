package servicesetup

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	GetHttpClient,

	NewAsyncData,
	wire.Struct(new(AsyncDataOpt), "*"),
	wire.Bind(new(AsyncData), new(*AsyncDataImpl)),
)
