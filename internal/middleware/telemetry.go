package middleware

import (
	"context"

	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DatadogTracer struct{}

func NewDatadogTracer() *DatadogTracer {
	ddtrace.Start()
	return &DatadogTracer{}
}

func (d *DatadogTracer) StartSpan(ctx context.Context, name string) context.Context {
	_, newCtx := ddtrace.StartSpanFromContext(ctx, name)
	return newCtx
}

func (d *DatadogTracer) FinishSpan(ctx context.Context) {
	span, ok := ddtrace.SpanFromContext(ctx)
	if ok {
		span.Finish()
	}
}
