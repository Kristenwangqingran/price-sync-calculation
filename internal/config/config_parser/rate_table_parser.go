package config_parser

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

// parse utils
type rateTableParse struct{}

var RateTableParse = &rateTableParse{}
var Inf int64 = (1<<63 - 1) / 100000 * 100000
var InfStr = "inf"

func (*rateTableParse) FormatWeightRange(weightRange int64) string {
	if weightRange == Inf {
		return InfStr
	}
	return fmt.Sprintf("%v", calcutil.DbWeightToGram(int64(int(weightRange))))
}

func (*rateTableParse) ParseWeightRange(value string) (int64, error) {
	if strings.ToLower(value) == InfStr || value == "" {
		return Inf, nil
	}

	weightRange, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logging.GetLogger(context.Background()).Error(fmt.Sprintf("parse weight range fail, err=%s", err.Error()))
		return 0, cerr.New(fmt.Sprintf("failed to parse weight_range, err=%v", err), uint32(pb.Constant_ERROR_PARAMS))
	}

	if weightRange < 0 || weightRange > 9999 {
		return 0, cerr.New("weight range out of range:[0,9999]", uint32(pb.Constant_ERROR_PARAMS))
	}
	return int64(calcutil.GramToDbWeight(weightRange)), nil
}

func (*rateTableParse) ParseStartPrice(value string) (int64, error) {
	startPrice, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logging.GetLogger(context.Background()).Error(fmt.Sprintf("parse start_price fail, err=%s", err.Error()))

		return 0, cerr.New(fmt.Sprintf("failed to parse start_price, err=%v", err), uint32(pb.Constant_ERROR_PARAMS))
	}

	if startPrice < 0 || startPrice > 100000 {
		return 0, cerr.New("start price out of range:[0,100000]", uint32(pb.Constant_ERROR_PARAMS))
	}
	return calcutil.ToDBPrice(startPrice), nil
}

func (*rateTableParse) ParseStartWeight(value string) (int64, error) {
	startWeight, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logging.GetLogger(context.Background()).Error(fmt.Sprintf("parse start_weight fail, err=%s", err.Error()))
		return 0, cerr.New(fmt.Sprintf("failed to parse start_weight, err=%v", err), uint32(pb.Constant_ERROR_PARAMS))
	}
	if startWeight < 0 || startWeight > 9999 {
		return 0, cerr.New("start weight out of range:[0,9999]", uint32(pb.Constant_ERROR_PARAMS))
	}
	return int64(calcutil.GramToDbWeight(startWeight)), nil
}

func (*rateTableParse) ParseRoundSize(value string) (int64, error) {
	roundSize, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		logging.GetLogger(context.Background()).Error(fmt.Sprintf("parse round_size fail, err=%s", err.Error()))
		return 0, cerr.New(fmt.Sprintf("failed to parse round_size, err=%v", err), uint32(pb.Constant_ERROR_PARAMS))
	}
	if roundSize < -99 || roundSize > 99 {
		return 0, cerr.New("round place out of range:[-99,99]", uint32(pb.Constant_ERROR_PARAMS))
	}
	return roundSize, nil
}

func (*rateTableParse) ParsePrice(value string) (int64, error) {
	price, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logging.GetLogger(context.Background()).Error(fmt.Sprintf("parse price fail, err=%s", err.Error()))
		return 0, cerr.New(fmt.Sprintf("failed to parse price, err=%v", err), uint32(pb.Constant_ERROR_PARAMS))
	}
	if price < 0 || price > 9999.9999 {
		return 0, cerr.New("price out of range:[0-9999.9999]", uint32(pb.Constant_ERROR_PARAMS))
	}
	return calcutil.ToDBPrice(price), nil
}

func (*rateTableParse) ParseWeightStep(value string) (int64, error) {
	weightStep, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logging.GetLogger(context.Background()).Error(fmt.Sprintf("parse weight step fail, err=%s", err.Error()))
		return 0, cerr.New(fmt.Sprintf("failed to parse weight_step, err=%v", err), uint32(pb.Constant_ERROR_PARAMS))
	}
	if weightStep <= 0 || weightStep > 9999 {
		return 0, cerr.New("round place out of range:(0,9999]", uint32(pb.Constant_ERROR_PARAMS))
	}
	return int64(calcutil.GramToDbWeight(weightStep)), nil
}

func (*rateTableParse) ParseAdjustment(value string) (int64, error) {
	adjustment, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logging.GetLogger(context.Background()).Error(fmt.Sprintf("parse adjustment fail, err=%s", err.Error()))
		return 0, cerr.New(fmt.Sprintf("failed to parse adjustment, err=%v", err), uint32(pb.Constant_ERROR_PARAMS))
	}
	if adjustment < -99999 || adjustment > 99999 {
		return 0, cerr.New("adjustment out of range:[-99999,99999]", uint32(pb.Constant_ERROR_PARAMS))
	}
	return calcutil.ToDBPrice(adjustment), nil
}

func buildHpfnConfigMap(confs []map[string]string) (map[string][]*sip_db.HpfnConfig, bool) {
	regions := []string{"PH", "TH", "SG", "MY", "VN", "ID", "TW", "BR"}
	for _, region := range regions {
		find := false
		for _, conf := range confs {
			c := conf["country"]
			if c == region {
				find = true
			}
		}
		if !find {
			logging.GetLogger(context.Background()).Error(fmt.Sprintf("region=%s not config", region))
			return nil, false
		}
	}

	ret := map[string][]*sip_db.HpfnConfig{}
	for _, conf := range confs {
		region := conf["country"]

		weightRange, err := RateTableParse.ParseWeightRange(conf["weight_range"])
		if err != nil {
			logging.GetLogger(context.Background()).Error(err.Error())
			return nil, false
		}
		startPrice, err := RateTableParse.ParseStartPrice(conf["start_price"])
		if err != nil {
			logging.GetLogger(context.Background()).Error(err.Error())
			return nil, false
		}
		startWeight, err := RateTableParse.ParseStartWeight(conf["start_weight"])
		if err != nil {
			logging.GetLogger(context.Background()).Error(err.Error())
			return nil, false
		}
		roundSize, err := RateTableParse.ParseRoundSize(conf["round_size"])
		if err != nil {
			logging.GetLogger(context.Background()).Error(err.Error())
			return nil, false
		}
		price, err := RateTableParse.ParsePrice(conf["price"])
		if err != nil {
			logging.GetLogger(context.Background()).Error(err.Error())
			return nil, false
		}
		weightStep, err := RateTableParse.ParseWeightStep(conf["weight_step"])
		if err != nil {
			logging.GetLogger(context.Background()).Error(err.Error())
			return nil, false
		}
		adjustment, err := RateTableParse.ParseAdjustment(conf["adjustment"])
		if err != nil {
			logging.GetLogger(context.Background()).Error(err.Error())
			return nil, false
		}

		rateTable := &sip_db.HpfnConfig{
			WeightRange: weightRange,
			StartPrice:  startPrice,
			StartWeight: startWeight,
			RoundSize:   roundSize,
			Price:       price,
			WeightStep:  weightStep,
			Adjustment:  adjustment,
		}
		oldTable := ret[region]
		ret[region] = append(oldTable, rateTable)
	}
	newRet := map[string][]*sip_db.HpfnConfig{}
	for pRegion, xs := range ret {
		sort.Slice(xs, func(i, j int) bool {
			return xs[i].WeightRange < xs[j].WeightRange
		})
		newRet[pRegion] = xs
	}
	return newRet, true
}

func GetHpfnConfigMapFromConfig() (map[string][]*sip_db.HpfnConfig, error) {
	cfg, err := config.GetDefaultRateTableForCbHiddenPrice()
	if err != nil {
		return nil, err
	}
	res, ok := buildHpfnConfigMap(cfg)
	if !ok {
		return nil, cerr.New(fmt.Sprintf("failed to build hpfn config map, cfg=%v", cfg), uint32(pb.Constant_ERROR_PARAMS))
	}
	return res, nil
}
