package trace

import (
	"exapp-go/config"

	uuid "github.com/satori/go.uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func exporter() (sdktrace.SpanExporter, error) {
	traceConf := config.Conf().Trace
	if traceConf.Exporter == "jaeger" {
		return jaeger.New(jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(traceConf.JaegerEndpoint),
		))
	}

	return stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
}

func Init(serviceName string) error {
	exporter, err := exporter()
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNamespaceKey.String("exapp-go"),
			semconv.ServiceInstanceIDKey.String(uuid.NewV4().String()),
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("v1.0.0"),
		)),
	)
	otel.SetTracerProvider(tp)

	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)
	return nil
}
