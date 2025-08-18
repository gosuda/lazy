package lazy

import (
	"errors"
	"sync/atomic"
)

type reader[T any] struct {
	ch        <-chan T
	propagate func(err error)
	ctx       Context
}

func (r *reader[T]) Close(reason error) {
	r.propagate(reason)
}

type writer[T any] struct {
	ch chan T

	// reason is the error that caused the stream to be closed.
	reason *atomic.Pointer[error]
	ctx    Context
}

func (e *writer[T]) Emit(v T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = *e.reason.Load()
		}
	}()
	if e.reason.Load() != nil {
		return *e.reason.Load()
	}
	select {
	case <-e.ctx.Done():
		err := e.ctx.Err()
		e.reason.Store(&err)
		return err
	case e.ch <- v:
	}
	return nil
}

func (e *writer[T]) Len() int {
	return len(e.ch)
}

func (e *writer[T]) Close() {
	// possibly panic if the stream is already closed
	defer func() { recover() }()

	if e.reason.Load() != nil {
		close(e.ch)
	} else {
		err := errors.New("stream is closed")
		e.reason.Store(&err)
		close(e.ch)
	}
}
