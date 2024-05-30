package calculate

import (
	"github.com/golang/protobuf/proto"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_business.pb"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

func GetAItemCustomizedOPLFromPItemOPLPrices(oplPrices []*price_business.ItemLevelPrice, purchaseLimit uint32) *pb.CustomizedOPL {
	var oplStartTime, oplEndTime, maxEndTime uint32
	var lastEndingPromotionId uint64

	// business requirement can refer to https://confluence.shopee.io/pages/viewpage.action?pageId=472626027 or
	//  https://confluence.shopee.io/display/SPPT/%5BSPPT-66631%5D%5BTD%5DOPL+%28OPL2%29+sync+decouple+from+SIP

	for _, price := range oplPrices {
		if oplStartTime == 0 {
			oplStartTime = price.GetStartTime()
		} else {
			if price.GetStartTime() > 0 && price.GetStartTime() < oplStartTime {
				oplStartTime = price.GetStartTime()
			}
		}
		if oplEndTime == 0 {
			oplEndTime = price.GetEndTime()
			lastEndingPromotionId = price.GetRuleId()
			maxEndTime = oplEndTime
		} else {
			if price.GetEndTime() > 0 {
				if price.GetEndTime() < oplEndTime {
					oplEndTime = price.GetEndTime()
				}
				if price.GetEndTime() > maxEndTime {
					maxEndTime = price.GetEndTime()
					lastEndingPromotionId = price.GetRuleId()
				}
			}
		}
	}
	return &pb.CustomizedOPL{
		StartTime:     proto.Uint32(oplStartTime),
		EndTime:       proto.Uint32(oplEndTime),
		PurchaseLimit: proto.Uint32(purchaseLimit),
		RepeatedTimes: proto.Uint32(uint32(len(oplPrices))),
		PromotionId:   proto.Uint64(lastEndingPromotionId),
	}
}
