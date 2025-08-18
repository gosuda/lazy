package lazy

import (
	"github.com/gosuda/lazy/errgroup"
)

type consumeFunc[IN any] func(ctx Context, v IN) error

func Consume[IN any](stream reader[IN], consumer consumeFunc[IN], opts ...optionFunc) error {
	opts = append(opts, withFname("Consume"))
	opt := buildOpts(opts)

	eg, ctx := errgroup.WithContext(opt.ctx)

	for i := 0; i < opt.parallel; i++ {
		eg.Go(func() error {
			ctx := &context{Context: ctx, name: opt.name, workerIdx: i}
			for v := range stream.ch {
				if err := consumer(ctx, v); err != nil {
					stream.Close(err)
					return err
				}
			}
			return nil
		})
	}
	return eg.Wait()
}
