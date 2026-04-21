package tag

import (
	"context"
)

type contextKey struct{}

// Attach stores a Bag in the context, returning the derived context.
func Attach(ctx context.Context, b *Bag) context.Context {
	return context.WithValue(ctx, contextKey{}, b)
}

// FromContext retrieves the Bag stored in ctx.
// If no Bag is present, a new empty Bag is returned so callers never
// need to nil-check.
func FromContext(ctx context.Context) *Bag {
	if b, ok := ctx.Value(contextKey{}).(*Bag); ok && b != nil {
		return b
	}
	return New()
}

// SetInContext is a convenience helper that attaches key/value to the Bag
// stored in ctx. If no Bag exists one is created and the new context is
// returned alongside it.
func SetInContext(ctx context.Context, key, value string) (context.Context, *Bag) {
	b := FromContext(ctx)
	b.Set(key, value)
	return Attach(ctx, b), b
}
