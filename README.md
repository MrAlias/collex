# collex

Use [OpenTelemetry Collector] Exporters with [OpenTelemetry Go]

:construction: This project is still a work in progress.
Breaking changes may be introduced.

## Getting Started

OpenTelemetry Collector exporters are generated from [ExporterFactory]s.
First wrap this facotry with `collex`.

```go
factory, err := collex.NewFactory(your.NewFactory(), nil)
if err != nil {
    // Handle error appropiately.
}
```

### Tracing

Generate a [SpanExporter] from your `collex.Factory`.

```go
exp, err := factory.SpanExporter(context.Background(), nil)
if err != nil {
    // Handle error appropiately.
}
provider := trace.NewTracerProvider(trace.WithBatcher(exp))
```

Use `provider` as any other OpenTelemetry Go [TracerProvider] to generate tracing telemetry.

[OpenTelemetry Collector]: https://github.com/open-telemetry/opentelemetry-collector
[OpenTelemetry Go]: https://github.com/open-telemetry/opentelemetry-go
[ExporterFactory]: https://pkg.go.dev/go.opentelemetry.io/collector@v0.60.0/component#ExporterFactory
[SpanExporter]: https://pkg.go.dev/go.opentelemetry.io/otel/sdk@v1.10.0/trace#SpanExporter
[TracerProvider]: https://pkg.go.dev/go.opentelemetry.io/otel/sdk@v1.10.0/trace#TracerProvider
