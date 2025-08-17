package lazy

import "context"

type mapManyFunc[IN, OUT any] func(ctx context.Context, v <-chan IN, yield func(OUT) error) error

func MapMany[IN, OUT any](stream reader[IN], f mapManyFunc[IN, OUT], opts ...optionFunc) reader[OUT] {
	opt := buildOpts(opts)
	r, w := New[OUT](opts...)
	ctx := opt.ctx

	go func() {
		defer w.Close()
		if err := f(ctx, stream.ch, w.Emit); err != nil {
			stream.Close(err)
		}
	}()

	return r
}
