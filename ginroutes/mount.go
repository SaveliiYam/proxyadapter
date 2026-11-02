package ginroutes

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	pajwt "github.com/SaveliiYam/proxyadapter/jwt"
	"github.com/SaveliiYam/proxyadapter/upstream"
	"github.com/SaveliiYam/yamlroutes"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// MountConfig — монтирование proxy-адаптера (ingress без basic, опционально jwt).
type MountConfig struct {
	Upstream          *upstream.Client
	AccessParser      pajwt.AccessParser
	InvalidTokenCode  string
	InternalErrCode   string
	ProxyDo           string
	AllowIngressBasic bool
}

var ingressAuth = map[string]struct{}{
	"basic": {},
	"jwt":   {},
}

// MountHandlersFile регистрирует handlers и монтирует routes.yaml.
//
// Если задан Upstream, автоматически регистрируется ProxyDo (по умолчанию auth_client_proxy).
// auth: [basic] на ingress запрещён, если AllowIngressBasic == false.
func MountHandlersFile(router gin.IRouter, path string, handlers map[string]gin.HandlerFunc, cfg MountConfig) error {
	if cfg.ProxyDo == "" {
		cfg.ProxyDo = DefaultProxyDo
	}
	if handlers == nil {
		handlers = make(map[string]gin.HandlerFunc)
	}
	if cfg.Upstream != nil {
		if _, exists := handlers[cfg.ProxyDo]; !exists {
			handlers[cfg.ProxyDo] = NewProxyHandler(cfg.Upstream, cfg.InternalErrCode)
		}
	}

	spec, err := loadSpec(path)
	if err != nil {
		return err
	}
	for _, route := range spec.Routes {
		h, ok := handlers[route.Do]
		if !ok {
			return fmt.Errorf("proxyadapter: handler %q not registered", route.Do)
		}
		chain := []gin.HandlerFunc{h}
		if route.HasAuth("jwt") {
			if cfg.AccessParser == nil {
				return fmt.Errorf("proxyadapter: %s %s requires jwt but AccessParser is nil", route.Method, route.Path)
			}
			chain = append([]gin.HandlerFunc{RequireAccessJWT(cfg.AccessParser, cfg.InvalidTokenCode)}, chain...)
		}
		if route.HasAuth("basic") && !cfg.AllowIngressBasic {
			return fmt.Errorf("proxyadapter: %s %s: ingress basic auth is disabled", route.Method, route.Path)
		}
		router.Handle(route.Method, route.Path, chain...)
	}
	return nil
}

func loadSpec(path string) (yamlroutes.Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return yamlroutes.Spec{}, err
	}
	var spec yamlroutes.Spec
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)
	if err := dec.Decode(&spec); err != nil {
		return yamlroutes.Spec{}, fmt.Errorf("proxyadapter: parse: %w", err)
	}
	if err := validateIngressSpec(spec); err != nil {
		return yamlroutes.Spec{}, err
	}
	return spec, nil
}

func validateIngressSpec(s yamlroutes.Spec) error {
	seen := make(map[string]bool, len(s.Routes))
	for i, r := range s.Routes {
		where := fmt.Sprintf("routes[%d] (%s %s -> %s)", i, r.Method, r.Path, r.Do)
		switch r.Method {
		case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
			http.MethodDelete, http.MethodHead, http.MethodOptions:
		default:
			return fmt.Errorf("proxyadapter: %s: unknown method %q", where, r.Method)
		}
		if !strings.HasPrefix(r.Path, "/") {
			return fmt.Errorf("proxyadapter: %s: path must start with /", where)
		}
		if strings.TrimSpace(r.Do) == "" {
			return fmt.Errorf("proxyadapter: %s: do is empty", where)
		}
		for _, a := range r.Auth {
			key := strings.ToLower(strings.TrimSpace(a))
			if key == "" {
				return fmt.Errorf("proxyadapter: %s: empty auth entry", where)
			}
			if _, ok := ingressAuth[key]; !ok {
				return fmt.Errorf("proxyadapter: %s: unknown auth %q (supported: basic, jwt)", where, a)
			}
		}
		key := r.Method + " " + r.Path
		if seen[key] {
			return fmt.Errorf("proxyadapter: %s: duplicate route", where)
		}
		seen[key] = true
	}
	return nil
}
