package service

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	marketplaceOrderAccountingIntegratedFeePb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/marketplace_order_accounting_integrated_fee.pb"
)

// OrderAccountIntegratedFeeService
// Note: the rate from order accounting side is inflated by 10^5,
// for the usages of these functions, sometimes need /10 * 10, sometimes need /10,
// this is due to listing side legacy logic,
// for the inflation between mtsku and mpsku is 10^4, and for sip, it is 10^5.
type OrderAccountIntegratedFeeService interface {
	// GetShopCommissionRateMap get commission rate map for shop list.
	// If failed to get for one shop id, then only record error in the response and continue to handle remaining part.
	GetShopCommissionRateMap(ctx context.Context, shopIdList []*model.ShopIdRegion) map[uint64]*ShopCommissionRateInfo

	// GetCommissionRate get commission rate map for single shop.
	// If failed, then return error
	GetCommissionRate(ctx context.Context, shopId uint64, region string) (uint64, error)

	GetReferenceServiceFeeRate(ctx context.Context, shopId uint64, region string) (uint64, error)

	GetShopFee(ctx context.Context, region string, shopId int64, feeType model.ShopFeeType) (float64, error)
}

type ShopCommissionRateInfo struct {
	Err            error
	CommissionRate uint64
}

type AppliableRuleDetailSlice []*marketplaceOrderAccountingIntegratedFeePb.AppliableRuleDetail

func (a AppliableRuleDetailSlice) Len() int {
	return len(a)
}
func (a AppliableRuleDetailSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a AppliableRuleDetailSlice) Less(i, j int) bool {
	if a[j].GetGroupId() < a[i].GetGroupId() {
		return true
	}
	if a[j].GetGroupId() == a[i].GetGroupId() {
		return a[j].GetRuleId() < a[i].GetRuleId()
	}
	return false
}
