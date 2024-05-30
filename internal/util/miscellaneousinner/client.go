package miscellaneousinner

import (
	"bytes"
	"io"
	"net/http"
	nu "net/url"
	"strconv"
	"strings"
	"sync"

	"git.garena.com/shopee/platform/golang_splib/client"
	sphttp "git.garena.com/shopee/platform/golang_splib/client/http"
	"git.garena.com/shopee/platform/golang_splib/env"
)

type Client struct {
	c    sphttp.Client
	once sync.Once
}

func defaultClient() *Client {
	cli, err := NewClient()
	if err != nil {
		return nil
	}

	return cli
}

var dc = defaultClient()

func NewClient(opts ...client.Option) (*Client, error) {
	c := Client{}

	enableMetrics, err := strconv.ParseBool(env.Get(env.SpEnableMetrics, ""))
	if err == nil && enableMetrics {
		enableMetrics = true
	}

	if enableMetrics {
		opts = append(opts, client.WithMiddleware(ClientVersionReport(version, serviceName)))
	}

	if err := c.init(opts...); err != nil {
		return nil, err
	}

	return &c, nil
}

func Get(url string) (resp *http.Response, err error) {
	return dc.Get(url)
}

func (c *Client) init(opts ...client.Option) error {
	var err error

	c.once.Do(func() {
		if c.c == nil {
			c.c, err = sphttp.NewClient(opts...)
			if err != nil {
				return
			}

			c.c.WithRouter(router())
		}
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if err := c.init(); err != nil {
		return nil, err
	}

	return c.c.Do(req)
}

func Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	return dc.Post(url, contentType, body)
}

func (c *Client) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	// nolint:noctx // go1.12 can't set context
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return c.Do(req)
}

func PostForm(url string, data nu.Values) (resp *http.Response, err error) {
	return dc.PostForm(url, data)
}

func (c *Client) PostForm(url string, data nu.Values) (resp *http.Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func Head(url string) (resp *http.Response, err error) {
	return dc.Head(url)
}

func (c *Client) Head(url string) (resp *http.Response, err error) {
	// nolint:noctx // go1.12 can't set context
	req, err := http.NewRequest("HEAD", url, bytes.NewReader([]byte{}))
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) CloseIdleConnections() {
}

func (c *Client) Command(method, path string) (string, bool) {
	r, ok := router().Lookup(method, path)
	return r.Command, ok
}
