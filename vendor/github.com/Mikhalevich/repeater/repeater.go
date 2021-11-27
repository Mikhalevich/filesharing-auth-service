package repeater

import (
	"time"
)

type Logger interface {
	Infof(format string, args ...interface{})
}

type options struct {
	count      int
	timeout    time.Duration
	l          Logger
	logMessage string
}

type Option func(opts *options)

func WithCount(c int) Option {
	return func(opts *options) {
		opts.count = c
	}
}

func WithTimeout(d time.Duration) Option {
	return func(opts *options) {
		opts.timeout = d
	}
}

func WithLogger(l Logger) Option {
	return func(opts *options) {
		opts.l = l
	}
}

func WithLogMessage(m string) Option {
	return func(opts *options) {
		opts.logMessage = m
	}
}

func Do(fn func() error, opts ...Option) error {
	params := &options{
		count: 3,
	}

	for _, o := range opts {
		o(params)
	}

	var err error
	for i := 0; i < params.count; i++ {
		err = fn()
		if err == nil {
			break
		}

		if params.l != nil {
			if params.logMessage != "" {
				params.l.Infof("message: %s, repeating attempt: %d, err: %v\n", params.logMessage, i+1, err)
			} else {
				params.l.Infof("repeating attempt: %d, err: %v\n", i+1, err)
			}
		}

		if params.timeout > 0 {
			time.Sleep(params.timeout)
		}
	}

	return err
}
