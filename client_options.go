package tesla

import (
	"net/http"

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
