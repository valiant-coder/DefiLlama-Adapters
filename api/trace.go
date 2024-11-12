package api

import (
	"strings"

	"exapp-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/otel/propagation"

	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func Trace(tracerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		propagator := otel.GetTextMapPropagator()
		request := c.Request
		savedCtx := c.Request.Context()
		defer func() {
			c.Request = c.Request.WithContext(savedCtx)
		}()
		ctx := propagator.Extract(savedCtx, propagation.HeaderCarrier(c.Request.Header))

		opts := []trace.SpanStartOption{
			trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", request)...),
			trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(request)...),
			trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest("api", c.FullPath(), request)...),
			trace.WithSpanKind(trace.SpanKindServer),
		}

		tracer := otel.Tracer(tracerName)
		var span trace.Span
		ctx, span = tracer.Start(ctx, c.FullPath(), opts...)
		defer span.End()

		c.Request = request.WithContext(ctx)
		c.Header("X-Trace-Id", span.SpanContext().TraceID().String())

		req, _ := utils.MarshalToString(reqData(c))
		span.SetAttributes(attribute.String("http.request.body", req))

		jwtToken := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		span.SetAttributes(attribute.String("jwt.token", jwtToken))

		c.Next()

		status := c.Writer.Status()
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(status)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(status)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)
		respData, _ := c.Get("resp")
		resp, _ := utils.MarshalToString(respData)
		span.SetAttributes(attribute.String("http.response.body", resp))

	}
}
