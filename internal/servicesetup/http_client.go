package servicesetup

import (
	"fmt"
	"sync"

	"git.garena.com/shopee/seller-server/seller-listing/service-kits/ssk/component/xhttp"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/httpcliutil"
)

var (
	once       = sync.Once{}
	httpClient httpcliutil.HTTPCli
)

func InitializeHttpClient() {
	once.Do(func() {
		router, err := httpcliutil.NewHttpOverSpexRouter()
		if err != nil {
			panic(fmt.Sprintf("failed to initializeHttpClient, err=%s", err.Error()))
		}
		v := xhttp.NewDefaultHttpClientConfig()
		fastHttpClient := xhttp.NewClientV2(router, v...)
		httpClient = httpcliutil.NewHTTPCli(fastHttpClient)
	})
}

func GetHttpClient() httpcliutil.HTTPCli {
	once.Do(InitializeHttpClient)
	return httpClient
}
