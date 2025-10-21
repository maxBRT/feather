package feather

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Route defines how incoming requests are matched and forwarded.
//
// Each route includes one or more path patterns, the backend service
// address to proxy matching requests to, and an optional list of
// plugins that modify request handling (e.g., authentication,
// rate limiting, logging).
type Route struct {
	Name    string   `yaml:"name"`
	Paths   []string `yaml:"paths"`
	Backend string   `yaml:"backend"`
	Plugins []Plugin `yaml:"plugins"`
}

// routeAndProxy handles request forwarding for a matched route.
//
// It finds the appropriate backend using the route definition,
// builds a reverse proxy, and delegates the request to the target
// service. This method is used internally by the Gateway and is not
// intended to be called directly.
func (g *Gateway) routeAndProxy(w http.ResponseWriter, r *http.Request) {
	route, ok := g.match(r.URL.Path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	fmt.Printf("Backend Found: %s\n", route.Name)

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: route.Backend})
	originalDirector := proxy.Director

	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		fmt.Printf("Redirecting request to: %s\n", req.URL.String())
	}

	proxy.ServeHTTP(w, r)
}
