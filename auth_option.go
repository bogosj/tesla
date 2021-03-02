package tesla

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/oauth2"
)

// github.com/uhthomas/tesla
func state() string {
	var b [9]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:])
}

// https://www.oauth.com/oauth2-servers/pkce/
func pkce() (verifier, challenge string, err error) {
	var p [87]byte
	if _, err := io.ReadFull(rand.Reader, p[:]); err != nil {
		return "", "", fmt.Errorf("rand read full: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(p[:])
	b := sha256.Sum256([]byte(challenge))
	challenge = base64.RawURLEncoding.EncodeToString(b[:])
	return verifier, challenge, nil
}

func authHandler() *auth {
	return &auth{
		SelectDevice: mfaUnsupported,
	}
}

// WithMFAHandler allows a consumer to provide a different configuration from the default.
func WithMFAHandler(handler func(context.Context, []Device) (Device, string, error)) ClientOption {
	return func(c *Client) error {
		if c.authHandler == nil {
			c.authHandler = authHandler()
		}
		c.authHandler.SelectDevice = handler
		return nil
	}
}

func mfaUnsupported(_ context.Context, _ []Device) (Device, string, error) {
	return Device{}, "", errors.New("multi factor authentication is not supported")
}

// WithCredentials allows a consumer to provide a different configuration from the default.
func WithCredentials(username, password string) ClientOption {
	return func(c *Client) error {
		if c.authHandler == nil {
			c.authHandler = authHandler()
		}

		verifier, challenge, err := pkce()
		if err != nil {
			return err
		}

		c.authHandler.AuthURL = c.oc.AuthCodeURL(state(), oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)

		ctx := context.Background()
		code, err := c.authHandler.Do(ctx, username, password)
		if err != nil {
			return err
		}

		token, err := c.oc.Exchange(ctx, code,
			oauth2.SetAuthURLParam("code_verifier", verifier),
		)

		if err == nil {
			c.token = token
		}

		return err
	}
}
