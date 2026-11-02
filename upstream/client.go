package upstream

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	httpclient "github.com/SaveliiYam/http_client"
)

type Client struct {
	hc *httpclient.Client
}

// New создаёт клиент из http-client Config (тот же JSON, что auth_client.json / jwt_client.json).
func New(cfg httpclient.Config) (*Client, error) {
	hc, err := httpclient.New(&cfg)
	if err != nil {
		return nil, err
	}
	return &Client{hc: hc}, nil
}

// Forward — method/path/query/body; клиентский Authorization не пробрасывается (задаётся Auth в Config).
func (c *Client) Forward(ctx context.Context, method, path string, query url.Values, body []byte, clientHeaders http.Header) (status int, respBody []byte, err error) {
	raw, status, err := c.hc.DoPath(ctx, method, path, body, valuesToMap(query), filterHeaders(clientHeaders))
	if err != nil {
		var he *httpclient.HTTPError
		if errors.As(err, &he) {
			return he.Status, append([]byte(nil), he.Body...), nil
		}
		return 0, nil, err
	}
	return status, raw, nil
}

func valuesToMap(q url.Values) map[string]string {
	if len(q) == 0 {
		return nil
	}
	out := make(map[string]string, len(q))
	for k, vals := range q {
		if len(vals) > 0 {
			out[k] = vals[0]
		}
	}
	return out
}

var hopByHop = map[string]struct{}{
	"Connection":          {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"Te":                  {},
	"Trailers":            {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
	"Authorization":       {},
	"Proxy-Connection":    {},
	"Content-Length":      {},
}

func filterHeaders(src http.Header) map[string]string {
	if len(src) == 0 {
		return map[string]string{"Content-Type": "application/json"}
	}
	out := make(map[string]string)
	for k, vals := range src {
		if _, skip := hopByHop[http.CanonicalHeaderKey(k)]; skip {
			continue
		}
		if len(vals) > 0 {
			out[k] = vals[0]
		}
	}
	if out["Content-Type"] == "" {
		out["Content-Type"] = "application/json"
	}
	return out
}
