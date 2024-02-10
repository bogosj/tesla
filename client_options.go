package tesla

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

// ClientOption can be passed when creating the client
type ClientOption func(c *Client) error

// WithToken provides an oauth2.Token to the client for auth.
func WithToken(t *oauth2.Token) ClientOption {
	return func(c *Client) error {
		c.token = t
		return nil
	}
}

func loadToken(path string) (*oauth2.Token, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tok := new(oauth2.Token)
	if err := json.Unmarshal(b, tok); err != nil {
		return nil, err
	}
	return tok, nil
}

// WithTokenFile reads a JSON serialized oauth2.Token struct from disk and provides it
// to the client for auth.
func WithTokenFile(path string) ClientOption {
	t, err := loadToken(path)
	if err != nil {
		return func(c *Client) error {
			return err
		}
	}
	return WithToken(t)
}

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

// WithOAuth2Config allows a consumer to provide a different configuration from the default.
func WithOAuth2Config(oc *oauth2.Config) ClientOption {
	return func(c *Client) error {
		c.oc = oc
		return nil
	}
}
