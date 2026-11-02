package upstream

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client проксирует HTTP-запросы на upstream с исходящей auth (Basic/Bearer).
type Client struct {
	host    string
	auth    Auth
	timeout time.Duration
	http    *http.Client
}

func New(cfg Config) (*Client, error) {
	def := DefaultConfig()
	if cfg.Host == "" {
		cfg.Host = def.Host
	}
	if cfg.ResponseTimeout == "" {
		cfg.ResponseTimeout = def.ResponseTimeout
	}
	if cfg.Auth.Type == "" {
		cfg.Auth = def.Auth
	}
	timeout, err := time.ParseDuration(cfg.ResponseTimeout)
	if err != nil {
		return nil, fmt.Errorf("upstream: response_timeout: %w", err)
	}
	return &Client{
		host:    strings.TrimRight(cfg.Host, "/"),
		auth:    cfg.Auth,
		timeout: timeout,
		http:    &http.Client{Timeout: timeout},
	}, nil
}

// Forward — тот же method/path/query/body, без клиентского Authorization.
func (c *Client) Forward(ctx context.Context, method, path string, query url.Values, body []byte, clientHeaders http.Header) (status int, respBody []byte, err error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := c.host + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return 0, nil, err
	}
	if len(body) > 0 {
		req.Body = io.NopCloser(bytes.NewReader(body))
		req.ContentLength = int64(len(body))
	}
	c.applyOutgoingAuth(req)
	copyForwardHeaders(req.Header, clientHeaders)

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, raw, nil
}

func (c *Client) applyOutgoingAuth(req *http.Request) {
	switch strings.ToLower(c.auth.Type) {
	case "basic":
		req.SetBasicAuth(c.auth.BasicLogin, c.auth.BasicPassword)
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+c.auth.BearerToken)
	}
}

var hopByHop = map[string]struct{}{
	"Connection":        {},
	"Keep-Alive":        {},
	"Proxy-Authenticate": {},
	"Proxy-Authorization": {},
	"Te":                {},
	"Trailers":          {},
	"Transfer-Encoding": {},
	"Upgrade":           {},
	"Authorization":     {},
	"Proxy-Connection":  {},
	"Content-Length":    {},
}

func copyForwardHeaders(dst, src http.Header) {
	for k, vals := range src {
		if _, skip := hopByHop[http.CanonicalHeaderKey(k)]; skip {
			continue
		}
		for _, v := range vals {
			dst.Add(k, v)
		}
	}
	if dst.Get("Content-Type") == "" {
		dst.Set("Content-Type", "application/json")
	}
}
