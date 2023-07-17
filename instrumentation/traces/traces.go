package traces

import (
	"context"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/instrumentation"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var provider *sdktrace.TracerProvider

func RegisterTracesPipeline() error {
	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(viper.GetString(config.Instrumentation_Traces_Endpoint)),
	}

	if viper.GetBool(config.Instrumentation_Traces_Insecure) {
		options = append(options, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(
		context.Background(),
		options...,
	)
	if err != nil {
		return err
	}

	resource, err := instrumentation.Resource()
	if err != nil {
		return err
	}

	provider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return nil
}

func Shutdown(ctx context.Context) error {
	if provider != nil {
		return provider.Shutdown(ctx)
	}

	return nil
}

func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("github.com/alfreddobradi/empires-of-avalon/instrumentation").Start(ctx, spanName, opts...)
}
