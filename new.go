package lazy

type stream[T any] struct {
	ch <-chan T
}

// New wraps a user-provided channel as a lazy stream.
// It does not close the channel, so the caller must ensure it is closed when done.
// Even the channel is closed, it will process the remaining values in the channel.
func New[T any](in <-chan T) stream[T] {
	return stream[T]{ch: in}
}

// NewSlice creates a lazy stream from a slice.
// It will close the channel when the context is done, or when the slice is exhausted.
//
//	Available options:
//	  WithContext(context.Context): set the context for cancellation. When the context is done, the stream will be closed immediately.
func NewSlice[T any](slice []T, opts ...optionFunc) stream[T] {
	opt := buildOpts(opts)
	ch := make(chan T)
	ctx := opt.ctx
	go func() {
		defer recover()
		defer close(ch)
		for _, v := range slice {
			select {
			case <-ctx.Done():
				return
			case ch <- v:
			}
		}
	}()
	return New(ch)
}
