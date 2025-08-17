package lazy

import "sync/atomic"

// New wraps a user-provided channel as a lazy stream.
// It does not close the channel, so the caller must ensure it is closed when done.
// Even the channel is closed, it will process the remaining values in the channel.
//
//	Available options:
//	// the size of the new stream created by this function. Will prepare some items in advance.
//	WithSize(int) // default: 0
//	// set the context for cancellation. When the context is done, the stream will be closed immediately.
//	WithContext(context.Context) // default: context.Background()
func New[T any](opts ...optionFunc) (reader[T], writer[T]) {
	opt := buildOpts(opts)
	ch := make(chan T, opt.size)
	ctx := opt.ctx

	w := writer[T]{
		ch:     ch,
		ctx:    ctx,
		reason: &atomic.Pointer[error]{},
	}
	r := reader[T]{
		ch: ch,
		propagate: func(err error) {
			w.reason.Store(&err)
			w.Close()
		},
	}

	return r, w
}

// NewSlice creates a lazy stream from a slice.
// It will close the channel when the context is done, or when the slice is exhausted.
//
//	Available options:
//	// the size of the new stream created by this function. Will prepare some items in advance.
//	WithSize(int) // default: 0
//	// set the context for cancellation. When the context is done, the stream will be closed immediately.
//	WithContext(context.Context) // default: context.Background()
func NewSlice[T any](slice []T, opts ...optionFunc) reader[T] {
	r, w := New[T](opts...)

	go func() {
		defer w.Close()
		for _, v := range slice {
			if err := w.Emit(v); err != nil {
				return
			}
		}
	}()
	return r
}
