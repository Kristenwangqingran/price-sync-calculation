package http

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/constant"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"

	uuid "github.com/satori/go.uuid"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	priceSyncPriceCalculationPb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/httpcliutil"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	"git.garena.com/shopee/seller-server/seller-listing/service-kits/ssf/metadata"
	"git.garena.com/shopee/seller-server/seller-listing/service-kits/ssk/component/xhttp"
)

type SLSBasicRequest interface {
	SetTimestamp(timestamp int64)
	SetToken(token string)
	SetRequestID(requestID string)
}

func GetFulfillmentHttp(httpClient httpcliutil.HTTPCli, ctx context.Context, url string, headers map[string]string, body map[string]string) (xhttp.HttpResponse, error) {
	httpApiConf := config.GetHTTPApiConfig()
	requestUrl := httpApiConf.SlsFulfillmentUrl + url
	headers["client-key"] = httpApiConf.SlsFulfillmentToken
	headers["client"] = httpApiConf.SlsFulfillmentClient
	timeout := time.Duration(httpApiConf.SlsRetryTimeoutMs) * time.Millisecond

	hTTPResponse, err := httpClient.Get(ctx, requestUrl, body, headers, timeout)
	if hTTPResponse == nil || !hTTPResponse.IsOk() || err != nil {
		errMsg := "failed in GetFulfillmentHttp"
		if err != nil {
			errMsg = fmt.Sprintf("%s, err=%s", errMsg, err.Error())
		}
		logging.GetLogger(ctx).Error(errMsg)
		return nil, cerr.New(errMsg, uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
	}
	return hTTPResponse, nil
}

func PostSlsOldUrlHttp(httpClient httpcliutil.HTTPCli, ctx context.Context, regionSuffix string, url string, body map[string]interface{}, timeoutMs int32) (xhttp.HttpResponse, error) {
	requestUrl := config.GetHTTPApiConfig().SlsOldUrlFormat + url
	if regionSuffix != "" {
		requestUrl = fmt.Sprintf(requestUrl, regionSuffix)
	}
	headers := map[string]string{"token": config.GetHTTPApiConfig().SlsOldToken}
	body["token"] = config.GetHTTPApiConfig().SlsOldToken
	param := map[string]string{}

	return postSlsLogisticHttp(httpClient, ctx, requestUrl, body, param, headers, timeoutMs)
}

func PostSlsLpsHttp(httpClient httpcliutil.HTTPCli, ctx context.Context, regionSuffix string, url string, body interface{}, timeoutMs int32) (xhttp.HttpResponse, error) {
	requestUrl := config.GetHTTPApiConfig().SlsLpsUrlFormat + url
	if regionSuffix != "" {
		requestUrl = fmt.Sprintf(requestUrl, regionSuffix)
	}
	headers := map[string]string{"token": config.GetHTTPApiConfig().SlsLpsToken}
	param := map[string]string{}

	return postSlsLogisticHttp(httpClient, ctx, requestUrl, body, param, headers, timeoutMs)
}

func postSlsLogisticHttp(httpClient httpcliutil.HTTPCli, ctx context.Context, url string, body interface{}, param map[string]string, headers map[string]string, timeoutMs int32) (xhttp.HttpResponse, error) {
	hTTPResponse, err := httpClient.PostJson(ctx, url, param, body, headers, time.Duration(timeoutMs)*time.Millisecond)
	if hTTPResponse == nil || !hTTPResponse.IsOk() || err != nil {
		errMsg := fmt.Sprintf("failed in postSlsLogisticHttp, url=%s", url)

		if hTTPResponse != nil {
			errMsg = fmt.Sprintf("%s, hTTPResponse=%d", errMsg, hTTPResponse.GetHttpStatus())
		}
		if err != nil {
			errMsg = fmt.Sprintf("%s, err=%s", errMsg, err.Error())
		}

		logging.GetLogger(ctx).Error(errMsg)
		return nil, cerr.New(errMsg, uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
	}
	return hTTPResponse, nil
}

func PostChargeCoreHTTP(
	httpCli httpcliutil.HTTPCli, ctx context.Context, regionDomain string, api string, request SLSBasicRequest,
) (xhttp.HttpResponse, error) {
	requestUrl := fmt.Sprintf(config.GetHTTPApiConfig().ChargeCoreHost, regionDomain) + api
	request.SetToken(config.GetHTTPApiConfig().ChargeCoreToken)
	nonce := hex.EncodeToString(uuid.NewV4().Bytes())
	request.SetRequestID(metadata.GetRequestId(ctx) + "-" + nonce[:8])
	request.SetTimestamp(time.Now().Unix())
	timeout := time.Duration(config.GetHTTPApiConfig().ChargeCoreTimeoutMs) * time.Millisecond

	resp, err := httpCli.PostJson(ctx, requestUrl, nil, request, nil, timeout)

	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf("Post sls http api fail, api=%s, request=%v, err=%v",
			api, cutil.JSONEncode(request), err.Error()))
	} else {
		body, _ := resp.Unmarshal()
		logging.GetLogger(ctx).Info(fmt.Sprintf("Post sls http success,  api=%s, request=%v, status=%d, body=%v",
			api, cutil.JSONEncode(request), resp.GetHttpStatus(), string(body.([]byte))))
	}

	return resp, err
}

func BatchCalcHiddenFee(httpCli httpcliutil.HTTPCli, ctx context.Context,
	regionDomain string, req *model.BatchCalcHiddenFeeRequest) (*model.BatchCalcHiddenFeeResponse, error) {
	httpResp, err := PostChargeCoreHTTP(httpCli, ctx, regionDomain, constant.ChargeCoreBatchCalcHiddenFee, req)
	if err != nil {
		errMsg := fmt.Sprintf("failed to calc hidden fee, queryList=%+v, err=%s",
			cutil.LazyJSONEncoder(req.List), err)
		logging.GetLogger(ctx).Error(errMsg)
		return nil, cerr.New(errMsg, uint32(priceSyncPriceCalculationPb.Constant_ERROR_HTTP_API))
	}

	body, err := httpResp.Unmarshal()
	if err != nil {
		return nil, cerr.New(fmt.Sprintf(
			"failed to unmarshal SLSBasicHTTP response, err=%s", err.Error()),
			uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	var resp model.BatchCalcHiddenFeeResponse
	if err = json.Unmarshal(body.([]byte), &resp); err != nil {
		errMsg := fmt.Sprintf("failed to unmarshal BatchCalcHiddenFeeResponse, queryList=%+v, err=%s, code=%v",
			cutil.LazyJSONEncoder(req.List), err.Error(), httpResp.GetHttpStatus())
		return nil, cerr.New(errMsg, uint32(priceSyncPriceCalculationPb.Constant_ERROR_MARSHAL))
	}

	return &resp, nil
}
