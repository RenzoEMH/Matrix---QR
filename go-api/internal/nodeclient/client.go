package nodeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout     = 5 * time.Second
	statsPath          = "/api/v1/stats"
	maxStatsResponse   = 1 << 20 // 1 MiB
)

// ErrUnavailable means the Node stats API could not be reached or returned
// a response this service cannot use.
var ErrUnavailable = errors.New("stats service unavailable")

// Client calls the Node stats API over HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New builds a client pointing at NODE_API_URL with a 5s timeout.
func New(baseURL string) *Client {
	return newClient(baseURL, defaultTimeout)
}

func newClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// StatsRequest is the body sent to POST /api/v1/stats.
type StatsRequest struct {
	Matrices map[string][][]float64 `json:"matrices"`
}

// StatsResponse is the JSON returned by the Node API.
type StatsResponse struct {
	Max      float64         `json:"max"`
	Min      float64         `json:"min"`
	Average  float64         `json:"average"`
	Sum      float64         `json:"sum"`
	Diagonal map[string]bool `json:"diagonal"`
}

// ComputeStats POSTs Q, R and the rotated matrix to Node.
func (c *Client) ComputeStats(ctx context.Context, req StatsRequest) (*StatsResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	endpoint, err := url.JoinPath(c.baseURL, statsPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(req); err != nil {
		return nil, fmt.Errorf("%w: encode request: %v", ErrUnavailable, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, resp.StatusCode)
	}

	var out StatsResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxStatsResponse)).Decode(&out); err != nil {
		return nil, fmt.Errorf("%w: decode response: %v", ErrUnavailable, err)
	}
	return &out, nil
}
