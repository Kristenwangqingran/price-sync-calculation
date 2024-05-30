package main

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"

	"git.garena.com/shopee/common/spkit/runtime"

	"github.com/denisbrodbeck/machineid"

	"git.garena.com/shopee/common/spkit/pkg/spex"
)

const (
	spkitApp       = "spkit"
	spexGatewayURL = "https://http-gateway.spex.test.shopee.sg"
)

type httpHandler struct {
	client *http.Client
}

func NewHTTPHandler() *httpHandler {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &httpHandler{
		client: client,
	}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ski will do smoke test for every submodule, need to register this HTTPApp in all env
	// but only serve request in localhost
	if !runtime.IsLocalHost() {
		http.Error(w, "only enabled in localhost", http.StatusInternalServerError)
	}

	// in localhost spex will use machineId as serve rule
	machineId, err := machineid.ProtectedID(spkitApp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	instanceId, err := spex.InstanceID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	targetURL, _ := url.Parse(spexGatewayURL)
	params := url.Values{}
	params.Add("param", machineId)
	targetURL.RawQuery = params.Encode() // set serve rule in query param

	r.Host = targetURL.Host
	r.Header.Set("x-sp-destination", instanceId) // add instanceId as x-sp-destination header

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = h.client.Transport // use https
	proxy.ServeHTTP(w, r)
}

type healthCheckHandler struct{}

func (h *healthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
