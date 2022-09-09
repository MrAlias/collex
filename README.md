# collex

Use [OpenTelemetry Collector] Exporters with [OpenTelemetry Go]

## Getting Started

Pass your `collex` wrapped exporter to [OpenTelemetry Go] providers

### Tracing

```go
// collectorExporter is assumed to be fully configured and started.
tracerProvider := trace.NewTracerProvider(
	trace.WithBatcher(collex.TracesExporter(collectorExporter)),
	/* ... */
)
```

[OpenTelemetry Collector]: https://github.com/open-telemetry/opentelemetry-collector
[OpenTelemetry Go]: https://github.com/open-telemetry/opentelemetry-go
