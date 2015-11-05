package lib

import (
	"golang.org/x/net/context"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// pathKey is the context key for a path.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const pathKey key = 0

const traceKey = 1

// NewContext returns a new Context carrying an Identity pat.
func NewContext(ctx context.Context, p string) context.Context {
	return context.WithValue(ctx, pathKey, p)
}

// FromContext extracts the Identity pat from ctx, if present.
func FromContext(ctx context.Context) (string, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	p, ok := ctx.Value(pathKey).(string)
	return p, ok
}

// MustFromContext extracts the identity from ctx.
// If not present it panics.
func MustFromContext(ctx context.Context) string {
	idt, ok := ctx.Value(pathKey).(string)
	if !ok {
		panic("path is not registered")
	}
	return idt
}

// NewTraceContext returns a new Context carrying an Identity pat.
func NewTraceContext(ctx context.Context, trace string) context.Context {
	return context.WithValue(ctx, traceKey, trace)
}

// FromTraceContext extracts the Identity pat from ctx, if present.
func FromTraceContext(ctx context.Context) (string, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	t, ok := ctx.Value(traceKey).(string)
	return t, ok
}

// MustFromTraceContext extracts the identity from ctx.
// If not present it panics.
func MustFromTraceContext(ctx context.Context) string {
	t, ok := ctx.Value(traceKey).(string)
	if !ok {
		panic("trace is not registered")
	}
	return t
}
