package model

// BatchCalcHiddenFeeRequest defines request for calculating hidden fee by SLS
//
//	refer to API doc: http://apidoc.i.ssc.shopeemobile.com/project/514/interface/api/133179
//	(ask @ruofei.miu for yapi access)
type BatchCalcHiddenFeeRequest struct {
	Token     string           `json:"token,omitempty"`
	RequestID string           `json:"request_id,omitempty"`
	Timestamp int64            `json:"timestamp,omitempty"`
	List      []*CalcHiddenFee `json:"list"`
}

func (b *BatchCalcHiddenFeeRequest) SetTimestamp(timestamp int64) {
	b.Timestamp = timestamp
}

func (b *BatchCalcHiddenFeeRequest) SetToken(token string) {
	b.Token = token
}

func (b *BatchCalcHiddenFeeRequest) SetRequestID(requestID string) {
	b.RequestID = requestID
}

type CalcHiddenFee struct {
	ProductIDs      []int64    `json:"product_ids,omitempty"`
	Region          string     `json:"region,omitempty"`
	ShopID          uint64     `json:"shop_id,omitempty"`
	ShopGroup       []int32    `json:"shop_group,omitempty"`
	SkuInfos        []*SkuInfo `json:"sku_infos,omitempty"`
	Direction       int8       `json:"direction,omitempty"`
	AddCodFee       int32      `json:"add_cod_fee,omitempty"`
	WmsFlag         int32      `json:"wms_flag,omitempty"`
	DgFlag          int32      `json:"dg_flag,omitempty"`
	CogAmount       float64    `json:"cog_amount,omitempty"`
	Cogs            float64    `json:"cogs,omitempty"`
	BusinessModel   int32      `json:"business_model,omitempty"`
	PickupTag       int8       `json:"pickup_tag,omitempty"`
	FallbackFlag    int32      `json:"fallback_flag,omitempty"`
	PickupLocation  *Location  `json:"pickup_location,omitempty"`
	DeliverLocation *Location  `json:"deliver_location,omitempty"`
}

func GroupHiddenFeeQueriesByRegion(queries []*CalcHiddenFee) (result map[string][]*CalcHiddenFee, rawIndexes map[string][]int) {
	result = make(map[string][]*CalcHiddenFee)
	rawIndexes = make(map[string][]int)
	for i, query := range queries {
		result[query.Region] = append(result[query.Region], query)
		rawIndexes[query.Region] = append(rawIndexes[query.Region], i)
	}
	return result, rawIndexes
}

type SkuInfo struct {
	ItemID       uint64  `json:"item_id,omitempty"`
	CategoryID   uint64  `json:"category_id,omitempty"`
	Weight       float64 `json:"weight,omitempty"`
	Quantity     uint32  `json:"quantity,omitempty"`
	Length       float64 `json:"length,omitempty"`
	Width        float64 `json:"width,omitempty"`
	Height       float64 `json:"height,omitempty"`
	ItemPriceUSD float64 `json:"item_price_usd,omitempty"`
	ItemPrice    float64 `json:"item_price,omitempty"`
	ShopID       string  `json:"shop_id,omitempty"`
	ItemSizeID   int32   `json:"item_size_id,omitempty"`
}

type BatchCalcHiddenFeeResponse struct {
	RetCode int64                    `json:"retcode,omitempty"`
	Message string                   `json:"message,omitempty"`
	Data    [][]*CalcHiddenFeeResult `json:"data,omitempty"`
	Detail  string                   `json:"detail,omitempty"`
}

type CalcHiddenFeeResult struct {
	ProductID         int64   `json:"product_id,omitempty"`
	ESF               float64 `json:"esf,omitempty"`
	ASF               float64 `json:"asf,omitempty"`
	ESFRateID         int64   `json:"esf_rate_id,omitempty"`
	ASFRateID         int64   `json:"asf_rate_id,omitempty"`
	HiddenFee         float64 `json:"hidden_fee,omitempty"`
	AfterTaxHiddenFee float64 `json:"after_tax_hidden_fee,omitempty"`
	AfterTaxESF       float64 `json:"after_tax_esf,omitempty"`
	AfterTaxASF       float64 `json:"after_tax_asf,omitempty"`
	RetCode           int64   `json:"retcode,omitempty"`
	Message           string  `json:"message,omitempty"`
}

type Location struct {
	LocationIds []uint64 `json:"location_ids,omitempty"`
	PostCode    string   `json:"post_code,omitempty"`
	Longitude   string   `json:"longitude,omitempty"`
	Latitude    string   `json:"latitude,omitempty"`
}

type SlsHiddenPriceQuery struct {
	PItemEnabledChannelIds []uint32 `json:"p_item_enabled_channel_ids,omitempty"`
	WeightInGram           float64  `json:"weight_in_gram,omitempty"`
	DeliveryLocationIds    []uint64 `json:"delivery_location_ids,omitempty"`
}

type SlsHiddenPriceQueryKey struct {
	WeightInDB          int64  `json:"weight_in_db,omitempty"`
	DeliveryLocationStr string `json:"delivery_location_str,omitempty"`
}

type SlsHiddenPriceResult struct {
	Err         error
	HiddenPrice float64
}

type ChannelInfo struct {
	ChannelId uint64 `json:"channel_id,omitempty"`
	Tag       int32  `json:"tag,omitempty"`
}
