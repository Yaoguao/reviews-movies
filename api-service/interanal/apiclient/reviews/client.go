package reviews

import (
	"context"
	"fmt"
	"net/url"
	"resty.dev/v3"
	"reviews-movies/api-service/interanal/apiclient"
)

type Client struct {
	base *apiclient.BaseClient
}

func NewReviewClient(base *apiclient.BaseClient) *Client {
	return &Client{base: base}
}

func (c *Client) GetByID(ctx context.Context, id uint64) (body []byte, status int, contentType string, err error) {
	path := fmt.Sprintf("/api/reviews/%d", id)
	return c.base.DoRequest(ctx, resty.MethodGet, path, nil, nil)
}

func (c *Client) GetByCorrelationID(ctx context.Context, corrID string) (body []byte, status int, contentType string, err error) {
	path := fmt.Sprintf("/api/reviews/by-correlation/%s", corrID)
	return c.base.DoRequest(ctx, resty.MethodGet, path, nil, nil)
}

func (c *Client) UpdateReview(ctx context.Context, id uint64, payload []byte, headers map[string][]string) (body []byte, status int, contentType string, err error) {
	path := fmt.Sprintf("/api/reviews/%d", id)
	return c.base.DoRequest(ctx, resty.MethodPatch, path, payload, headers)
}

func (c *Client) DeleteReview(ctx context.Context, id uint64) (body []byte, status int, contentType string, err error) {
	path := fmt.Sprintf("/api/reviews/%d", id)
	return c.base.DoRequest(ctx, resty.MethodDelete, path, nil, nil)
}

func (c *Client) GetListReview(ctx context.Context, movieID, author, rating, sort, page, pageSize string) (body []byte, status int, contentType string, err error) {
	query := url.Values{}
	if movieID != "" {
		query.Set("movie_id", movieID)
	}
	if author != "" {
		query.Set("author", author)
	}
	if rating != "" {
		query.Set("rating", rating)
	}
	if sort != "" {
		query.Set("sort", sort)
	}
	if page != "" {
		query.Set("page", page)
	}
	if pageSize != "" {
		query.Set("page_size", pageSize)
	}
	path := "/api/reviews" + "?" + query.Encode()

	return c.base.DoRequest(ctx, resty.MethodGet, path, nil, nil)
}
