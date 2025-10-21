package feather

import "net/http"

// Middleware represents a function that wraps an http.HandlerFunc.
// It is used by plugins to intercept or modify request handling.
type Middleware func(http.HandlerFunc) http.HandlerFunc

// PluginFactory defines a factory function for creating Middleware
// instances from a configuration map. Factories are registered under
// a unique name and used to construct plugins dynamically.
type PluginFactory func(config map[string]any) Middleware

// Plugin describes a plugin reference in the gateway configuration.
//
// Each plugin entry specifies the plugin's name and its configuration
// parameters, which are passed to the corresponding PluginFactory
// during initialization.
type Plugin struct {
	Name   string         `yaml:"name"`
	Config map[string]any `yaml:"config,omitempty"`
}

// RegisterPlugin registers a plugin factory under the given name.
//
// Registered plugins can later be referenced in route definitions to
// inject custom middleware such as authentication, logging, or rate limiting.
func (g *Gateway) RegisterPlugin(name string, factory PluginFactory) {
	if g.PluginRegistry == nil {
		g.PluginRegistry = make(map[string]PluginFactory)
	}
	g.PluginRegistry[name] = factory
}
