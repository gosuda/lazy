package lazy

import (
	"context"
)

var _ context.Context = &Context{}

// Context is a wrapper around context.Context that implements the context.Context interface.
// It provides contextual information of the lazy component such as the stream size, the number of workers, etc.
type Context struct {
	context.Context

	name      string
	workerIdx int
}

func (c *Context) Name() string {
	return c.name
}

func (c *Context) WorkerIdx() int {
	return c.workerIdx
}
