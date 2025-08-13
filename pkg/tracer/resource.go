package tracer

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type resourceOptions = resourceConfig

type ResourceOption interface {
	apply(*resourceOptions)
}

type resourceOptionFunc func(*resourceOptions)

func (o resourceOptionFunc) apply(cfg *resourceOptions) {
	o(cfg)
}

func apply(obj *resourceOptions, opts ...ResourceOption) {
	for _, opt := range opts {
		opt.apply(obj)
	}
}

func WithServiceName(name string) ResourceOption {
	return resourceOptionFunc(func(o *resourceOptions) {
		o.serviceName = name
	})
}

func WithServiceVersion(version string) ResourceOption {
	return resourceOptionFunc(func(o *resourceOptions) {
		o.serviceVersion = version
	})
}

func WithEnvironment(environment string) ResourceOption {
	return resourceOptionFunc(func(o *resourceOptions) {
		o.environment = environment
	})
}

func WithAttributes(attributes map[string]string) ResourceOption {
	return resourceOptionFunc(func(o *resourceOptions) {
		o.attributes = attributes
	})
}

type resourceConfig struct {
	serviceName    string
	serviceVersion string
	environment    string

	attributes map[string]string
}

func NewResource(opts ...ResourceOption) *resource.Resource {

	rc := &resourceConfig{
		serviceName:    "demo-service",
		serviceVersion: "v0.0.0",
		environment:    "dev",
	}
	apply(rc, opts...)

	kvs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(rc.serviceName),
		semconv.ServiceVersionKey.String(rc.serviceVersion),
		attribute.String("env", rc.environment),
	}
	for k, v := range rc.attributes {
		kvs = append(kvs, attribute.String(k, v))
	}

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, kvs...),
	)
	if err != nil {
		panic(err)
	}

	return r
}
