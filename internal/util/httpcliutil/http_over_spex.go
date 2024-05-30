package httpcliutil

import (
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/miscellaneousinner"
	"git.garena.com/shopee/seller-server/mktzlib/http_over_spex"
	"git.garena.com/shopee/seller-server/seller-listing/service-kits/ssk/component/xhttp"
)

func NewHttpOverSpexRouter() (xhttp.Router, error) {
	router := xhttp.NewRouter()

	sellerPlatformMiscellaneousinnerClient, err := miscellaneousinner.NewClient()
	if err != nil {
		return nil, err
	}
	err = router.RegisterMap(map[string]http_over_spex.Client{
		config.GetHTTPApiConfig().PlatformMiscellaneousinnerHspexRouterPath: sellerPlatformMiscellaneousinnerClient,
	})
	if err != nil {
		return nil, err
	}

	return router, nil
}
