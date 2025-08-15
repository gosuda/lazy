package lazy

// Map applies the mapper function to each value in the stream and returns a new stream.
//
//	Options:
//	- WithContext(context.Context): set the context for cancellation. When the context is done, the stream will be closed immediately.
//	- WithSize(int): the size of the new stream created by this function. Will prepare some items in advance. (default: 0)
//	- WithErrorHandler(func(error)): specify how to handle errors. If the error handler returns DecisionStop, the stream will be closed immediately. (default: IgnoreErrorHandler)
func Map[IN any, OUT any](obj stream[IN], mapper func(v IN) (OUT, error), opts ...optionFunc) stream[OUT] {
	opt := buildOpts(opts)
	ch := make(chan OUT, opt.size)
	ctx := opt.ctx

	go func() {
		defer recover()
		defer close(ch)
		for v := range obj.ch {
			result, err := mapper(v)
			if err != nil {
				if decision := opt.onError(err); decision == DecisionStop {
					return
				}
				continue
			}
			select {
			case <-ctx.Done():
				return
			case ch <- result:
			}
		}
	}()

	return New(ch)
}
