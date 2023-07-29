package instrumentation

import (
	"github.com/0xa1-red/empires-of-avalon/version"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func Resource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("avalond"),
			semconv.ServiceVersion(version.Tag),
		),
	)
}
