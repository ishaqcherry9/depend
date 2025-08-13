package tracer

import (
	"io"
	"os"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

func NewConsoleExporter() (sdkTrace.SpanExporter, error) {
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

func NewFileExporter(filename string) (sdkTrace.SpanExporter, *os.File, error) {
	if filename == "" {
		filename = "traces.json"
	}

	f, err := os.Create(filename)
	if err != nil {
		panic("os.Create error: " + err.Error())
	}

	exporter, err := newExporter(f)
	if err != nil {
		panic("newExporter error: " + err.Error())
	}

	return exporter, f, nil
}

func newExporter(w io.Writer) (sdkTrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),

		stdouttrace.WithPrettyPrint(),

		stdouttrace.WithoutTimestamps(),
	)
}
