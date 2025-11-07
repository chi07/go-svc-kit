package usecasex_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/chi07/go-svc-kit/usecasex"
)

func newLoggerBuf() (*zerolog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	log := zerolog.New(&buf).With().Timestamp().Logger()
	return &log, &buf
}

func TestEnrichBatch_TableDriven(t *testing.T) {
	type tcDef struct {
		name       string
		timeout    time.Duration
		tasks      map[string]func(context.Context) error
		wantSubstr []string // substrings expected to appear in logs (all must be present)
	}

	errBoom := errors.New("boom")

	tests := []tcDef{
		{
			name:    "no tasks -> no log",
			timeout: 50 * time.Millisecond,
			tasks:   map[string]func(context.Context) error{},
		},
		{
			name:    "all success -> no warn log",
			timeout: 100 * time.Millisecond,
			tasks: map[string]func(context.Context) error{
				"a": func(ctx context.Context) error { return nil },
				"b": func(ctx context.Context) error { return nil },
			},
		},
		{
			name:    "task returns error -> warn log with task and error",
			timeout: 100 * time.Millisecond,
			tasks: map[string]func(context.Context) error{
				"a": func(ctx context.Context) error { return nil },
				"b": func(ctx context.Context) error { return errBoom },
			},
			wantSubstr: []string{
				`"level":"warn"`,
				`"task":"b"`,
				`"message":"enrich_failed"`,
				`"error":"boom"`,
			},
		},
		{
			name:    "task times out -> warn log with deadline exceeded",
			timeout: 50 * time.Millisecond,
			tasks: map[string]func(context.Context) error{
				"slow": func(ctx context.Context) error {
					// Block until context is done, then return its error
					<-ctx.Done()
					return ctx.Err()
				},
			},
			wantSubstr: []string{
				`"level":"warn"`,
				`"task":"slow"`,
				`"message":"enrich_failed"`,
				`"error":"context deadline exceeded"`,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			log, buf := newLoggerBuf()
			usecasex.EnrichBatch(context.Background(), tt.timeout, *log, tt.tasks)
			got := buf.String()
			for _, sub := range tt.wantSubstr {
				if !strings.Contains(got, sub) {
					t.Fatalf("log missing substring %q\nlog:\n%s", sub, got)
				}
			}
			// When no substrings expected, ensure no warn/error from enrich_failed appears
			if len(tt.wantSubstr) == 0 && strings.Contains(got, "enrich_failed") {
				t.Fatalf("expected no enrich_failed logs, got:\n%s", got)
			}
		})
	}
}

func TestEnrichBatch_RunsConcurrently(t *testing.T) {
	// Two tasks each sleep ~120ms; if they ran sequentially we'd see ~240ms+.
	// With concurrency we expect < 200ms (give some buffer for CI).
	log, _ := newLoggerBuf()
	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(2)

	tasks := map[string]func(context.Context) error{
		"t1": func(ctx context.Context) error {
			defer wg.Done()
			time.Sleep(120 * time.Millisecond)
			return nil
		},
		"t2": func(ctx context.Context) error {
			defer wg.Done()
			time.Sleep(120 * time.Millisecond)
			return nil
		},
	}

	usecasex.EnrichBatch(context.Background(), 2*time.Second, *log, tasks)
	wg.Wait()
	elapsed := time.Since(start)

	if elapsed > 200*time.Millisecond {
		t.Fatalf("expected concurrent execution (<200ms), got %v", elapsed)
	}
}
