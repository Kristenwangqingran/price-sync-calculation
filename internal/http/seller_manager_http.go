package http

import (
	"context"
	"strings"
	"time"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/httpcliutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/serverutil"
	"git.garena.com/shopee/seller-server/seller-listing/service-kits/ssk/component/xhttp"
)

func PostSellerManagerHttp(httpClient httpcliutil.HTTPCli, ctx context.Context, url string, body interface{}) (xhttp.HttpResponse, error) {
	en := strings.ToUpper(serverutil.GetEnv())
	if en == "DEV" {
		en = "TEST"
	}
	requestUrl := config.GetHTTPApiConfig().SellerManagerUrl + url
	param := map[string]string{}
	// TODO: for future, we need apply for our own token and use own service-name
	headers := map[string]string{
		"api-token":    config.GetHTTPApiConfig().SellerManagerToken,
		"service-name": "listing_service",
	}

	hTTPResponse, err := httpClient.PostJson(ctx, requestUrl, param, body, headers, 5*time.Second)
	if err != nil {
		logging.GetLogger(ctx).Error("request sellerManagerUrlMap err", ulog.Error(err))
		return nil, err
	}

	return hTTPResponse, nil
}
