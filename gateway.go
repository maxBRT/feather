package gokeeper

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Gateway struct {
	Routes         []Route `yaml:"routes"`
	PluginRegistry map[string]PluginFactory
}

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

	return &gw, nil
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := http.HandlerFunc(g.routeAndProxy)

	route, ok := g.match(r.URL.Path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	for _, plugin := range route.Plugins {
		if factory, ok := g.PluginRegistry[plugin.Name]; ok {
			middleware := factory(plugin.Config)

			handler = middleware(handler)
		} else {
			fmt.Printf("unknown plugin: %s", plugin.Name)
		}
	}

	handler.ServeHTTP(w, r)
}

func (g *Gateway) Run(port string) error {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: g,
	}
	fmt.Printf("Listening on port: %s\n", port)
	return srv.ListenAndServe()
}

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
