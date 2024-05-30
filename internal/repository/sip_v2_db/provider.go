package sip_v2_db

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	wire.Bind(new(SipV2Repo), new(*SipV2RepoImpl)),
	NewSipV2RepoImpl,
)
