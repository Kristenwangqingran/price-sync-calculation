package model

import (
	marketplaceOrderAccountingIntegratedFeePb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/marketplace_order_accounting_integrated_fee.pb"
)

type AShopData struct {
	AffiShopId int64

	ShopMargin int64

	PriceRatio  int
	PromotionId int64
}

type ShopDetail struct {
	ShopId int64

	PickupAddressId int32

	UserId      int64
	Name        string
	Region      string
	Status      int
	Ctime       int64
	Mtime       int64
	CbOption    int
	Cover       string
	Description string
	IsBbcSeller bool

	// extinfo
	IsSipPrimary      bool
	IsSipAffiliated   bool
	IsSipCb           bool
	Covers            []string
	UpdateShopCovers  bool
	ReturnAddressId   int32
	CbReturnAddressId int32
	SipPrimaryRegion  string
}

type ShopFeeType int

const (
	ShopFeeTypeServiceFee    ShopFeeType = ShopFeeType(marketplaceOrderAccountingIntegratedFeePb.Constant_FEE_TYPE_SERVICE_FEE)
	ShopFeeTypeCommissionFee ShopFeeType = ShopFeeType(marketplaceOrderAccountingIntegratedFeePb.Constant_FEE_TYPE_COMMISSION_FEE)
)
