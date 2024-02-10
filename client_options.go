package tesla

import (
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

// ClientOption can be passed when creating the client
type ClientOption func(c *Client) error

// WithTokenSource provides an oauth2.TokenSource to the client for auth
func WithTokenSource(ts oauth2.TokenSource) ClientOption {
	return func(c *Client) error {
		c.ts = ts
		return nil
	}
}

// WithClient provides set the http.Client
func WithClient(client *http.Client) ClientOption {
	return func(c *Client) error {
		c.hc = client
		return nil
	}
}

// WithBaseURL provides a method to set the base URL for standard API calls to differ
// from the default.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) error {
		c.baseURL = strings.TrimRight(url, "/")
		return nil
	}
}
