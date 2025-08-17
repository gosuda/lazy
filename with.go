package lazy

import "context"

type option struct {
	ctx      context.Context
	size     int
	parallel int
	onError  errHandlerFunc
}

type optionFunc func(opts *option)

func buildOpts(opts []optionFunc) option {
	opt := option{
		ctx:      context.Background(),
		size:     0,
		parallel: 1,
		onError:  IgnoreErrorHandler,
	}
	for _, f := range opts {
		f(&opt)
	}
	return opt
}

func WithSize(size int) optionFunc {
	return func(opts *option) {
		opts.size = size
	}
}

func WithContext(ctx context.Context) optionFunc {
	return func(opts *option) {
		opts.ctx = ctx
	}
}

type Decision string

const (
	DecisionStop   = "stop"
	DecisionIgnore = "ignore"
)

type errHandlerFunc func(err error) Decision

var (
	IgnoreErrorHandler errHandlerFunc = func(err error) Decision {
		return DecisionIgnore
	}
)

func WithErrHandler(handler errHandlerFunc) optionFunc {
	return func(opts *option) {
		opts.onError = handler
	}
}

func WithParallelism(parallel int) optionFunc {
	return func(opts *option) {
		opts.parallel = parallel
	}
}
