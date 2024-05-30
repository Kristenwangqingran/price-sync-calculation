package model

import (
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

type LocalSipCalculateAPriceQuery struct {
	QueryId int

	AShopId          uint64
	ARegion          string
	AItemId          uint64 // required for non-create scenario
	AModelId         uint64
	PNormalPrice     int64
	PPromotionPrices []int64

	// required for creation scenario
	LeafCategoryId       uint64
	EnabledChannelIdList []int64
}

func PickUniqARegionFromLocalSipCalculateAPriceQueries(queries []LocalSipCalculateAPriceQuery) []string {
	uniqMap := map[string]struct{}{}
	for _, query := range queries {
		uniqMap[query.ARegion] = struct{}{}
	}

	res := make([]string, 0)
	for aRegion := range uniqMap {
		res = append(res, aRegion)
	}
	return res
}

func PickUniqAShopIdFromLocalSipCalculateAPriceQueries(queries []LocalSipCalculateAPriceQuery) []uint64 {
	uniqMap := map[uint64]struct{}{}
	for _, query := range queries {
		uniqMap[query.AShopId] = struct{}{}
	}

	res := make([]uint64, 0)
	for aShopId := range uniqMap {
		res = append(res, aShopId)
	}
	return res
}

func PickUniqAItemIdFromLocalSipCalculateAPriceQueries(queries []LocalSipCalculateAPriceQuery) []uint64 {
	uniqMap := map[uint64]struct{}{}
	for _, query := range queries {
		uniqMap[query.AItemId] = struct{}{}
	}

	res := make([]uint64, 0)
	for aItemId := range uniqMap {
		res = append(res, aItemId)
	}
	return res
}

type LocalSipCalculateAPriceResult struct {
	Err             error
	NormalPrice     int64
	PromotionPrices []int64
	AShopId         uint64
	ARegion         string
	AItemId         uint64
	AModelId        uint64
	PriceCalSnap    *pb.LocalSipPriceFactorSnap
}

type LocalSipHiddenPriceQuery struct {
	QueryId      int                `json:"query_id,omitempty"`
	CommonConfig *CommonPriceConfig `json:"common_config,omitempty"`
	AShopId      uint64             `json:"a_shop_id,omitempty"`
	ARegion      string             `json:"a_region,omitempty"`
	Weight       int64              `json:"weight,omitempty"`
	PRegion      string             `json:"p_region"`
}

func GroupLocalSipHiddenPriceQueryBySlsToggleStatus(queries []LocalSipHiddenPriceQuery) (slsQueries []LocalSipHiddenPriceQuery, dbQueries []LocalSipHiddenPriceQuery) {
	for _, query := range queries {
		if query.CommonConfig != nil && query.CommonConfig.HiddenPriceToggle != nil &&
			*query.CommonConfig.HiddenPriceToggle == PriceSyncToggleReadFromSLS &&
			config.UseSlsHiddenPriceOnSlsMode(query.AShopId, query.PRegion, query.ARegion) {
			slsQueries = append(slsQueries, query)
		} else {
			dbQueries = append(dbQueries, query)
		}
	}
	return slsQueries, dbQueries
}

type LocalSipHiddenPriceResult struct {
	QueryId int

	Err         error
	HiddenPrice float64
}

type LocalSipShippingFeeQuery struct {
	QueryId int

	CommonConfig *CommonPriceConfig

	AShopId  uint64
	AItemId  uint64 // required for non-create scenario
	AModelId uint64
	ARegion  string

	// required for creation scenario && calculate on sls mode
	LeafCategoryId       uint64
	EnabledChannelIdList []int64

	Weight int64
}

func PickAItemIdsFromLocalShippingShippingFeeQuery(queries []LocalSipShippingFeeQuery) []uint64 {
	res := make([]uint64, 0)
	for _, query := range queries {
		res = append(res, query.AItemId)
	}
	return res
}

func GroupLocalSipShippingFeeQueryByARegion(queries []LocalSipShippingFeeQuery) map[string][]LocalSipShippingFeeQuery {
	res := make(map[string][]LocalSipShippingFeeQuery)
	for _, query := range queries {
		res[query.ARegion] = append(res[query.ARegion], query)
	}
	return res
}

func GroupLocalSipShippingFeeQueryBySlsToggle(queries []LocalSipShippingFeeQuery) (slsQueries []LocalSipShippingFeeQuery, dbQueries []LocalSipShippingFeeQuery) {
	for _, query := range queries {
		if query.CommonConfig != nil && query.CommonConfig.ShippingFeeToggle != nil && *query.CommonConfig.ShippingFeeToggle == PriceSyncToggleReadFromSLS {
			slsQueries = append(slsQueries, query)
		} else {
			dbQueries = append(dbQueries, query)
		}
	}
	return slsQueries, dbQueries
}

type LocalSipShippingFeeResult struct {
	QueryId int

	Err         error
	ShippingFee float64
}

type LocalSipPriceFactorInfoType int32

const (
	BasicInfo       = LocalSipPriceFactorInfoType(pb.Constant_BASIC_INFO)
	HiddenFeeInfo   = LocalSipPriceFactorInfoType(pb.Constant_HIDDEN_FEE)
	ShippingFeeInfo = LocalSipPriceFactorInfoType(pb.Constant_SHIPPING_FEE)
)

type GetLocalSipPriceFactorQuery struct {
	QueryId int

	PRegion string
	ARegion string
}

type LocalSipPriceFactorInfo struct {
	BasicInfo            *LocalSipFactorBasicInfo
	LocalHiddenFeeInfo   []*LocalSipFactorHiddenFeeInfo
	LocalShippingFeeInfo []*LocalSipFactorShippingFeeInfo
}

type LocalSipFactorBasicInfo struct {
	CountryMargin          *float64
	MinCountryMargin       float64
	MaxCountryMargin       float64
	ExchangeRate           *float64
	MinExchangeRate        float64
	MaxExchangeRate        float64
	InitHiddenPrice        *float64
	MinInitHiddenPrice     float64
	MaxInitHiddenPrice     float64
	InitialHiddenFeeToggle *int32
	ShippingFeeToggle      *int32
}

type LocalSipFactorShippingFeeInfo struct {
	Id               int64
	Ctime            int64
	PRegion          string
	ARegion          string
	Weight           int64
	ShippingFeePrice int64
}

type LocalSipFactorHiddenFeeInfo struct {
	Id          int64
	Ctime       int64
	PRegion     string
	ARegion     string
	Weight      int64
	HiddenPrice int64
}
