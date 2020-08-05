package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/proxy"
)

type Client struct {
	logger log.Logger

	transports []func() interface{}

	Client *resty.Client // go-resty用法灵活，大写公开，不必局限于该文件下的SendPOST、SendGET等方法；
}

type Options func(*Client)

type ClientOptions struct{}

func NewClient(opts ...Options) *Client {
	c := &Client{Client: resty.New()}

	for _, opt := range opts {
		opt(c)
	}

	if c.logger == nil {
		c.logger = log.NewLogger()
	}

	if c.transports == nil {
		c.Client.SetTransport(http.DefaultTransport)
	} else {
		transport := proxy.NewProxyFactory().
			GetFirstInstance("http", http.DefaultTransport, c.transports...).(http.RoundTripper)
		c.Client.SetTransport(transport)
	}

	if c.Client.GetClient().Timeout <= 0 {
		c.Client.SetTimeout(30 * time.Second)
	}

	return c
}

func (ClientOptions) WithProxy(proxys ...func() interface{}) Options {
	return func(hc *Client) {
		hc.transports = append(hc.transports, proxys...)
	}
}

func (ClientOptions) WithTimeOut(timeout time.Duration) Options {
	return func(hc *Client) {
		hc.Client.SetTimeout(timeout * time.Second)
	}
}

func (ClientOptions) WithLogger(logger log.Logger) Options {
	return func(hc *Client) {
		hc.logger = logger
	}
}

// with TLS/SSL
func (ClientOptions) WithInsecureSkip() Options {
	return func(hc *Client) {
		hc.Client.SetTransport(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}})
	}
}

func (c *Client) SetKeepAliveDisable(b bool) *Client {
	c.Client = c.Client.SetTransport(&http.Transport{
		DisableKeepAlives: b,
	})
	return c
}

func (c *Client) SendGet(ctx context.Context, addr string) (rtyResp *resty.Response, err error) {
	return c.Client.R().SetContext(ctx).EnableTrace().Get(strings.TrimSpace(addr))
}

func (c *Client) SendPost(ctx context.Context, addr, contentType string, body io.Reader) (rtyResp *resty.Response, err error) {
	return c.Client.R().SetContext(ctx).EnableTrace().
		SetHeader("Content-Type", contentType).
		SetBody(body).
		Post(strings.TrimSpace(addr))
}

// 兼容老的Get方法, 不建议再使用
func (c *Client) Get(ctx context.Context, addr string) (resp *http.Response, err error) {
	var (
		rtyResp *resty.Response
	)

	rtyResp, err = c.Client.R().SetContext(ctx).EnableTrace().Get(strings.TrimSpace(addr))
	if err != nil {
		return nil, err
	}

	rtyResp.RawResponse.Body = ioutil.NopCloser(bytes.NewReader(rtyResp.Body()))
	return rtyResp.RawResponse, err
}

// 兼容老的POST方法，不建议再使用
func (c *Client) Post(ctx context.Context, addr, contentType string, body io.Reader) (resp *http.Response, err error) {
	var (
		rtyResp *resty.Response
	)

	rtyResp, err = c.Client.R().SetContext(ctx).EnableTrace().
		SetHeader("Content-Type", contentType).
		SetBody(body).
		Post(strings.TrimSpace(addr))
	if err != nil {
		return nil, err
	}

	rtyResp.RawResponse.Body = ioutil.NopCloser(bytes.NewReader(rtyResp.Body()))

	return rtyResp.RawResponse, err
}

// 兼容老PostForm方法，不建议再使用
func (c *Client) PostForm(ctx context.Context, addr string, data url.Values) (resp *http.Response, err error) {
	return c.Post(ctx, addr, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c *Client) CloseIdleConnections(ctx context.Context) {
	c.Client.SetCloseConnection(true)
}
