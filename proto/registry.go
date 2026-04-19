package proto

import (
	"fmt"
	"sync"
)

// MethodInfo holds metadata about a registered gRPC method.
type MethodInfo struct {
	FullMethod string
	InputType  string
	OutputType string
}

// Registry stores known gRPC method descriptors.
type Registry struct {
	mu      sync.RWMutex
	methods map[string]MethodInfo
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{methods: make(map[string]MethodInfo)}
}

// Register adds a method to the registry.
func (r *Registry) Register(info MethodInfo) error {
	if info.FullMethod == "" {
		return fmt.Errorf("proto: FullMethod must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.methods[info.FullMethod] = info
	return nil
}

// Lookup retrieves a method by its full name.
func (r *Registry) Lookup(fullMethod string) (MethodInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.methods[fullMethod]
	return info, ok
}

// List returns all registered method names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.methods))
	for k := range r.methods {
		names = append(names, k)
	}
	return names
}
