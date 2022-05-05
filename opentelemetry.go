package oteltracing

import (
	"context"
	"fmt"

	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/metadata"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type otWrapper struct {
	ot oteltrace.TracerProvider
	client.Client
}

func (o *otWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	name := fmt.Sprintf("call.%s.%s", req.Service(), req.Endpoint())
	ctx, span := o.ot.Tracer("").Start(ctx, name)
	ctx = Inject(ctx)
	defer span.End()
	var err error
	if err = o.Client.Call(ctx, req, rsp, opts...); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (o *otWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	ctx, span := o.ot.Tracer("").Start(ctx, name)
	ctx = Inject(ctx)
	defer span.End()
	stream, err := o.Client.Stream(ctx, req, opts...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return stream, err
}

func (o *otWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	name := fmt.Sprintf("Pub to %s", p.Topic())
	ctx, span := o.ot.Tracer("").Start(ctx, name)
	ctx = Inject(ctx)
	defer span.End()
	var err error
	if err = o.Client.Publish(ctx, p, opts...); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

// NewClientWrapper accepts an open tracing Trace and returns a Client Wrapper
func NewClientWrapper() client.Wrapper {
	return func(c client.Client) client.Client {
		ot := otel.GetTracerProvider()
		return &otWrapper{ot, c}
	}
}

// NewHandlerWrapper accepts an opentracing Tracer and returns a Handler Wrapper
func NewHandlerWrapper() server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			name := fmt.Sprintf("handle.%s.%s", req.Service(), req.Endpoint())
			ctx = Extract(ctx)
			ctx, span := otel.Tracer("").Start(ctx, name)
			defer span.End()
			var err error
			if err = h(ctx, req, rsp); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return err
		}
	}
}

// NewCallWrapper accepts an opentracing Tracer and returns a Call Wrapper
func NewCallWrapper() client.CallWrapper {
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) (err error) {
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			ctx, span := otel.Tracer("").Start(ctx, name)
			defer span.End()
			if err = cf(ctx, node, req, rsp, opts); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return
		}
	}
}

func Inject(ctx context.Context) context.Context {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(metadata.Metadata)
	}
	otel.GetTextMapPropagator().Inject(ctx, &metadataSupplier{metadata: &md})
	ctx = metadata.NewContext(ctx, md)

	return ctx
}

func Extract(ctx context.Context) context.Context {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		fmt.Println("getTraceFromCtx metadata not found")
	}
	ctx = otel.GetTextMapPropagator().Extract(ctx, &metadataSupplier{metadata: &md})
	return ctx
}
