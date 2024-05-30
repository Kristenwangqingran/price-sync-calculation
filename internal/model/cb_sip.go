package model

import (
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
)

type CbSipCalculateAPriceByPItemRequest struct {
	MerchantId     uint64
	MerchantRegion string
	PShopId        uint64
	PRegion        string
	PItemId        uint64

	AShopId            uint64
	ARegion            string
	AItemId            uint64
	Queries            []AItemCbSipQueryId
	CalculateForCreate bool
}

type CbSipCalculateAOPLByPItemRequest struct {
	PRegion string
	PItemId uint64

	AShopId uint64
	ARegion string
}

type AItemCbSipQueryId struct {
	AModelId        uint64 `json:"a_model_id"`
	PItemPrice      int64  `json:"p_item_price"`
	PNormalPrice    int64  `json:"p_normal_price"`
	PPromotionPrice *int64 `json:"p_promotion_price"`
}

type CbSipCalculateAPriceByPItemResult struct {
	ANormalPrice             int64
	APromotionPrice          int64
	ASettlementPrice         int64
	ASettlementPriceCurrency string
	Snap                     *pb.CbSipPriceFactorSnap
}

type CbSipCalculateAOPLByPItemResult struct {
	Opl *pb.CustomizedOPL
}

// reference:
// https://docs.google.com/spreadsheets/d/1cDRXoF48m2IWu-8oIZiQmti2gpRYoSJ8O8_kYXQjQ7E/edit#gid=234202052
// https://docs.google.com/spreadsheets/d/1BfWuV64-JLVUDm9lLaHUNS6vo_AfbxjnXdUAnU1WxNM/edit?pli=1#gid=371273628  2020-05-07
// https://confluence.shopee.io/display/SCPM/%5BCB+SIP%5DRate+table+logic+improvement 2021-10-20

type hpfnConfigLevel int

const (
	HpfnNoneCfg         hpfnConfigLevel = 0
	HpfnItemLevelCfg    hpfnConfigLevel = 1
	HpfnShopLevelCfg    hpfnConfigLevel = 2
	HpfnRegionLevelCfg  hpfnConfigLevel = 3
	HpfnDefaultLevelCfg hpfnConfigLevel = 4
)

type HiddenPriceConf struct {
	StartPrice  int64
	StartWeight int64
	RoundSize   int64
	Price       int64
	WeightStep  int64
	Adjustment  int64
	// debug info
	HpfnKey      string
	HpfnCfgLevel hpfnConfigLevel
}

type RateTableCfg struct {
	MstRateKey  string            `json:"mst_rate_key"`
	AffiRateKey map[string]string `json:"affi_rate_key"`
}

type ModelPriceInfo struct {
	NormalPrice    int64
	ItemPrice      int64
	SellerDiscount int64
	SellingPrice   int64

	ItemPriceCurrency string
	// if true, and shopID is in whitelist, then don't calc & set item_price
	StopSipPriceAutoSync bool
}

type ExchangeRateCacheData struct {
	SourceCurrency string `json:"source_currency"`
	TargetCurrency string `json:"target_currency"`
	ExchangeRate   string `json:"exchange_rate"`
}

type CbSipCalculateSipItemPriceRequest struct {
	ShopId uint64
	Region string
	ItemId uint64

	// used to calc hidden fee
	ChannelIdList  []uint32
	LeafCategoryId uint64
	Weight         uint64

	Queries []CbSipCalculateSipItemPriceSingleQuery
}

type CbSipCalculateSipItemPriceSingleQuery struct {
	ModelId uint64
	Price   int64
}

type CbSipCalculateSipItemPriceResult struct {
	ModelId        uint64
	CbSipItemPrice int64
	Currency       string
}

type CbSipGetRegionLevelConfigRequest struct {
	InfoType uint32
	Queries  []RegionPair
}

type RegionPair struct {
	SrcRegion string
	DstRegion string
}

type CbSipGetRegionLevelConfigResult struct {
	ExchangeRateList  []ExchangeRate
	CountryMarginList []CountryMargin
}

type ExchangeRate struct {
	SrcCurrency  string
	DstCurrency  string
	ExchangeRate string
}

type CountryMargin struct {
	SrcRegion     string
	DstRegion     string
	CountryMargin float64
}

type CbSipGetShopLevelConfigRequest struct {
	PShopId uint64
}

type CbSipGetShopLevelConfigResult struct {
	AShopConfigList []AShopConfigInfo
}

type AShopConfigInfo struct {
	AShopId uint64
	SipRate float64
}

type CbSipRateConfigResult struct {
	DefaultSipRateStr string
	SipRateLimitStr   string
}

type CbSipAHiddenFeeInfoType = uint32

const (
	RulesListWithPagination = CbSipAHiddenFeeInfoType(pb.Constant_RULES_HPFN_CONFIG_LIST_WITH_PAGINATION)
	RulesListAll            = CbSipAHiddenFeeInfoType(pb.Constant_RULES_HPFN_CONFIG_LIST_ALL)
	RuleDetail              = CbSipAHiddenFeeInfoType(pb.Constant_RULE_HPFN_CONFIG_DETAIL)
	RuleRegionSetting       = CbSipAHiddenFeeInfoType(pb.Constant_RULE_HPFN_REGION_SETTING)
)

type CbSipGetAHiddenPriceConfigRequest struct {
	InfoType  CbSipAHiddenFeeInfoType // refer to CbSipAHiddenFeeInfoType
	PageIndex uint32                  // for RulesListWithPagination
	PageSize  uint32                  // for RulesListWithPagination
	RuleKey   string                  // for RuleDetail
}

type CbSipGetAHiddenPriceConfigResult struct {
	Total                 uint32 // for RulesListWithPagination & RulesListAll
	RuleRegionSettingsStr string // for RuleRegionSetting
	Rules                 []AHiddenFeeRuleInfo
}

type AHiddenFeeRuleInfo struct {
	RuleKey  string
	DescInfo string
	Details  []AHiddenFeeRuleRow
}

type AHiddenFeeRuleRow struct {
	WeightRange int64
	StartPrice  int64
	StartWeight int64
	RoundSize   int64
	Price       int64
	WeightStep  int64
	Adjustment  int64
	DescInfo    string
}
