package httpcliutil

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"git.garena.com/shopee/seller-server/seller-listing/service-kits/ssk/component/cid"
	"git.garena.com/shopee/seller-server/seller-listing/service-kits/ssk/component/xhttp"
)

type HTTPCli interface {
	Get(ctx context.Context, url string, param map[string]string,
		headers map[string]string, timeOut time.Duration) (xhttp.HttpResponse, error)

	Post(ctx context.Context, url string, contentType string, param map[string]string, body interface{},
		headers map[string]string, timeOut time.Duration) (xhttp.HttpResponse, error)

	PostJson(ctx context.Context, url string, param map[string]string, body interface{},
		headers map[string]string, timeOut time.Duration) (xhttp.HttpResponse, error)
}

type httpCli struct {
	cli xhttp.FastHttpClientV2
}

func NewHTTPCli(httpClient xhttp.FastHttpClientV2) HTTPCli {
	cli := &httpCli{
		cli: httpClient,
	}
	return cli
}

func (c *httpCli) Get(ctx context.Context, url string, param map[string]string,
	headers map[string]string, timeOut time.Duration) (xhttp.HttpResponse, error) {

	ctxTime, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	req := xhttp.NewHttpRequest(url, xhttp.FormContentType, xhttp.GET, headers, param, nil)
	resp := xhttp.NewHttpResponse(func(httpCode int, respBody []byte) (data interface{}, busiCode int64, canCache bool, err error) {
		if httpCode != http.StatusOK {
			return nil, 0, false, fmt.Errorf("http response %v", httpCode)
		}

		return respBody, 0, true, nil
	})

	err := c.cli.DoRequest(ctxTime, req, resp, cid.WithCtxCid())

	return resp, err
}

func (c *httpCli) Post(ctx context.Context, url string, contentType string, param map[string]string, body interface{},
	headers map[string]string, timeOut time.Duration) (xhttp.HttpResponse, error) {

	ctxTime, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	req := xhttp.NewHttpRequest(url, contentType, xhttp.POST, headers, param, body)
	resp := xhttp.NewHttpResponse(func(httpCode int, respBody []byte) (data interface{}, busiCode int64, canCache bool, err error) {
		if httpCode != http.StatusOK {
			return nil, 0, false, fmt.Errorf("http response %v", httpCode)
		}

		return respBody, 0, true, nil
	})

	err := c.cli.DoRequest(ctxTime, req, resp, cid.WithCtxCid())

	return resp, err
}

func (c *httpCli) PostJson(ctx context.Context, url string, param map[string]string, body interface{},
	headers map[string]string, timeOut time.Duration) (xhttp.HttpResponse, error) {

	return c.Post(ctx, url, xhttp.JsonContentType, param, body, headers, timeOut)
}
