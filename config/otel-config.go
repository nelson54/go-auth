package config

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var serviceName string
var metricNamespace string
var globalHandler func(handler http.HandlerFunc) http.HandlerFunc
var otelShutdown func()

func CreateHandler(handler http.HandlerFunc, url string) http.Handler {
	return otelhttp.NewHandler(globalHandler(handler), url)
}

func OtelShutDown() {
	otelShutdown()
}

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initProvider() func() {

	if serviceNameVar, ok := os.LookupEnv("GOLANG_SERVICE_NAME"); !ok {
		serviceName = "golang-app"
	} else {
		serviceName = serviceNameVar
	}

	if metricNamespaceVar, ok := os.LookupEnv("GOLANG_SERVICE_NAME"); !ok {
		metricNamespace = "golang_app"
	} else {
		metricNamespace = metricNamespaceVar
	}

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	handleErr(err, "failed to create resource")

	otelAgentAddr, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if !ok {
		otelAgentAddr = "0.0.0.0:4317"
	}

	metricExp, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(otelAgentAddr))
	handleErr(err, "Failed to create the collector metric exporter")

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				metricExp,
				sdkmetric.WithInterval(2*time.Second),
			),
		),
	)
	otel.SetMeterProvider(meterProvider)

	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otelAgentAddr))
	traceExp, err := otlptrace.New(ctx, traceClient)
	handleErr(err, "Failed to create the collector trace exporter")

	bsp := sdktrace.NewBatchSpanProcessor(traceExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracerProvider)

	err = host.Start(host.WithMeterProvider(meterProvider))
	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := traceExp.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
		// pushes any last exports to the receiver
		if err := meterProvider.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

func OtelContrib() {
	otelShutdown = initProvider()

	meter := otel.Meter(serviceName)
	serverAttribute := attribute.String("service", serviceName)
	commonLabels := []attribute.KeyValue{serverAttribute}

	requestCountMetricName := fmt.Sprintf("%s/request_counts", metricNamespace)
	requestCount, _ := meter.Int64Counter(
		requestCountMetricName,
		metric.WithDescription("The number of requests received"),
	)

	// create a handler wrapped in OpenTelemetry instrumentation
	globalHandler = func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			requestCount.Add(ctx, 1, metric.WithAttributes(commonLabels...))
			span := trace.SpanFromContext(ctx)
			bag := baggage.FromContext(ctx)
			var baggageAttributes []attribute.KeyValue
			baggageAttributes = append(baggageAttributes, serverAttribute)
			for _, member := range bag.Members() {
				baggageAttributes = append(baggageAttributes, attribute.String("baggage key:"+member.Key(), member.Value()))
			}
			span.SetAttributes(baggageAttributes...)

			handler(w, req)
		}
	}
}
