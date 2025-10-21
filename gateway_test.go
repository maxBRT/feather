package feather

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGateway(t *testing.T) {
	// Case 1: valid configuration
	yamlData := `
routes:
  - name: users
    paths: ["/users/*"]
    backend: "localhost:8080"
`
	tmp, err := os.CreateTemp("", "config.yaml")

	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(yamlData); err != nil {
		t.Fatalf("failed to write to temporary file: %v", err)
	}

	gw, err := New(tmp.Name())
	assert.NoError(t, err)
	assert.NotNil(t, gw)

	// Case 2: invalid backend URL
	yamlData = `
routes:
  - name: users
    paths: ["/users/*"]
    backend: "localhost8080"
`
	tmp, err = os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(yamlData); err != nil {
		t.Fatalf("failed to write to temporary file: %v", err)
	}

	gw, err = New(tmp.Name())
	assert.Error(t, err)
	assert.Nil(t, gw)

	// Case 3: invalid route name
	yamlData = `
routes:
  - name: 
    paths: ["/users/*"]
    backend: "localhost:8080"
`
	tmp, err = os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(yamlData); err != nil {
		t.Fatalf("failed to write to temporary file: %v", err)
	}

	gw, err = New(tmp.Name())
	assert.Error(t, err)
	assert.Nil(t, gw)

	// Case 4: Empty path
	yamlData = `
routes:
  - name: users
    paths: []
    backend: "localhost:8080"
`
	tmp, err = os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(yamlData); err != nil {
		t.Fatalf("failed to write to temporary file: %v", err)
	}

	gw, err = New(tmp.Name())
	assert.Error(t, err)
	assert.Nil(t, gw)
}

func TestMatch(t *testing.T) {
	// Case 1: exact match
	yamlData := `
routes:
  - name: users
    paths: ["/users"]
    backend: "localhost:8080"
`
	tmp, err := os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(yamlData); err != nil {
		t.Fatalf("failed to write to temporary file: %v", err)
	}

	gw, err := New(tmp.Name())
	if err != nil {
		t.Fatalf("failed to create gateway: %v", err)
	}

	route, ok := gw.match("/users")
	assert.True(t, ok)
	assert.Equal(t, "users", route.Name)

	// Case 2: no match
	route, ok = gw.match("/payments")
	assert.False(t, ok)
	assert.Nil(t, route)

	// Case 3: prefix match
	yamlData = `
routes:
  - name: users
    paths: ["/users/*"]
    backend: "localhost:8080"
`
	tmp, err = os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(yamlData); err != nil {
		t.Fatalf("failed to write to temporary file: %v", err)
	}

	gw, err = New(tmp.Name())
	if err != nil {
		t.Fatalf("failed to create gateway: %v", err)
	}

	route, ok = gw.match("/users/123")
	assert.True(t, ok)
	assert.Equal(t, "users", route.Name)

	// Case 4: multiple routes
	yamlData = `
routes:
  - name: users
    paths: ["/users", "/payments"]
    backend: "localhost:8080"
`
	tmp, err = os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(yamlData); err != nil {
		t.Fatalf("failed to write to temporary file: %v", err)
	}

	gw, err = New(tmp.Name())
	if err != nil {
		t.Fatalf("failed to create gateway: %v", err)
	}

	route, ok = gw.match("/users")
	assert.True(t, ok)
	assert.Equal(t, "users", route.Name)

	route, ok = gw.match("/payments")
	assert.True(t, ok)
	assert.Equal(t, "users", route.Name)
}
