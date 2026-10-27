package telemetry

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
)

type shutdownFunc func(context.Context) error

type Runtime struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
	propagator     propagation.TextMapPropagator
	resource       *resource.Resource
	failures       *failureReporter
	shutdownTrace  shutdownFunc
	shutdownMetric shutdownFunc
	shutdownOnce   sync.Once
	shutdownErr    error
}

type FailureCounts struct {
	Trace  uint64
	Metric uint64
}

func (runtime *Runtime) TracerProvider() trace.TracerProvider {
	return runtime.tracerProvider
}

func (runtime *Runtime) MeterProvider() metric.MeterProvider {
	return runtime.meterProvider
}

func (runtime *Runtime) ExportFailures() FailureCounts {
	return runtime.failures.snapshot()
}

func (runtime *Runtime) Shutdown(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("telemetry shutdown context is required")
	}
	runtime.shutdownOnce.Do(func() {
		traceErr := runtime.shutdownTrace(ctx)
		metricErr := runtime.shutdownMetric(ctx)
		if traceErr != nil {
			runtime.failures.report(ctx, signalTrace, operationShutdown)
		}
		if metricErr != nil {
			runtime.failures.report(ctx, signalMetric, operationShutdown)
		}
		if traceErr != nil || metricErr != nil {
			runtime.shutdownErr = fmt.Errorf("telemetry shutdown did not flush every signal")
		}
	})
	return runtime.shutdownErr
}

func assembleRuntime(
	tracerProvider trace.TracerProvider,
	meterProvider metric.MeterProvider,
	res *resource.Resource,
	failures *failureReporter,
	shutdownTrace shutdownFunc,
	shutdownMetric shutdownFunc,
) *Runtime {
	return &Runtime{
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
		propagator:     propagation.TraceContext{},
		resource:       res,
		failures:       failures,
		shutdownTrace:  shutdownTrace,
		shutdownMetric: shutdownMetric,
	}
}
