package price_sync_db

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	wire.Bind(new(MerchantConfigSettingRepo), new(*MerchantConfigSettingRepoImpl)),
	NewMerchantConfigSettingRepoImpl,
	wire.Bind(new(MerchantConstraintsRepo), new(*MerchantConstraintsRepoImpl)),
	NewMerchantConstraintsRepoImpl,
)
