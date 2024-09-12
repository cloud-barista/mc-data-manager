package middlewares

import (
	"context"
	"fmt"
	"time"

	"github.com/cloud-barista/mc-data-manager/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// TracingMiddleware intercepts the request, sets up tracing information, and logs both request and response details.
func TracingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the context and initialize trace and span IDs
		ctx := c.Request().Context()
		traceId := c.Response().Header().Get(echo.HeaderXRequestID)
		spanId := fmt.Sprintf("%d", time.Now().UnixNano())

		// Store trace and span IDs in the context
		ctx = context.WithValue(ctx, logger.TraceIdKey, traceId)
		ctx = context.WithValue(ctx, logger.SpanIdKey, spanId)

		// Create a logger with trace_id and span_id and store it in the context
		requestLogger := log.With().
			Str("Host", c.Request().Host).
			Str("RemoteAddr", c.Request().RemoteAddr).
			Str("RequestURI", c.Request().RequestURI).
			Str("UserAgent", c.Request().UserAgent()).
			Str("X-Request-ID", c.Request().Header.Get("X-Request-ID")).
			Str("X-Trace-ID", c.Request().Header.Get("X-Trace-ID")).
			Str("X-Forwarded-For", c.Request().Header.Get("X-Forwarded-For")).
			Str("X-Real-IP", c.Request().Header.Get("X-Real-IP")).
			Str("Authorization", c.Request().Header.Get("Authorization")).
			Str(string(logger.TraceIdKey), traceId).
			Str(string(logger.SpanIdKey), spanId).
			Caller().
			Logger()

		// Add the logger with context
		ctx = requestLogger.WithContext(ctx)
		c.SetRequest(c.Request().WithContext(ctx))

		// Log the incoming request
		log.Ctx(ctx).Info().Msg("[tracing] receive request")

		// Measure the latency
		startTime := time.Now()
		latency := time.Since(startTime)
		// Log the response details
		c.Response().Before(func() {
			log.Ctx(ctx).Info().
				Int("Status", c.Response().Status).
				Int64("Latency", latency.Nanoseconds()).
				Str("LatencyHuman", latency.String()).
				Int64("BytesIn", c.Request().ContentLength).
				Int64("BytesOut", c.Response().Size).
				Msg("[tracing] send response")
		})

		// Return the error if any
		return next(c)
	}
}
