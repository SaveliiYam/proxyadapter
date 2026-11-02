package ginroutes_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SaveliiYam/proxyadapter/ginroutes"
	"github.com/SaveliiYam/proxyadapter/upstream"
	"github.com/gin-gonic/gin"
)

func TestProxyHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstreamSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/ping" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "svc" || pass != "secret" {
			t.Fatalf("basic auth missing")
		}
		b, _ := io.ReadAll(r.Body)
		if string(b) != `{"x":1}` {
			t.Fatalf("body %q", b)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	t.Cleanup(upstreamSrv.Close)

	cli, err := upstream.New(upstream.Config{
		Host:            upstreamSrv.URL,
		ResponseTimeout: "2s",
		Auth: upstream.Auth{
			Type:          "basic",
			BasicLogin:    "svc",
			BasicPassword: "secret",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.POST("/api/v1/ping", ginroutes.NewProxyHandler(cli, "internal"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ping", strings.NewReader(`{"x":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer client-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != `{"data":{"ok":true}}` {
		t.Fatalf("body %s", rec.Body.String())
	}
}
