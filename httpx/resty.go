package httpx

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type ctxKey string

const (
	ctxKeyRequestID ctxKey = "request_id"
	ctxKeyStart     ctxKey = "resty_start_time"
)

func ContextWithRequestID(ctx context.Context, reqID string) context.Context {
	if reqID == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyRequestID, reqID)
}
func RequestIDFromContext(ctx context.Context) string {
	if v := ctx.Value(ctxKeyRequestID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Resty ↔ Zerolog adapter
type zerologAdapter struct{ log zerolog.Logger }

func (a *zerologAdapter) Errorf(f string, v ...interface{}) { a.log.Error().Msgf(f, v...) }
func (a *zerologAdapter) Warnf(f string, v ...interface{})  { a.log.Warn().Msgf(f, v...) }
func (a *zerologAdapter) Debugf(f string, v ...interface{}) { a.log.Debug().Msgf(f, v...) }

func NewBaseResty(log zerolog.Logger, debug bool, timeout time.Duration) *resty.Client {
	if timeout <= 0 {
		timeout = 3200 * time.Millisecond
	}

	c := resty.New().
		SetTimeout(timeout).
		SetRetryCount(1).
		SetRetryWaitTime(200 * time.Millisecond).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
				return false
			}
			return r != nil && r.StatusCode() >= 500
		})

	if debug {
		c.SetLogger(&zerologAdapter{log})
		c.EnableTrace()
	}

	c.OnBeforeRequest(func(_ *resty.Client, r *resty.Request) error {
		r.SetContext(context.WithValue(r.Context(), ctxKeyStart, time.Now()))
		if reqID := RequestIDFromContext(r.Context()); reqID != "" {
			r.SetHeader("X-Request-ID", reqID)
		}
		if debug {
			log.Debug().
				Str("event", "resty_before").
				Str("method", r.Method).
				Str("url", r.URL).
				Interface("headers", redactHeaders(r.Header)).
				Interface("query", r.QueryParam).
				Str("request_id", RequestIDFromContext(r.Context())).
				Msg("resty_before")
		}
		return nil
	})

	c.OnAfterResponse(func(_ *resty.Client, resp *resty.Response) error {
		ctx := resp.Request.Context()
		start, _ := ctx.Value(ctxKeyStart).(time.Time)
		elapsed := time.Since(start)
		evt := log.Info()
		switch sc := resp.StatusCode(); {
		case sc >= 500:
			evt = log.Error()
		case sc >= 400:
			evt = log.Warn()
		}
		evt = evt.
			Str("event", "resty_after").
			Str("method", resp.Request.Method).
			Str("url", resp.Request.URL).
			Int("status", resp.StatusCode()).
			Dur("duration", elapsed).
			Int64("resp_bytes", resp.Size()).
			Int("attempt", resp.Request.Attempt)
		if reqID := RequestIDFromContext(ctx); reqID != "" {
			evt = evt.Str("request_id", reqID)
		}
		if debug {
			ti := resp.Request.TraceInfo()
			evt = evt.Dict("trace", zerolog.Dict().
				Dur("total", ti.TotalTime).
				Dur("dns", ti.DNSLookup).
				Dur("tcp", ti.TCPConnTime).
				Dur("tls", ti.TLSHandshake).
				Dur("server", ti.ServerTime).
				Dur("conn_idle", ti.ConnIdleTime).
				Bool("conn_reused", ti.IsConnReused).
				Bool("conn_was_idle", ti.IsConnWasIdle),
			)
		}
		evt.Msg("resty_after")
		return nil
	})

	return c
}

var sensitive = []string{"authorization", "x-api-key", "api-key", "x-auth-token", "proxy-authorization", "set-cookie", "cookie"}

func redactHeaders(h http.Header) map[string]string {
	if len(h) == 0 {
		return nil
	}
	out := make(map[string]string, len(h))
	for k, vals := range h {
		lk := strings.ToLower(k)
		red := false
		for _, s := range sensitive {
			if lk == s {
				out[k] = "***REDACTED***"
				red = true
				break
			}
		}
		if !red {
			out[k] = strings.Join(vals, ",")
		}
	}
	return out
}
