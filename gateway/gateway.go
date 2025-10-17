package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Gateway represents the core reverse-proxy and middleware engine of the framework.
//
// It routes incoming HTTP requests to backend targets based on their path patterns,
// applies user-defined middleware chains (for logging, authentication, rate limiting, etc.),
// and optionally integrates an AuthProvider for JWT/OIDC verification.
//
// Fields:
//
//	Port         – The listening port for the gateway’s HTTP server.
//	router       – Internal route mapping of URL patterns (keys) to backend targets (values).
//	               For example: "/api/*" -> "http://localhost:8080".
//	middlewares  – A list of middleware functions applied in registration order
//	               around each route handler. Each middleware wraps an http.HandlerFunc.
//
// Typical usage:
//
//	gw := gateway.New("8080")
//	gw.Use(loggingMiddleware, metricsMiddleware)
//	gw.Route("/api/*", "http://localhost:8080")
//	gw.Run()
//
// This abstraction allows framework users to define routes, inject custom logic,
// and plug in authentication for any OIDC-compatible provider with minimal setup.
type Gateway struct {
	Port        string
	router      map[string]string
	middlewares []func(http.HandlerFunc) http.HandlerFunc
}

// New initializes and returns a new Gateway instance.
//
// It sets up the internal router map and prepares the gateway to register routes,
// middlewares, and (optionally) an AuthProvider before starting the HTTP server.
//
// Parameters:
//
//	port – The port number (as a string) on which the gateway will listen.
//
// Example:
//
//	gw := gateway.New("8080")
func New(port string) *Gateway {
	return &Gateway{
		Port:   port,
		router: make(map[string]string),
	}
}

// Route registers a new path pattern and its associated backend target.
//
// Parameters:
//
//	pattern – The route pattern to match (exact or with "/*").
//	backend – The backend base URL to which matching requests are forwarded.
//
// Example:
//
//	gw.Route("/api/*", "http://localhost:8080")
//	gw.Route("/status", "http://status.example.com")
func (g *Gateway) Route(pattern, backend string) {
	g.router[pattern] = backend
}

// Use registers one or more middleware functions that wrap route handlers.
//
// Each middleware must have the signature func(http.HandlerFunc) http.HandlerFunc.
// Middlewares are executed in the order they are added, with the first registered
// wrapping the handler outermost (i.e., the last added runs closest to the handler).
//
// Example:
//
//	gw.Use(loggingMiddleware, authMiddleware)
func (g *Gateway) Use(middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// Apply all middlewares in registration before calling ServeHTTP to handle the initial request
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := http.HandlerFunc(g.routeAndProxy)

	for i := len(g.middlewares) - 1; i >= 0; i-- {
		handler = g.middlewares[i](handler)
	}

	handler.ServeHTTP(w, r)
}

// routeAndProxy performs the lookup with the request path and, if it finds a match,
// create a single host reverse-proxy
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

// Run starts the server on the given port
func (g *Gateway) Run() error {
	srv := &http.Server{
		Addr:    ":" + g.Port,
		Handler: g,
	}
	return srv.ListenAndServe()
}

// match returns the backend target that corresponds to a given request path.
//
// It performs a simple routing lookup supporting both exact matches and
// prefix patterns ending with "/*".
func (g *Gateway) match(path string) (string, bool) {
	for pattern, backend := range g.router {
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
