# go-micro-opentelemetry

## usage

```go
url := "http://127.0.0.1:14268/api/traces"
exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
tp := tracesdk.NewTracerProvider(
    tracesdk.WithBatcher(exporter),
    tracesdk.WithResource(resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String(serviceName),
        attribute.String("environment", environment),
    )),
)

otel.SetTracerProvider(tp)
otel.SetTextMapPropagator(propagation.TraceContext{})


//add a wrapper
micro.WrapHandler(oteltracing.NewHandlerWrapper())
```