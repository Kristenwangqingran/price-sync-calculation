package constant

const (
	// ChargeCoreBatchCalcHiddenFee
	// can refer https://apidoc.i.ssc.shopeemobile.com/project/2817/interface/api/133195
	ChargeCoreBatchCalcHiddenFee = "/core/hidden_fee/batch_calc_hidden_fee"
)

// subaccount
const (
	GetAllCnscShops = "/internalservice/v1/cnsc_shops/all/"
)

const (
	// fulfillment && logistic
	GetShopChannelListCmd      = "/api/fulfillment/basic/misc/get_shop_channel_list"
	DefaultItemLogisticInfoCmd = "/logistics/shop/default_item_logistic_info/get/"
	// BatchGetSlsLocationInfoCmd get location ids via address ids.
	// can refer api doc https://apidoc.i.ssc.shopeemobile.com/project/667/interface/api/204056
	BatchGetSlsLocationInfoCmd = "/api/v3/logistics/batch_get_sls_location_info/"

	// exchange rate
	GetCurrencyExchangeRate = "/api/inner/miscellaneous/exchange_rate/get"
)
