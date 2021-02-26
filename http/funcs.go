package http

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
)

// ---------------------------------------------------------------------------------------------- //
// ----------------------------------------- 兼容旧方法 ------------------------------------------ //
// --------------------------------------- 以下方法不建议使用 -------------------------------------- //
// --------------------------------------------------------------------------------------------- //
func (c *Client) Get(ctx context.Context, addr string) (resp *http.Response, err error) {
	var (
		rtyResp *resty.Response
	)

	rtyResp, err = c.client.R().SetContext(ctx).EnableTrace().Get(strings.TrimSpace(addr))
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

	rtyResp, err = c.client.R().SetContext(ctx).EnableTrace().
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
	return c.client.R().SetContext(ctx).EnableTrace().Get(strings.TrimSpace(addr))
}

func (c *Client) SendPost(ctx context.Context, addr, contentType string, body io.Reader) (rtyResp *resty.Response, err error) {
	return c.client.R().SetContext(ctx).EnableTrace().
		SetHeader("Content-Type", contentType).
		SetBody(body).
		Post(strings.TrimSpace(addr))
}
