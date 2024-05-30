package sip_v2_db

import (
	"git.garena.com/shopee/common/gdbc/gdbc/tablereflect"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
)

func init() {
	tablereflect.TypeInit(
		&internal.AItemData{},
		tablereflect.Table("item_map_tab"),
	)

	tablereflect.TypeInit(
		&MstItemRecord{},
		tablereflect.Table("mst_item_tab"),
	)

	tablereflect.TypeInit(
		&MskuMapRecord{},
		tablereflect.Table("msku_map_tab"),
	)
	tablereflect.TypeInit(
		&Msku{},
		tablereflect.Table("msku_tab"),
	)
}

type MstItemRecord struct {
	ItemId uint64 `gdbc:"column=itemid"`

	Weight       int64  `gdbc:"column=weight"`
	RateTableCfg string `gdbc:"column=rate_table_cfg"`
}

type MskuMapRecord struct {
	Id uint64 `gdbc:"column=id"`

	AffiItemId  uint64 `gdbc:"column=affi_itemid"`
	AffiModelId uint64 `gdbc:"column=affi_modelid"`
	AffiShopId  uint64 `gdbc:"column=affi_shopid"`

	MstItemId  uint64 `gdbc:"column=mst_itemid"`
	MstModelId uint64 `gdbc:"column=mst_modelid"`
}

type MskuStatus = int

const (
	MskuStatusDelete     MskuStatus = 0
	MskuStatusNormal     MskuStatus = 1
	MskuStatusDeleteWait MskuStatus = 2
)

type Msku struct {
	MskuId        string `gdbc:"column=msku_id"`
	ItemId        int64  `gdbc:"column=itemid"`
	ModelId       int64  `gdbc:"column=modelid"`
	ShopId        int64  `gdbc:"column=shopid"`
	ReservedStock int    `gdbc:"column=reserved_stock"`
	PreAddStock   string `gdbc:"column=pre_add_stock"`
	NetPrice      int64  `gdbc:"column=net_price"`
	ItemPrice     int64  `gdbc:"column=item_price"`
	PriceStrategy int    `gdbc:"column=price_strategy"`
	Status        int    `gdbc:"column=status"`
	Ctime         int64  `gdbc:"column=ctime"`
	Mtime         int64  `gdbc:"column=mtime"`
}

type MstItemModel struct {
	Id          int64  `gdbc:"column=id;primary_key"`
	ModelId     int64  `gdbc:"column=modelid"`
	ItemId      int64  `gdbc:"column=itemid"`
	Name        string `gdbc:"column=name"`
	OrigPrice   int64  `gdbc:"column=orig_price"`
	PromotionId int64  `gdbc:"column=promotionid"`
	PromoPrice  int64  `gdbc:"column=promo_price"`
	PromoSource int    `gdbc:"column=promo_source"`
	Currency    string `gdbc:"column=currency"`
	TierIndex   string `gdbc:"column=tier_index"`
	Status      int    `gdbc:"column=status"`
	Sku         string `gdbc:"column=sku"`
	ExtInfo     string `gdbc:"column=extinfo"`
	Ctime       int64  `gdbc:"column=ctime"`
	Mtime       int64  `gdbc:"column=mtime"`
}
