package logx_test

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chi07/go-svc-kit/logx"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestNew_ProdAndStagingUseStdout(t *testing.T) {
	out := captureStdout(func() {
		logger := logx.New("prod")
		logger.Info().Msg("prod message")
	})
	if !strings.Contains(out, "prod message") {
		t.Fatalf("expected prod message in output, got %q", out)
	}

	out = captureStdout(func() {
		logger := logx.New("staging")
		logger.Info().Msg("staging message")
	})
	if !strings.Contains(out, "staging message") {
		t.Fatalf("expected staging message in output, got %q", out)
	}
}

func TestNew_DevUsesConsoleWriter(t *testing.T) {
	out := captureStdout(func() {
		logger := logx.New("dev")
		logger.Info().Msg("hello dev")
	})

	if !strings.Contains(out, "hello dev") {
		t.Fatalf("expected message printed, got %q", out)
	}

	if !(strings.Contains(out, "INF") || strings.Contains(strings.ToLower(out), "info") || strings.Contains(out, "level=")) {
		t.Fatalf("expected console-style hint (INF/info/level=), got %q", out)
	}
}

func TestNew_OtherEnvFallbackToConsole(t *testing.T) {
	for _, env := range []string{"local", "test", "anything"} {
		out := captureStdout(func() {
			logger := logx.New(env)
			logger.Info().Msg("msg " + env)
		})
		if !strings.Contains(out, "msg "+env) {
			t.Fatalf("[%s] expected output to contain message, got %q", env, out)
		}
	}
}

func TestNew_DefaultsDifferBetweenDevAndProd(t *testing.T) {
	prodOut := captureStdout(func() {
		logger := logx.New("prod")
		logger.Info().Msg("x")
	})
	devOut := captureStdout(func() {
		logger := logx.New("dev")
		logger.Info().Msg("x")
	})

	if prodOut == devOut {
		t.Fatalf("expected different outputs for dev vs prod; got identical")
	}

	if !strings.HasPrefix(strings.TrimSpace(prodOut), "{") {
		t.Fatalf("expected prod output to be JSON, got: %q", prodOut)
	}

	if strings.HasPrefix(strings.TrimSpace(devOut), "{") {
		t.Fatalf("expected dev output to be console (non-JSON), got: %q", devOut)
	}
}

func TestNew_TimestampPresent(t *testing.T) {
	out := captureStdout(func() {
		logger := logx.New("prod")
		time.Sleep(1 * time.Millisecond)
		logger.Info().Msg("ts check")
	})
	if !(strings.Contains(out, "\"time\"") || strings.Contains(out, "202") || strings.Contains(out, "level")) {
		t.Fatalf("expected timestamp-ish output, got %q", out)
	}
}
