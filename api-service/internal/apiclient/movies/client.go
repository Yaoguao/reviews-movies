package movies

import (
	"context"
	"fmt"
	"net/url"
	"resty.dev/v3"
	"reviews-movies/api-service/internal/apiclient"
)

type Client struct {
	base *apiclient.BaseClient
}

func NewMovieClient(base *apiclient.BaseClient) *Client {
	return &Client{base: base}
}

func (c *Client) GetByID(ctx context.Context, id uint64) (body []byte, status int, contentType string, err error) {
	path := fmt.Sprintf("/api/movies/%d", id)
	return c.base.DoRequest(ctx, resty.MethodGet, path, nil, nil)
}

func (c *Client) GetByCorrelationID(ctx context.Context, corrID string) (body []byte, status int, contentType string, err error) {
	path := fmt.Sprintf("/api/movies/by-correlation/%s", corrID)
	return c.base.DoRequest(ctx, resty.MethodGet, path, nil, nil)
}

func (c *Client) UpdateMovie(ctx context.Context, id uint64, payload []byte, headers map[string][]string) (body []byte, status int, contentType string, err error) {
	path := fmt.Sprintf("/api/movies/%d", id)
	return c.base.DoRequest(ctx, resty.MethodPatch, path, payload, headers)
}

func (c *Client) DeleteMovie(ctx context.Context, id uint64) (body []byte, status int, contentType string, err error) {
	path := fmt.Sprintf("/api/movies/%d", id)
	return c.base.DoRequest(ctx, resty.MethodDelete, path, nil, nil)
}

func (c *Client) GetListMovie(ctx context.Context, title, sort, page, pageSize string, genres []string) (body []byte, status int, contentType string, err error) {
	query := url.Values{}
	if title != "" {
		query.Set("title", title)
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
	for _, genre := range genres {
		if genre != "" {
			query.Add("genres", genre)
		}
	}
	path := "/api/movies" + "?" + query.Encode()

	return c.base.DoRequest(ctx, resty.MethodGet, path, nil, nil)
}

func (c *Client) GetTopRatedMovies(ctx context.Context) (body []byte, status int, contentType string, err error) {
	path := "/api/movies/top"
	return c.base.DoRequest(ctx, resty.MethodGet, path, nil, nil)
}

func (c *Client) GetWithoutReviews(ctx context.Context) (body []byte, status int, contentType string, err error) {
	path := "/api/movies/without-reviews"
	return c.base.DoRequest(ctx, resty.MethodGet, path, nil, nil)
}

func (c *Client) GetControversialMovies(ctx context.Context) (body []byte, status int, contentType string, err error) {
	path := "/api/movies/variance"
	return c.base.DoRequest(ctx, resty.MethodGet, path, nil, nil)
}

func (c *Client) GetAvgRatingByGenre(ctx context.Context) (body []byte, status int, contentType string, err error) {
	path := "/api/movies/avg-rating"
	return c.base.DoRequest(ctx, resty.MethodGet, path, nil, nil)
}
