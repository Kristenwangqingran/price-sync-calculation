package model

type MtskuMpskuPriceQuery struct {
	// required
	SourcePrice       int64
	MpskuShopId       uint64
	MpskuRegion       string
	MpskuItemId       uint64
	Weight            uint64
	LeafCategoryId    uint64
	EnabledChannelIds []uint32
}

type MtskuMpskuPriceCalcResult struct {
	Err                error
	DstPrice           int64
	HidePrice          int64
	HidePriceErrorCode int32
}

type SetCbscPriceFactorQuery struct {
	MerchantId   uint64
	ShopSettings []ShopCbscPriceFactorSetting
}

type ShopCbscPriceFactorSetting struct {
	ShopId         uint64
	Region         string
	ProfitRate     *uint64
	ServiceFeeRate *uint64
}

type CbscPriceFactorLimit struct {
	MinProfitRate     int64
	MaxProfitRate     int64
	MinServiceFeeRate int64
	MaxServiceFeeRate int64
}

type GetHidePriceForCbscRequest struct {
	QueryId int

	Region         string
	Weight         uint64
	IsMtskuToMpsku bool

	// for SLS query
	ShopId               uint64
	ItemId               uint64
	LeafCategoryId       uint64
	EnabledChannelIdList []uint32
	IgnoreChannelErr     bool
}

type GetHidePriceForCbscResult struct {
	QueryId   int
	Err       error
	HidePrice float64
}

type GetCommissionRateRequest struct {
	ShopId      uint64
	MpskuRegion string
}

type GetCommissionRateResult struct {
	Err            error
	CommissionRate uint64
}

func GroupGetCommissionRateReqByShopId(queries []GetCommissionRateRequest) map[uint64][]GetCommissionRateRequest {
	res := make(map[uint64][]GetCommissionRateRequest)

	for _, query := range queries {
		res[query.ShopId] = append(res[query.ShopId], query)
	}

	return res
}

func FilterGetCommissionRateReqByShopId(queries []GetCommissionRateRequest) []GetCommissionRateRequest {
	queriesGroupByShopId := GroupGetCommissionRateReqByShopId(queries)
	res := make([]GetCommissionRateRequest, 0)
	for _, reqs := range queriesGroupByShopId {
		for _, req := range reqs {
			res = append(res, req)
		}
	}

	return res
}

type GetCbscPriceRateRequest struct {
	ShopId            uint64
	CommissionRate    uint64
	CommissionRateErr error
}

type GetCbscPriceRateResult struct {
	Err           error
	CbscPriceRate float64
}

type GetCbscProfitRateRequest struct {
	ShopId uint64
}

type GetCbscProfitRateResult struct {
	Err        error
	ProfitRate float64
}
