package http

import (
	"context"
	"time"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/seller-server/seller-listing/service-kits/ssf/metadata"
	"git.garena.com/shopee/seller-server/seller-listing/service-kits/ssk/component/xhttp"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/httpcliutil"
)

func GetSubAccountHttp(ctx context.Context, httpClient httpcliutil.HTTPCli, url string, param map[string]string) (xhttp.HttpResponse, error) {
	subAccountServerConfig := config.GetHTTPApiConfig().SubAccountServer

	if subAccountServerConfig == nil {
		return nil, cerr.New("subAccountServerConfig not found", uint32(pb.Constant_ERROR_INTERNAL))
	}

	token := subAccountServerConfig.Token
	requestUrl := subAccountServerConfig.Host + url
	requestId := metadata.GetRequestId(ctx)
	headers := map[string]string{
		"Remote-Service": subAccountServerConfig.RemoteService,
		"Auth-Token":     token,
		"Api-Token":      "",
		"x_request_id":   requestId,
	}
	timeout := time.Duration(subAccountServerConfig.Timeout) * time.Second
	hTTPResponse, err := httpClient.Get(ctx, requestUrl, param, headers, timeout)
	return hTTPResponse, err
}
