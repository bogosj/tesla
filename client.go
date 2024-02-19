package tesla

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

// Set to true to make the json decoder fail to decode if it encounters an
// unknown field.
const disallowUnknownFields = false

const (
	FleetAudienceNA = "https://fleet-api.prd.na.vn.cloud.tesla.com/"
	FleetAudienceEU = "https://fleet-api.prd.eu.vn.cloud.tesla.com/"
)

// Client provides the client and associated elements for interacting with the Tesla API.
type Client struct {
	baseURL string
	hc      *http.Client
	ts      oauth2.TokenSource
}

// NewClient creates a new Tesla API client. You must provided one of WithToken or WithTokenFile
// functional options to initialize the client with an OAuth token.
func NewClient(ctx context.Context, options ...ClientOption) (*Client, error) {
	client := &Client{
		baseURL: FleetAudienceNA + "api/1",
	}

	for _, option := range options {
		err := option(client)
		if err != nil {
			return nil, err
		}
	}

	if client.hc == nil {
		client.hc = http.DefaultClient

		if client.ts == nil {
			return nil, errors.New("missing token source")
		}

		client.hc.Transport = &oauth2.Transport{
			Source: client.ts,
			Base:   client.hc.Transport,
		}
	}

	return client, nil
}

// Sets new base url. Use after obtaining user's region.
func (c *Client) SetBaseUrl(url string) {
	c.SetApiUrl(strings.TrimRight(url, "/") + "/api/1")
}

// Sets new api url. Use after obtaining user's region.
// It is clients responsibility to append `/api/1` for use with the Fleet API.
func (c *Client) SetApiUrl(url string) {
	c.baseURL = url
}

// Calls an HTTP GET
func (c *Client) get(url string) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	return c.processRequest(req)
}

// getJSON performs an HTTP GET and then unmarshals the result into the provided struct.
func (c *Client) getJSON(url string, out interface{}) error {
	body, err := c.get(url)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if disallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	return decoder.Decode(out)
}

// Calls an HTTP POST with a JSON body
func (c *Client) post(url string, body []byte) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	return c.processRequest(req)
}

// Processes a HTTP POST/PUT request
func (c *Client) processRequest(req *http.Request) ([]byte, error) {
	c.setHeaders(req)
	res, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return body, errors.New(res.Status)
	}
	return body, nil
}

// Sets the required headers for calls to the Tesla API
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}
