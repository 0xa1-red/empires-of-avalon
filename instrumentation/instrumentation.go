package instrumentation

import (
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func Resource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("avalond"),
			semconv.ServiceVersion("0.10.2"), // Replace this with dynamic version after EOA-37
		),
	)
}
