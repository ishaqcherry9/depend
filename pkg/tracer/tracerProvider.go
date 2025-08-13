package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

var tp *trace.TracerProvider

func Init(exporter trace.SpanExporter, res *resource.Resource, fractions ...float64) {
	var fraction = 1.0
	if len(fractions) > 0 {
		if fractions[0] <= 0 {
			fraction = 0
		} else if fractions[0] < 1 {
			fraction = fractions[0]
		}
	}

	tp = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(fraction))),
	)

	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func Close(ctx context.Context) error {
	if tp == nil {
		return nil
	}
	return tp.Shutdown(ctx)
}

func InitWithConfig(appName string, appEnv string, appVersion string,
	jaegerAgentHost string, jaegerAgentPort string, jaegerSamplingRate float64) {
	res := NewResource(
		WithServiceName(appName),
		WithEnvironment(appEnv),
		WithServiceVersion(appVersion),
	)

	exporter, err := NewJaegerAgentExporter(jaegerAgentHost, jaegerAgentPort)
	if err != nil {
		panic("init trace error:" + err.Error())
	}

	Init(exporter, res, jaegerSamplingRate)

	SetTraceName(appName)
}

func GetProvider() *trace.TracerProvider {
	if tp == nil {
		panic("tracer provider is nil, initialize it first with InitWithConfig(...)")
	}
	return tp
}
