package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

type Gateway struct {
	Port        string
	Router      map[string]string
	Config      *viper.Viper
	middlewares []func(http.HandlerFunc) http.HandlerFunc
}

func New(port string) *Gateway {
	return &Gateway{
		Port:   port,
		Router: make(map[string]string),
	}
}

func (g *Gateway) match(path string) (string, bool) {
	for pattern, backend := range g.Router {
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, prefix) {
				return backend, true
			}
		} else if pattern == path {
			return backend, true
		}
	}
	return "", false
}

func (g *Gateway) Route(pattern, backend string) {
	g.Router[pattern] = backend
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := http.HandlerFunc(g.routeAndProxy)
	handler.ServeHTTP(w, r)
}

func (g *Gateway) routeAndProxy(w http.ResponseWriter, r *http.Request) {
	backendURL, ok := g.match(r.URL.Path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	fmt.Printf("Backend Found: %s\n", backendURL)
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: backendURL})

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		fmt.Printf("Redirecting request to: %s\n", req.URL.String())
	}
	proxy.ServeHTTP(w, r)
}

func (g *Gateway) Run() error {
	srv := &http.Server{
		Addr:    ":" + g.Port,
		Handler: g,
	}
	return srv.ListenAndServe()
}
