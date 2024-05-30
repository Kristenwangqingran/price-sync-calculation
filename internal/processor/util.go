package processor

import (
	"git.garena.com/shopee/core-server/core-logic/cerr"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	spcommon "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

func GetErrorCode(err error) uint32 {
	if err == nil {
		return 0
	}

	code := cerr.Code(err)
	if code == uint32(spcommon.Constant_ERROR_UNKNOWN) {
		return uint32(pb.Constant_ERROR_INTERNAL)
	}
	return code
}
