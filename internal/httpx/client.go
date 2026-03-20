package httpx

import (
	"context"
	"net/http"

	"github.com/nazar256/shortcut-cli/internal/config"
	shortcutv3 "github.com/nazar256/shortcut-cli/internal/gen/shortcutv3"
)

type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		token:      cfg.APIToken,
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.applyHeaders(req)
	return c.httpClient.Do(req)
}

func (c *Client) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.Do(req.WithContext(ctx))
}

func (c *Client) RequestEditorFn() shortcutv3.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		_ = ctx
		c.applyHeaders(req)
		return nil
	}
}

func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("Shortcut-Token", c.token)
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
}
