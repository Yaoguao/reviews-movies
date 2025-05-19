package apiclient

import (
	"context"
	"fmt"
	"io"
	"resty.dev/v3"
	"time"
)

type BaseClient struct {
	http *resty.Client
}

func NewBaseClient(baseURL string, timeoutSec int) *BaseClient {
	cli := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(time.Duration(timeoutSec) * time.Second)
	return &BaseClient{http: cli}
}

func (c *BaseClient) DoRequest(ctx context.Context, method string, path string, payload []byte, headers map[string][]string) (body []byte, httpStatus int, contentType string, err error) {
	req := c.http.R().SetContext(ctx)

	if payload != nil {
		req.SetBody(payload)
	}

	for key, values := range headers {
		for _, v := range values {
			req.SetHeader(key, v)
		}
	}

	var resp *resty.Response
	switch method {
	case resty.MethodGet:
		resp, err = req.Get(path)
	case resty.MethodPatch:
		resp, err = req.Patch(path)
	case resty.MethodDelete:
		resp, err = req.Delete(path)
	case resty.MethodPost:
		resp, err = req.Post(path)
	case resty.MethodPut:
		resp, err = req.Put(path)
	default:
		return nil, 0, "", fmt.Errorf("unsupported method %q", method)
	}
	if err != nil {
		return nil, 0, "", fmt.Errorf("resty %s error: %w", method, err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, "", err
	}

	return body, resp.StatusCode(), resp.Header().Get("Content-Type"), nil
}
