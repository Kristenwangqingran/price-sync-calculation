package logic

import (
	"github.com/google/wire"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic/cb_sip_logic"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic/cbsc_logic"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic/common_sip_logic"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic/currency_convert_logic"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/logic/local_sip_logic"
)

var ProviderSet = wire.NewSet(
	cbsc_logic.ProviderSet,
	wire.Bind(new(CbscLogic), new(*cbsc_logic.CbscLogicImpl)),
	cb_sip_logic.ProviderSet,
	wire.Bind(new(CbSipLogic), new(*cb_sip_logic.CbSipLogicImpl)),
	local_sip_logic.ProviderSet,
	wire.Bind(new(LocalSipLogic), new(*local_sip_logic.LocalSipLogicImpl)),
	currency_convert_logic.ProviderSet,
	wire.Bind(new(CurrencyConvertLogic), new(*currency_convert_logic.CurrencyConvertLogicImpl)),
	common_sip_logic.ProviderSet,
	wire.Bind(new(CommonSIPLogic), new(*common_sip_logic.CommonSIPLogicImpl)),
)
