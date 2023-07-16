package metrics

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

var server *http.Server

func RegisterMetricsPipeline() {
	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(provider)
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
		return server.Shutdown(ctx)
	}

	return nil
}
