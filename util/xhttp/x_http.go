package xhttp

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/gogf/gf/v2/frame/g"
)

type XHttpClient struct {
	client *resty.Client
}

func NewXHttpClient() *XHttpClient {
	client := resty.New()

	return &XHttpClient{
		client: client,
	}

}

func (c *XHttpClient) R(ctx context.Context) *resty.Request {
	r := c.client.R()
	r.SetContext(ctx)
	return r
}

func (c *XHttpClient) Client() *resty.Client {
	return c.client
}

func (w *XHttpClient) PostJsonWithHeader(
	ctx context.Context,
	url string,
	requestBody interface{},
	responseBody interface{},
	requestHeaders map[string]string,
	responseHeader map[string]string,
) error { // 返回响应头信息
	req := w.R(ctx).
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		SetResult(responseBody)

	// 添加调用方传入的自定义 header
	for k, v := range requestHeaders {
		req.SetHeader(k, v)
	}

	resp, err := req.Post(url)
	if err != nil {
		g.Log().Errorf(ctx, "HTTP request failed: %v", err)
		return err
	}

	g.Log().Infof(ctx, "HTTP response url:%s raw body: %s", url, resp.String())

	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	if responseHeader != nil {
		for k, v := range resp.Header() {
			responseHeader[k] = v[0]
		}
	}
	return nil
}
