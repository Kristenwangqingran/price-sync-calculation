package calculate

import (
	"github.com/google/wire"
)

var ProvicerSet = wire.NewSet(
	calculateAffiMpskuDmProvider,
	NewMtskuAndMpskuCalculateDm,
)
