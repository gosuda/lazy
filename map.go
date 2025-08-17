package lazy

type mapFunc[IN, OUT any] func(v IN) (OUT, error)

func Map[IN any, OUT any](stream reader[IN], mapper mapFunc[IN, OUT], opts ...optionFunc) reader[OUT] {
	r, w := New[OUT](opts...)
	opt := buildOpts(opts)

	for i := 0; i < opt.parallel; i++ {
		go func() {
			defer w.Close()
			for v := range stream.ch {
				result, err := mapper(v)
				if err != nil {
					if opt.onError(err) == DecisionStop {
						stream.Close(err)
						return
					} else {
						continue
					}
				}
				if err := w.Emit(result); err != nil {
					stream.Close(err)
					return
				}
			}
		}()
	}
	return r
}
