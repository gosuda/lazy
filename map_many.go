package lazy

type mapManyFunc[IN, OUT any] func(ctx Context, v <-chan IN, yield func(OUT) error) error

func MapMany[IN, OUT any](stream reader[IN], f mapManyFunc[IN, OUT], opts ...optionFunc) reader[OUT] {
	opts = append(opts, withFname("MapMany"))
	opt := buildOpts(opts)
	r, w := New[OUT](opts...)

	go func() {
		defer w.Close()
		ctx := &context{Context: opt.ctx, name: opt.name, workerIdx: 0}
		if err := f(ctx, stream.ch, w.Emit); err != nil {
			stream.Close(err)
		}
	}()

	return r
}
