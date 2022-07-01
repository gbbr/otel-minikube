package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.opentelemetry.io/contrib/detectors/aws/ec2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var grpcDriver = flag.Bool("grpc", false, "use GRPC exporter instead of HTTP")

func bootstrap() (tracer trace.Tracer, stop func()) {
	flag.Parse()
	ctx := context.Background()
	headers := map[string]string{"Custom-Header": "Custom-Value"}
	var (
		exporter *otlptrace.Exporter
		err      error
	)
	if *grpcDriver {
		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint("0.0.0.0:4317"),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithHeaders(headers),
		)
		exporter, err = otlptrace.New(ctx, client)
		if err != nil {
			log.Fatalf("failed to create exporter: %v", err)
		}
	} else {
		exporter, err = otlptracehttp.New(
			ctx,
			otlptracehttp.WithEndpoint("0.0.0.0:4318"),
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithHeaders(headers),
		)
		if err != nil {
			log.Fatalf("Failed to create HTTP client: %v", err)
		}
	}
	res1, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithDetectors(),
		resource.WithContainer(),
		resource.WithDetectors(ec2.NewResourceDetector()),
	)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}
	resource, err := resource.Merge(
		resource.NewSchemaless(
			attribute.String("service.name", "my-service"),
			attribute.String("service.version", "1.2.3"),
			attribute.String("env", "staging"),
		),
		res1,
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(resource),
	)
	//otel.SetTracerProvider(provider)
	return provider.Tracer("mytracer", trace.WithInstrumentationVersion("1.0.0")), func() {
		if err := exporter.Shutdown(ctx); err != nil {
			log.Fatalf("failed to stop exporter: %v", err)
		}
	}
}

func main() {
	ctx := context.Background()
	tracer, stop := bootstrap()
	defer stop()

	var n int
	for {
		n++
		ctx1, span := tracer.Start(ctx, fmt.Sprintf("operation%d", n), trace.WithAttributes(
			attribute.Bool("mybool", false),
			attribute.Int("myint", 1),
			attribute.Float64("myfloat64", 2),
			attribute.String("mystring", "asd"),
		))
		for i := 0; i < 5; i++ {
			_, span := tracer.Start(ctx1, fmt.Sprintf("child-%d", i), trace.WithAttributes(attribute.Int("x", i)))
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			span.End()
		}
		span.End()

		ctx2, span := tracer.Start(ctx, fmt.Sprintf("operation-mop%d", n), trace.WithAttributes(
			attribute.Bool("mybool", false),
			attribute.Int("myint", 1),
			attribute.Float64("myfloat64", 2),
			attribute.String("mystring", "asd"),
		))
		for i := 0; i < 5; i++ {
			_, span := tracer.Start(ctx2, fmt.Sprintf("mop-child-%d", i), trace.WithAttributes(attribute.Int("x", i)))
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			span.End()
		}
		span.End()
		fmt.Print(".")
	}
}
