package metrics

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/0xa1-red/empires-of-avalon/pkg/instrumentation"
	"github.com/0xa1-red/empires-of-avalon/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var server *http.Server

const (
	instrumentationName string = "github.com/alfreddobradi/empires-of-avalon/instrumentation"
)

func RegisterMetricsPipeline() error {
	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := prometheus.New(prometheus.WithNamespace("avalond"))
	if err != nil {
		log.Fatal(err)
	}

	resource, err := instrumentation.Resource()
	if err != nil {
		return err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(resource),
	)
	otel.SetMeterProvider(provider)

	return nil
}

func ServeMetrics(wg *sync.WaitGroup) {
	defer wg.Done()

	slog.Info("starting promhttp server", "address", "0.0.0.0:2223")

	mux := &http.ServeMux{}
	mux.Handle("/metrics", promhttp.Handler())

	server = &http.Server{ // nolint:exhaustruct
		Addr:         ":2223",
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		Handler:      mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("http server error", err)
	}
}

func Shutdown(ctx context.Context) error {
	if server != nil {
		slog.Debug("shutting down metrics server")
		return server.Shutdown(ctx)
	}

	return nil
}

func Meter() metric.Meter {
	return otel.Meter(
		instrumentationName, metric.WithInstrumentationVersion(version.Tag))
}
