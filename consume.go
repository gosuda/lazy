package lazy

type consumeFunc[IN any] func(ctx *Context, v IN) error

func Consume[IN any](stream reader[IN], consumer consumeFunc[IN], opts ...optionFunc) error {
	opts = append(opts, withFname("Consume"))
	opt := buildOpts(opts)

	for i := 0; i < opt.parallel; i++ {
		go func() {
			ctx := &Context{Context: opt.ctx, name: opt.name, workerIdx: i}
			for v := range stream.ch {
				if err := consumer(ctx, v); err != nil {
					stream.Close(err)
					return
				}
			}
		}()
	}
	return nil
}
