package http

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jukylin/esim/config"

	"github.com/jukylin/esim/container"
	logx "github.com/jukylin/esim/log"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/opentracing/opentracing-go"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	client     *resty.Client // go-resty灵活使用
	transports []func() interface{}
	isTrace    bool
	isMetric   bool
}

type Options func(*Client)

func NewClient(opts ...Options) *Client {
	c := &Client{client: resty.New()}

	for _, opt := range opts {
		opt(c)
	}

	if c.client.GetClient().Timeout <= 0 {
		c.client.SetTimeout(5 * time.Second)
	}

	c.isMetric = config.GetBool("http_client_metrics")
	c.isTrace = config.GetBool("http_client_tracer")

	return c
}

func WithProxy(proxy ...func() interface{}) Options {
	return func(hc *Client) {
		hc.transports = append(hc.transports, proxy...)
	}
}

func WithTimeOut(timeout time.Duration) Options {
	return func(hc *Client) {
		hc.client.SetTimeout(timeout)
	}
}

func (c *Client) CloseIdleConnections(ctx context.Context) {
	c.client.SetCloseConnection(true)
}

func (c *Client) RC() *resty.Client {
	return c.client
}

func (c *Client) RequestPostJson(ctx context.Context, addr string, data interface{}, transport http.RoundTripper) (*resty.Response, error) {
	client := c.client
	if transport != nil {
		client = client.SetTransport(transport)
	}

	req := client.R().SetHeader("Content-Type", "application/json;charset=UTF-8").SetBody(data)
	return c.Do(ctx, http.MethodPost, addr, req)
}

func (c *Client) RequestPost(ctx context.Context, addr string, data interface{},
	header map[string]string, transport http.RoundTripper) (*resty.Response, error) {
	client := c.client
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
	client := c.client
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
	var beg = time.Now()
	var u *url.URL

	u, err = url.Parse(addr)
	if err != nil {
		logx.Errorc(ctx, "url parse error:%v, addr[%v]", err, addr)
		return nil, err
	}
	if !c.isTrace {
		goto Next
	}

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

		ext.HTTPMethod.Set(childSpan, method)
		ext.HTTPUrl.Set(childSpan, u.Path)
		ext.PeerHostname.Set(childSpan, u.Host)
		ext.SpanKindRPCClient.Set(childSpan)
		_ = tracer.Inject(childSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	}

Next:
	resp, err = req.Execute(method, strings.TrimSpace(addr))
	if err != nil {
		logx.Errorc(ctx, "http call net error:%v, method[%v], host[%v], path[%v], cost[%v], body[%+v]",
			err, method, u.Host, u.Path, time.Since(beg).String(), req.Body)
		if c.isMetric {
			httpCallReqError.Inc(container.AppName(), u.Path)
		}
		return
	}

	if c.isMetric {
		httpCallRespCount.Inc(container.AppName(), u.Path, strconv.Itoa(resp.StatusCode()))
		httpCallReqDuration.Observe(float64(time.Since(beg)/time.Millisecond), u.Path)
	}

	return
}
