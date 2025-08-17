package lazy

type filterFunc[T any] func(v T) (bool, error)

func Filter[T any](stream reader[T], filter filterFunc[T], opts ...optionFunc) reader[T] {
	r, w := New[T](opts...)
	opt := buildOpts(opts)

	for i := 0; i < opt.parallel; i++ {
		go func() {
			defer w.Close()
			for v := range stream.ch {
				ok, err := filter(v)
				if err != nil {
					if opt.onError(err) == DecisionStop {
						stream.Close(err)
						return
					} else {
						continue
					}
				}
				if ok {
					if err := w.Emit(v); err != nil {
						stream.Close(err)
						return
					}
				}
			}
		}()
	}

	return r
}
