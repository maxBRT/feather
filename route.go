package gokeeper

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Route struct {
	Name     string   `yaml:"name"`
	Paths    []string `yaml:"paths"`
	Backends []string `yaml:"backends"`
	Plugins  []Plugin `yaml:"plugins"`
}

func (g *Gateway) routeAndProxy(w http.ResponseWriter, r *http.Request) {
	route, ok := g.match(r.URL.Path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	fmt.Printf("Backend Found: %s\n", route.Name)
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: route.Backends[0]})
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		fmt.Printf("Redirecting request to: %s\n", req.URL.String())
	}
	proxy.ServeHTTP(w, r)
}
