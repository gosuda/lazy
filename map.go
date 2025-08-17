package lazy

type mapFunc[IN, OUT any] func(ctx *Context, v IN) (OUT, error)

func Map[IN any, OUT any](stream reader[IN], mapper mapFunc[IN, OUT], opts ...optionFunc) reader[OUT] {
	opts = append(opts, withFname("Map"))
	r, w := New[OUT](opts...)
	opt := buildOpts(opts)

	for i := 0; i < opt.parallel; i++ {
		go func() {
			defer w.Close()
			ctx := &Context{Context: opt.ctx, name: opt.name, workerIdx: i}
			for v := range stream.ch {
				result, err := mapper(ctx, v)
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
