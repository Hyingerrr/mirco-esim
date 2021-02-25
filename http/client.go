package http

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/opentracing/opentracing-go"

	"github.com/go-resty/resty/v2"
	"github.com/jukylin/esim/proxy"
)

type Client struct {
	transports []func() interface{}

	Client *resty.Client // go-resty用法灵活，不必局限于该文件下的SendPOST、SendGET等方法；
}

type Options func(*Client)

type ClientOptions struct{}

func NewClient(opts ...Options) *Client {
	c := &Client{Client: resty.New()}

	for _, opt := range opts {
		opt(c)
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

func (c *Client) CloseIdleConnections(ctx context.Context) {
	c.Client.SetCloseConnection(true)
}

func (c *Client) RC() *resty.Client {
	return c.Client
}

func (c *Client) RequestPostJson(ctx context.Context, addr string, data interface{}, transport http.RoundTripper) (*resty.Response, error) {
	client := c.Client
	if transport != nil {
		client = client.SetTransport(transport)
	}

	req := client.R().SetHeader("Content-Type", "application/json;charset=UTF-8").SetBody(data)
	return c.Do(ctx, http.MethodPost, addr, req)
}

func (c *Client) RequestPost(ctx context.Context, addr string, data interface{},
	header map[string]string, transport http.RoundTripper) (*resty.Response, error) {
	client := c.Client
	if transport != nil {
		client = client.SetTransport(transport)
	}

	req := client.R()
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	req = req.SetBody(data)

	return c.Do(ctx, resty.MethodPost, addr, req)
}

func (c *Client) RequestGet(ctx context.Context, addr string, header map[string]string, transport http.RoundTripper) (*resty.Response, error) {
	client := c.Client
	if transport != nil {
		client = client.SetTransport(transport)
	}

	req := client.R()
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	return c.Do(ctx, resty.MethodGet, addr, req)
}

func (c *Client) Do(ctx context.Context, method, addr string, req *resty.Request) (resp *resty.Response, err error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		tracer := opentracing.GlobalTracer()
		childSpan := tracer.StartSpan(
			"http_call_server",
			opentracing.ChildOf(span.Context()),
		)
		defer func() {
			if err != nil {
				ext.Error.Set(childSpan, true)
				childSpan.LogKV("event", "error", "error.kind", "internal error", "message", err.Error())
				if resp != nil {
					ext.HTTPStatusCode.Set(childSpan, uint16(resp.StatusCode()))
				}
			} else {
				ext.HTTPStatusCode.Set(childSpan, 200)
			}
			childSpan.Finish()
		}()

		var u *url.URL
		u, err = url.Parse(addr)
		if err != nil {
			return nil, err
		}

		ext.HTTPMethod.Set(childSpan, method)
		ext.HTTPUrl.Set(childSpan, u.Path)
		ext.PeerHostname.Set(childSpan, u.Host)
		ext.SpanKindRPCClient.Set(childSpan)
		_ = tracer.Inject(childSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	}

	return req.Execute(method, strings.TrimSpace(addr))
}

// --------------------------------------------------------- //
// ----------------------- 兼容旧方法 ----------------------- //
// ---------------------- 以下方法废弃 ---------------------- //
// --------------------------------------------------------- //
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

func (c *Client) PostForm(ctx context.Context, addr string, data url.Values) (resp *http.Response, err error) {
	return c.Post(ctx, addr, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
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
