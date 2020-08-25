package ctxs

import "context"

type valueCtx struct {
	context.Context
}

func (c *valueCtx) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}
