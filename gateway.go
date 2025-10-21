package feather

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Gateway represents the core of the Feather API gateway.
//
// It holds all route configurations and registered plugin factories.
// Routes define how incoming requests are matched and forwarded,
// while the PluginRegistry enables dynamic extension of gateway behavior
// (e.g., authentication, rate limiting, logging).
type Gateway struct {
	Routes         []Route `yaml:"routes"`
	PluginRegistry map[string]PluginFactory
}

// New loads a Gateway configuration from the given YAML file.
//
// It reads the file at the provided path, unmarshals its contents,
// and returns an initialized Gateway instance.
func New(path string) (*Gateway, error) {
	var gw Gateway
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("config file not found")
		return nil, err
	}
	if err := yaml.Unmarshal(data, &gw); err != nil {
		fmt.Println("failed to unmarshal config")
		return nil, err
	}
	if err := validateGatewayConfig(gw); err != nil {
		return nil, err
	}
	return &gw, nil
}

// ServeHTTP handles incoming HTTP requests for the Gateway.
//
// This method is used internally to match a request to its route,
// apply the associated plugins as middleware, and proxy the request
// to the appropriate backend service.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := http.HandlerFunc(g.routeAndProxy)

	route, ok := g.match(r.URL.Path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Read the list of plugins in reverse order
	// so they apply in the order they where defined
	for i := len(route.Plugins) - 1; i >= 0; i-- {
		plugin := route.Plugins[i]
		if factory, ok := g.PluginRegistry[plugin.Name]; ok {
			middleware := factory(plugin.Config)
			handler = middleware(handler)
		} else {
			fmt.Printf("unknown plugin: %s\n", plugin.Name)
		}
	}

	handler.ServeHTTP(w, r)
}

// Run starts the HTTP server for the Gateway on the given port.
//
// It binds the Gateway as the main request handler and begins
// listening for incoming connections. The call blocks until the
// server is stopped or an error occurs.
func (g *Gateway) Run(port string) error {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: g,
	}
	fmt.Printf("Listening on port: %s\n", port)
	return srv.ListenAndServe()
}

// RunTLS starts the Gateway with HTTPS enabled on the given port.
//
// It uses the provided certificate and key files path to enable TLS encryption.
// The server listens until it is stopped or an error occurs.
func (g *Gateway) RunTLS(port, cert, key string) error {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: g,
	}

	fmt.Printf("Listening with TLS on port: %s\n", port)
	return srv.ListenAndServeTLS(cert, key)
}

// match returns the first Route that matches the given request path.
//
// It supports exact path matches as well as prefix matches using the
// "/*" wildcard suffix. This method is used internally by the Gateway
// to resolve which route should handle an incoming request.
func (g *Gateway) match(path string) (*Route, bool) {
	for _, route := range g.Routes {
		for _, p := range route.Paths {
			if strings.HasSuffix(p, "/*") {
				prefix := strings.TrimSuffix(p, "/*")
				if strings.HasPrefix(path, prefix) {
					return &route, true
				}
			} else if p == path {
				return &route, true
			}
		}
	}
	return nil, false
}

func validateGatewayConfig(c Gateway) error {
	if len(c.Routes) == 0 {
		return fmt.Errorf("no routes defined")
	}
	for _, r := range c.Routes {
		if r.Name == "" {
			return fmt.Errorf("route missing name")
		}
		if len(r.Paths) == 0 {
			return fmt.Errorf("route %q has no paths", r.Name)
		}
		if r.Backend == "" {
			return fmt.Errorf("route %q missing service reference", r.Name)
		}
	}
	for _, r := range c.Routes {
		if _, err := url.ParseRequestURI(r.Backend); err != nil {
			return fmt.Errorf("invalid URL for service %q: %v", r.Name, err)
		}
	}
	return nil
}
