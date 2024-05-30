package testconfig

import (
	"encoding/json"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
)

func BuildSIPConfigFromLiveCfg() *config.SIPMigrationConfig {
	currencyCommonConfStr := `
{
  "BRL": {
    "exchange_rate_max_limit": 12,
    "exchange_rate_min_limit": 1,
    "precision": -2,
    "precision_for_fee": -2,
    "use_special_roundup": true
  },
  "CLP": {
    "exchange_rate_max_limit": 1310,
    "exchange_rate_min_limit": 145,
    "precision": 3,
    "precision_for_fee": 1
  },
  "CNY": {
    "exchange_rate_max_limit": 12.5,
    "exchange_rate_min_limit": 2.5,
    "precision": -2,
    "precision_for_fee": -2
  },
  "COP": {
    "exchange_rate_max_limit": 6626,
    "exchange_rate_min_limit": 736,
    "precision": 1,
    "precision_for_fee": 1
  },
  "EUR": {
    "exchange_rate_max_limit": 1.5,
    "exchange_rate_min_limit": 0.4,
    "precision": -2,
    "precision_for_fee": -2
  },
  "IDR": {
    "exchange_rate_max_limit": 21539,
    "exchange_rate_min_limit": 7179,
    "precision": 2,
    "precision_for_fee": 1
  },
  "INR": {
    "exchange_rate_max_limit": 100,
    "exchange_rate_min_limit": 50,
    "precision": 0,
    "precision_for_fee": 0
  },
  "MXN": {
    "exchange_rate_max_limit": 38,
    "exchange_rate_min_limit": 4,
    "precision": -2,
    "precision_for_fee": -2,
    "use_special_roundup": true
  },
  "MYR": {
    "exchange_rate_max_limit": 7,
    "exchange_rate_min_limit": 2,
    "precision": -2,
    "precision_for_fee": -2,
    "use_special_roundup": true
  },
  "PHP": {
    "exchange_rate_max_limit": 73,
    "exchange_rate_min_limit": 24,
    "precision": 0,
    "precision_for_fee": -1
  },
  "PLN": {
    "exchange_rate_max_limit": 4.5,
    "exchange_rate_min_limit": 2,
    "precision": -2,
    "precision_for_fee": -2
  },
  "SGD": {
    "exchange_rate_max_limit": 2.5,
    "exchange_rate_min_limit": 0.5,
    "precision": -2,
    "precision_for_fee": -2,
    "use_special_roundup": true
  },
  "THB": {
    "exchange_rate_max_limit": 46,
    "exchange_rate_min_limit": 15,
    "precision": 0,
    "precision_for_fee": -2
  },
  "TWD": {
    "exchange_rate_max_limit": 42,
    "exchange_rate_min_limit": 13,
    "precision": 0,
    "precision_for_fee": -2
  },
  "USD": {
    "exchange_rate_max_limit": 1,
    "exchange_rate_min_limit": 1,
    "precision": -2,
    "precision_for_fee": -2
  },
  "VND": {
    "exchange_rate_max_limit": 34545,
    "exchange_rate_min_limit": 11515,
    "precision": 3,
    "precision_for_fee": 2
  }
}
`
	currencyCommonConf := config.CurrencyCommonConf{}
	if err := json.Unmarshal([]byte(currencyCommonConfStr), &currencyCommonConf.CommonSetting); err != nil {
		panic(err)
	}

	return &config.SIPMigrationConfig{
		RegionCommonConf: config.RegionCommonConf{
			CommonSetting: map[string]config.RegionCommonSetting{
				"BR": {
					Currency: "BRL",
				},
				"CL": {
					Currency: "CLP",
				},
				"CN": {
					Currency: "CNY",
				},
				"CO": {
					Currency: "COP",
				},
				"ES": {
					Currency: "EUR",
				},
				"FR": {
					Currency: "EUR",
				},
				"ID": {
					Currency: "IDR",
				},
				"IN": {
					Currency: "INR",
				},
				"KR": {},
				"MX": {
					Currency: "MXN",
				},
				"MY": {
					Currency: "MYR",
				},
				"PH": {
					Currency: "PHP",
				},
				"PL": {
					Currency: "PLN",
				},
				"SG": {
					Currency: "SGD",
				},
				"TH": {
					Currency: "THB",
				},
				"TW": {
					Currency: "TWD",
				},
				"VN": {
					Currency: "VND",
				},
			},
		},
		CurrencyCommonConf: currencyCommonConf,
	}
}
