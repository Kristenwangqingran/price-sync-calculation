package model

import (
	"encoding/json"
	"time"

	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
)

const (
	PriceSyncToggleReadFromSLS = int32(pb.Constant_PRICE_SYNC_TOGGLE_READ_FROM_SLS)
	PriceSyncToggleReadFromDB  = int32(pb.Constant_PRICE_SYNC_TOGGLE_READ_FROM_DB)
)

// CommonPriceConfig TODO migrate from sip-goservice, use same name
type CommonPriceConfig struct {
	Buffer            *float64 `json:"buffer,omitempty"`
	ExchangeRate      *float64 `json:"exchange_rate,omitempty"`
	InitHiddenPrice   *float64 `json:"init_hidden_price,omitempty"`
	HiddenPriceToggle *int32   `json:"hidden_price_toggle,omitempty"`
	ShippingFeeToggle *int32   `json:"shipping_fee_toggle,omitempty"`
}

type mstShopInnerFlag struct {
	SipRateControlledByOps int
	IsCnscShop             int
	CnscRelinkMerchant     int
}

var MstShopInnerFlag = &mstShopInnerFlag{
	// the first 3 bit was used for self manage
	SipRateControlledByOps: 1 << 3,
	IsCnscShop:             1 << 4,
	CnscRelinkMerchant:     1 << 5,
}

// TODO: can we remove unused fields, like start_task_id, stop_task_id?
type ExtPriceRate struct {
	StartTaskId int64              `json:"start_task_id"`
	StopTaskId  int64              `json:"stop_task_id"`
	IsCancel    bool               `json:"is_cancel"`
	StartTime   int64              `json:"start_time"`
	EndTime     int64              `json:"end_time"`
	SipRateMap  map[string]float64 `json:"sip_rate_map"`
}

func GetPShopOpsPriceRatioSettingFromMstShopRecord(s *sip_db.MstShop) (isCtlByOps bool, startTime, endTime int64) {
	if (s.InnerFlag & MstShopInnerFlag.SipRateControlledByOps) > 0 {
		isCtlByOps = true
	}

	extInfo := map[string]json.RawMessage{}
	rateinfo := new(ExtPriceRate)
	if len(s.Extinfo) > 0 {
		err := json.Unmarshal([]byte(s.Extinfo), &extInfo)
		if err != nil {
			return
		}
		raw, ok := extInfo["ops_price_rate"]
		if ok {
			err = json.Unmarshal(raw, rateinfo)
			if err != nil {
				return
			}
			if rateinfo.IsCancel {
				return
			}
			currTime := time.Now().Unix()
			if currTime > rateinfo.EndTime {
				return
			}

			startTime = rateinfo.StartTime
			endTime = rateinfo.EndTime
		}
	}
	return
}
