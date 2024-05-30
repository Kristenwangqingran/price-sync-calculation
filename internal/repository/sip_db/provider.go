package sip_db

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	wire.Bind(new(SipRepo), new(*SipRepoImpl)),
	NewSipRepoImpl,
)
