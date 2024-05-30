package convutil

import (
	"fmt"
	"strings"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_merchant_constraints.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"

	"github.com/golang/protobuf/proto"
)

func Uint32sToInt64s(input []uint32) []int64 {
	out := make([]int64, 0, len(input))
	for _, i := range input {
		out = append(out, int64(i))
	}
	return out
}

func Int64sToUint32s(input []int64) []uint32 {
	out := make([]uint32, 0, len(input))
	for _, i := range input {
		out = append(out, uint32(i))
	}
	return out
}

func Uint64sToUint32s(input []uint64) []uint32 {
	out := make([]uint32, 0, len(input))
	for _, i := range input {
		out = append(out, uint32(i))
	}
	return out
}

// format float64 to string with precision without trailing zeros
// example:
// FloatFormat(0.123, 2) => 0.12
// FloatFormat(0.123, 4) => 0.123
// FloatFormat(12, 4) => 12
func FloatFormat(num float64, precision int) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.*f", precision, num), "0"), ".")
}

func Int64sToUint64s(input []int64) []uint64 {
	res := make([]uint64, 0, len(input))
	for _, v := range input {
		res = append(res, uint64(v))
	}
	return res
}

func Int64ToFloat64(input int64) float64 {
	return float64(input)
}

func ConvertMerchantConstraintListToProfitRateLimitList(beforeList []*internal.MerchantConstraints) []*price_sync_price_calculation.ProfitRateLimit {
	profitRateLimitList := make([]*price_sync_price_calculation.ProfitRateLimit, len(beforeList))
	for i, before := range beforeList {
		profitRateLimitList[i] = ConvertMerchantConstraintToProfitRateLimit(before)
	}
	return profitRateLimitList
}

func ConvertMerchantConstraintToProfitRateLimit(before *internal.MerchantConstraints) *price_sync_price_calculation.ProfitRateLimit {
	return &price_sync_price_calculation.ProfitRateLimit{
		Id:            before.Id,
		Region:        before.Region,
		ProfitRateMin: proto.Float64(Int64ToFloat64(before.GetProfitRateMin())),
		ProfitRateMax: proto.Float64(Int64ToFloat64(before.GetProfitRateMax())),
		Operator:      before.Operator,
		UpdateTime:    proto.Uint32(uint32(before.GetMtime())),
	}
}
