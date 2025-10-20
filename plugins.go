package gokeeper

import "net/http"

type PluginFactory func(config map[string]any) Middleware

type Plugin struct {
	Name   string         `yaml:"name"`
	Config map[string]any `yaml:"config,omitempty"`
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func (g *Gateway) RegisterPlugin(name string, factory PluginFactory) {
	if g.PluginRegistry == nil {
		g.PluginRegistry = make(map[string]PluginFactory)
	}
	g.PluginRegistry[name] = factory
}
