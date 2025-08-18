package lazy

import (
	stdContext "context"
)

var _ stdContext.Context = &context{}

type Context interface {
	stdContext.Context
	Name() string
	WorkerIdx() int
}

// Context is a wrapper around context.Context that implements the context.Context interface.
// It provides contextual information of the lazy component such as the stream size, the number of workers, etc.
type context struct {
	stdContext.Context

	name      string
	workerIdx int
}

func (c *context) Name() string {
	return c.name
}

func (c *context) WorkerIdx() int {
	return c.workerIdx
}
