package tracer

import (
	"context"

	"cloud.google.com/go/compute/metadata"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/status"
)

var tracer = otel.GetTracerProvider().Tracer("nbn23.com/trace")

func Init(name string) error {
	projectId, err := metadata.ProjectID()

	if err != nil {
		return nil
	}

	// Create exporter.
	exporter, err := texporter.New(texporter.WithProjectID(projectId))

	if err != nil {
		return err
	}

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(semconv.ServiceNameKey.String(name)),
	)

	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return nil
}

func Shutdown() {
	ctx := context.Background()

	tp := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	tp.Shutdown(ctx)
}

type Trace struct {
	span trace.Span
}

func Start(name string) Trace {
	ctx := context.Background()

	_, span := tracer.Start(ctx, name)

	return Trace{span}
}

func (tr Trace) SetAttributes(values map[string]string) {
	attrs := []attribute.KeyValue{}

	for key, value := range values {
		attr := attribute.KeyValue{
			Key:   attribute.Key(key),
			Value: attribute.StringValue(value),
		}

		attrs = append(attrs, attr)
	}

	tr.span.SetAttributes(attrs...)
}

func (tr Trace) End(err error) {
	defer tr.span.End()

	if err == nil {
		tr.span.SetStatus(codes.Ok, codes.Ok.String())

		return
	}

	status := status.Convert(err)
	code := runtime.HTTPStatusFromCode(status.Code())

	tr.span.RecordError(err)
	tr.span.SetStatus(codes.Code(code), err.Error())
}
