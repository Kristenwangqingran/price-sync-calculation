package hpfn_config

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewHpfnConfigRepoImpl,
	wire.Bind(new(HpfnConfigRepo), new(*HpfnConfigRepoImpl)),
)
