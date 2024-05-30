package http

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/httpcliutil"
	"git.garena.com/shopee/seller-server/seller-listing/service-kits/ssk/component/xhttp"
)

func GetSellerAdminHttp(httpClient httpcliutil.HTTPCli, ctx context.Context, regionSuffix string, url string, param map[string]string) (xhttp.HttpResponse, error) {
	requestUrl := config.GetHTTPApiConfig().SellerAdminUrl + url
	if regionSuffix != "" {
		requestUrl = fmt.Sprintf(requestUrl, regionSuffix)
	}

	nonce := hex.EncodeToString(uuid.NewV4().Bytes())
	md := md5.New()
	md.Write([]byte(config.GetHTTPApiConfig().SellerAdminKey + nonce))
	sign := hex.EncodeToString(md.Sum(nil))
	headers := map[string]string{
		"X-Nonce": nonce,
		"X-Sign":  sign,
		"X-Appid": config.GetHTTPApiConfig().SellerAdminAppId,
	}
	timeout := time.Duration(config.GetHTTPApiConfig().SellerAdminTimeoutMs) * time.Millisecond

	hTTPResponse, err := httpClient.Get(ctx, requestUrl, param, headers, timeout)
	if err != nil {
		return nil, err
	}
	return hTTPResponse, nil
}
