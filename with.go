package lazy

import (
	stdContext "context"
	"fmt"
	"runtime"
)

type option struct {
	ctx      *context
	fname    string
	name     string
	size     int
	parallel int
	onError  errHandlerFunc
}

type optionFunc func(opts *option)

func buildOpts(opts []optionFunc) option {
	opt := option{
		ctx:      &context{Context: stdContext.Background()},
		fname:    "",
		name:     "",
		size:     0,
		parallel: 1,
		onError:  IgnoreErrorHandler,
	}
	for _, f := range opts {
		f(&opt)
	}
	if opt.name == "" {
		pc, _, line, _ := runtime.Caller(2)
		f := runtime.FuncForPC(pc)
		opt.name = fmt.Sprintf("%s:%d", f.Name(), line)
	}

	if opt.fname == "" {
		opt.fname = "New"
	}

	opt.name = fmt.Sprintf("%s(%s)", opt.fname, opt.name)
	return opt
}

func withFname(fname string) optionFunc {
	return func(opts *option) {
		opts.fname = fname
	}
}

func WithName(name string) optionFunc {
	return func(opts *option) {
		opts.name = name
	}
}

func WithSize(size int) optionFunc {
	return func(opts *option) {
		opts.size = size
	}
}

func WithContext(ctx stdContext.Context) optionFunc {
	return func(opts *option) {
		opts.ctx = &context{Context: ctx}
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
