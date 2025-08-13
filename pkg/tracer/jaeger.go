package tracer

import (
	"go.opentelemetry.io/otel/exporters/jaeger"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

type JaegerOption func(*jaegerOptions)

type jaegerOptions struct {
	username string
	password string
}

func (o *jaegerOptions) apply(opts ...JaegerOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func defaultJaegerOptions() *jaegerOptions {
	return &jaegerOptions{}
}

func WithUsername(username string) JaegerOption {
	return func(o *jaegerOptions) {
		o.username = username
	}
}

func WithPassword(password string) JaegerOption {
	return func(o *jaegerOptions) {
		o.password = password
	}
}

func NewJaegerExporter(url string, opts ...JaegerOption) (sdkTrace.SpanExporter, error) {
	ceps := []jaeger.CollectorEndpointOption{
		jaeger.WithEndpoint(url),
	}

	o := defaultJaegerOptions()
	o.apply(opts...)
	if o.username != "" {
		ceps = append(ceps, jaeger.WithUsername(o.username))
	}
	if o.password != "" {
		ceps = append(ceps, jaeger.WithPassword(o.password))
	}

	endpointOption := jaeger.WithCollectorEndpoint(ceps...)

	return jaeger.New(endpointOption)
}

func NewJaegerAgentExporter(host string, port string) (sdkTrace.SpanExporter, error) {
	return jaeger.New(
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(host),
			jaeger.WithAgentPort(port),
		),
	)
}
