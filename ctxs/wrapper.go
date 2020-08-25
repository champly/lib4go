package ctxs

import "context"

func Get(ctx context.Context, key interface{}) interface{} {
	return ctx.Value(key)
}

func WithValue(parent context.Context, key interface{}, value interface{}) context.Context {
	return context.WithValue(parent, key, value)
}

func Clone(parent context.Context) context.Context {
	if v, ok := parent.(*valueCtx); ok {
		clone := &valueCtx{Context: v}
		return clone
	}

	return parent
}
