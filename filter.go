package lazy

import "fmt"

type filterFunc[T any] func(v T) (bool, error)

func Filter[T any](stream reader[T], filter filterFunc[T]) reader[T] {
	r, w := New[T]()

	go func() {
		defer w.Close()
		for v := range stream.ch {
			ok, err := filter(v)
			if err != nil {
				stream.Close(fmt.Errorf("error from filter function: %w", err))
				return
			}
			if ok {
				if err := w.Emit(v); err != nil {
					stream.Close(err)
					return
				}
			}
		}
	}()

	return r
}
