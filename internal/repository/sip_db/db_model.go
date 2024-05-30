package sip_db

import (
	"git.garena.com/shopee/common/gdbc/gdbc/tablereflect"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
)

const (
	SystemConfigTypeContryMargin          = 9
	SystemConfigTypeRegionRateTableConfig = 20
	SystemConfigTypeCnscSipRateLimit      = 24
	SystemConfigTypeDefaultSipRate        = 26
)

type CbOption = int

const (
	CbOptionLocalShop CbOption = 0
	CbOptionCbShop    CbOption = 1
)

func init() {
	tablereflect.TypeInit(
		&LocalHiddenPriceConfigRecord{},
		tablereflect.Table("local_hidden_price_config_tab"),
	)
	tablereflect.TypeInit(
		&LocalShippingFeeConfigRecord{},
		tablereflect.Table("local_shipping_fee_config_tab"),
	)
	tablereflect.TypeInit(
		&SystemConfigRecord{},
		tablereflect.Table("system_config_tab"),
	)
	tablereflect.TypeInit(
		&ExchangeRate{},
		tablereflect.Table("exchange_rate_tab"),
	)
	tablereflect.TypeInit(
		&HpfnConfig{},
		tablereflect.Table("hpfn_config_tab"),
	)

	tablereflect.TypeInit(
		&internal.AShopData{},
		tablereflect.Table("shop_map_tab"),
	)

	tablereflect.TypeInit(
		&MstShop{},
		tablereflect.Table("mst_shop_tab"),
	)

	tablereflect.TypeInit(
		&EditItemPriceAllowList{},
		tablereflect.Table("edit_item_price_allow_list_tab"),
	)
}

type EditItemPriceAllowList struct {
	Id        int64 `gdbc:"column=id"`
	MstShopId int64 `gdbc:"column=mst_shopid"`
	Ctime     int64 `gdbc:"column=ctime"`
}

type MstShop struct {
	ShopId           int64    `gdbc:"column=shopid"`
	UserId           int64    `gdbc:"column=userid"`
	Country          string   `gdbc:"column=country"`
	SyncStatus       int      `gdbc:"column=sync_status"`
	ShopStatus       int      `gdbc:"column=shop_status"`
	OpenStockPect    int      `gdbc:"column=open_stock_pect"`
	MstSafeStock     int      `gdbc:"column=mst_safe_stock"` // use ShopHelper.GetShopSafeStock to fetch
	StockCountryProp string   `gdbc:"column=stock_country_prop"`
	InnerFlag        int      `gdbc:"column=inner_flag"`
	Extinfo          string   `gdbc:"column=extinfo"`
	RateTableCfg     string   `gdbc:"column=rate_table_cfg"`
	CbOption         CbOption `gdbc:"column=cb_option"`
	Ctime            int64    `gdbc:"column=ctime"`
	Mtime            int64    `gdbc:"column=mtime"`
}

func (m *MstShop) IsCbShop() bool {
	if m == nil {
		return false
	}
	return m.CbOption == CbOptionCbShop
}

type LocalHiddenPriceConfigRecord struct {
	Id          int64  `gdbc:"column=id"`
	MstRegion   string `gdbc:"column=mst_region"`
	AffiRegion  string `gdbc:"column=affi_region"`
	Weight      int64  `gdbc:"column=weight"`
	HiddenPrice int64  `gdbc:"column=hidden_price"`
	Ctime       int64  `gdbc:"column=ctime"`
}

type LocalShippingFeeConfigRecord struct {
	Id               int64  `gdbc:"column=id"`
	MstRegion        string `gdbc:"column=mst_region"`
	AffiRegion       string `gdbc:"column=affi_region"`
	Weight           int64  `gdbc:"column=weight"`
	ShippingFeePrice int64  `gdbc:"column=shipping_fee_price"`
	Ctime            int64  `gdbc:"column=ctime"`
}

type SystemConfigRecord struct {
	ID         int64  `gdbc:"column=id"`
	Type       int    `gdbc:"column=type"`
	ConfigData string `gdbc:"column=config_data"`
	Status     int    `gdbc:"column=status"`
	Ctime      int64  `gdbc:"column=ctime"`
	Mtime      int64  `gdbc:"column=mtime"`
}

type ExchangeRate struct {
	ID             int64  `gdbc:"column=id"`
	CurrencyPair   string `gdbc:"column=currency_pair"`
	SourceCurrency string `gdbc:"column=source_currency"`
	TargetCurrency string `gdbc:"column=target_currency"`
	ExchangeRate   string `gdbc:"column=exchange_rate"`
	Ctime          int64  `gdbc:"column=ctime"`
	Mtime          int64  `gdbc:"column=mtime"`
}

type HpfnConfig struct {
	Id          int64  `gdbc:"column=id"`
	HpfnKey     string `gdbc:"column=hpfn_key"`
	WeightRange int64  `gdbc:"column=weight_range"`
	StartPrice  int64  `gdbc:"column=start_price"`
	StartWeight int64  `gdbc:"column=start_weight"`
	RoundSize   int64  `gdbc:"column=round_size"`
	Price       int64  `gdbc:"column=price"`
	WeightStep  int64  `gdbc:"column=weight_step"`
	Adjustment  int64  `gdbc:"column=adjustment"`
	DescInfo    string `gdbc:"column=desc_info"`
	Ctime       int64  `gdbc:"column=ctime"`
	Mtime       int64  `gdbc:"column=mtime"`
}
