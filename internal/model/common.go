package model

type ShopIdRegion struct {
	ShopId uint64
	Region string
}

func GroupShopIdRegionPairsByRegion(list []ShopIdRegion) map[string][]ShopIdRegion {
	res := make(map[string][]ShopIdRegion)
	for _, pair := range list {
		res[pair.Region] = append(res[pair.Region], pair)
	}
	return res
}

type ItemModelId struct {
	ItemId  uint64
	ModelId uint64
}

func BuildCurrencyPair(currency1, currency2 string) string {
	if currency1 < currency2 {
		return currency1 + "-" + currency2
	}
	return currency2 + "-" + currency1
}
