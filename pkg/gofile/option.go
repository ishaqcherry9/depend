package gofile

const (
	prefix  = "prefix"
	suffix  = "suffix"
	contain = "contain"
)

var (
	defaultFilterType = ""
)

type options struct {
	filter string
	name   string

	noAbsolutePath bool
}

func defaultOptions() *options {
	return &options{
		filter: defaultFilterType,
	}
}

type Option func(*options)

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithSuffix(name string) Option {
	return func(o *options) {
		o.filter = suffix
		o.name = name
	}
}

func WithPrefix(name string) Option {
	return func(o *options) {
		o.filter = prefix
		o.name = name
	}
}

func WithContain(name string) Option {
	return func(o *options) {
		o.filter = contain
		o.name = name
	}
}

func WithNoAbsolutePath() Option {
	return func(o *options) {
		o.noAbsolutePath = true
	}
}
