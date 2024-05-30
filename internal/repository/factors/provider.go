package factors

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewCalculationFactorsRepoImpl,
	wire.Struct(new(CalculationFactorsRepoOpts), "*"),
	wire.Bind(new(CalculationFactorsRepo), new(*CalculationFactorsRepoImpl)),
)
