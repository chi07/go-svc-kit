package dbx_test

import (
	"crypto/tls"
	"reflect"
	"testing"
	"time"

	"github.com/go-pg/pg/v10"

	"github.com/chi07/go-svc-kit/dbx"
)

func TestOptionsFromDSN_SupportedSchemesAndBasics(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		pool        int
		sslInsecure bool
		want        *pg.Options
		wantTLS     *tls.Config // nil means expect nil
		wantErr     bool
	}{
		{
			name:        "postgresql scheme, host without port defaults to 5432, sslmode=require uses TLS (insecure=true respected)",
			dsn:         "postgresql://alice:secret@db.example.com/mydb?sslmode=require",
			pool:        7,
			sslInsecure: true,
			want: &pg.Options{
				Addr:         "db.example.com:5432",
				User:         "alice",
				Password:     "secret",
				Database:     "mydb",
				PoolSize:     7,
				IdleTimeout:  5 * time.Minute,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
				DialTimeout:  3 * time.Second,
				MaxConnAge:   30 * time.Minute,
				MinIdleConns: 2,
			},
			wantTLS: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true,
			},
		},
		{
			name:        "postgres scheme, explicit port, sslmode=disable -> TLS nil",
			dsn:         "postgres://u:p@127.0.0.1:5555/db1?sslmode=disable",
			pool:        3,
			sslInsecure: false,
			want: &pg.Options{
				Addr:         "127.0.0.1:5555",
				User:         "u",
				Password:     "p",
				Database:     "db1",
				PoolSize:     3,
				IdleTimeout:  5 * time.Minute,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
				DialTimeout:  3 * time.Second,
				MaxConnAge:   30 * time.Minute,
				MinIdleConns: 2,
			},
			wantTLS: nil,
		},
		{
			name:        "sslmode verify-full -> TLS (insecure=false respected)",
			dsn:         "postgres://u@host/db2?sslmode=verify-full",
			pool:        5,
			sslInsecure: false,
			want: &pg.Options{
				Addr:         "host:5432",
				User:         "u",
				Password:     "",
				Database:     "db2",
				PoolSize:     5,
				IdleTimeout:  5 * time.Minute,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
				DialTimeout:  3 * time.Second,
				MaxConnAge:   30 * time.Minute,
				MinIdleConns: 2,
			},
			wantTLS: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: false,
			},
		},
		{
			name:        "empty sslmode -> default no TLS",
			dsn:         "postgres://u:pw@h/db3",
			pool:        1,
			sslInsecure: true,
			want: &pg.Options{
				Addr:         "h:5432",
				User:         "u",
				Password:     "pw",
				Database:     "db3",
				PoolSize:     1,
				IdleTimeout:  5 * time.Minute,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
				DialTimeout:  3 * time.Second,
				MaxConnAge:   30 * time.Minute,
				MinIdleConns: 2,
			},
			wantTLS: nil,
		},
		{
			name:    "unsupported scheme -> error",
			dsn:     "mysql://user:pw@h/db",
			pool:    1,
			wantErr: true,
		},
		{
			name:    "missing database in path -> error",
			dsn:     "postgres://user:pw@h",
			pool:    1,
			wantErr: true,
		},
		{
			name:    "bad URL -> error",
			dsn:     "::::not-a-url",
			pool:    1,
			wantErr: true,
		},
	}

	// Compare pg.Options except TLSConfig (checked separately)
	eqOptionsSansTLS := func(got, want *pg.Options) bool {
		if got == nil || want == nil {
			return got == want
		}
		cg := *got
		cw := *want
		cg.TLSConfig = nil
		cw.TLSConfig = nil
		return reflect.DeepEqual(&cg, &cw)
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := dbx.OptionsFromDSN(tc.dsn, tc.pool, tc.sslInsecure)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (opts=%+v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !eqOptionsSansTLS(got, tc.want) {
				t.Fatalf("options mismatch.\n got: %+v\nwant: %+v", got, tc.want)
			}
			// TLS expectations
			if tc.wantTLS == nil {
				if got.TLSConfig != nil {
					t.Fatalf("expected TLSConfig=nil, got: %+v", got.TLSConfig)
				}
			} else {
				if got.TLSConfig == nil {
					t.Fatalf("expected TLSConfig non-nil")
				}
				if got.TLSConfig.MinVersion != tc.wantTLS.MinVersion {
					t.Fatalf("MinVersion mismatch: got %v want %v", got.TLSConfig.MinVersion, tc.wantTLS.MinVersion)
				}
				if got.TLSConfig.InsecureSkipVerify != tc.wantTLS.InsecureSkipVerify {
					t.Fatalf("InsecureSkipVerify mismatch: got %v want %v", got.TLSConfig.InsecureSkipVerify, tc.wantTLS.InsecureSkipVerify)
				}
			}
		})
	}
}

func TestOptionsFromDSN_DefaultsAndPoolSize(t *testing.T) {
	got, err := dbx.OptionsFromDSN("postgresql://u@h/db?sslmode=disable", 13, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PoolSize != 13 {
		t.Fatalf("PoolSize mismatch: got %d want %d", got.PoolSize, 13)
	}
	// Sanity check of the preconfigured timeouts
	if got.IdleTimeout != 5*time.Minute || got.ReadTimeout != 5*time.Second || got.WriteTimeout != 5*time.Second ||
		got.DialTimeout != 3*time.Second || got.MaxConnAge != 30*time.Minute || got.MinIdleConns != 2 {
		t.Fatalf("unexpected default timeouts/limits: %+v", got)
	}
}

func TestApplySetMap_NoPanicAndReturnsQuery(t *testing.T) {
	db := pg.Connect(&pg.Options{}) // creating a DB handle does not connect until a query is executed
	defer db.Close()

	q := db.Model((*struct{})(nil))
	fields := map[string]any{"col1": 1, "col2": "x", "col3": true}

	got := dbx.ApplySetMap(q, fields)
	if got == nil {
		t.Fatal("ApplySetMap returned nil query")
	}

	if got != q && reflect.ValueOf(*got).IsZero() {
		t.Fatal("returned query appears zeroed unexpectedly")
	}
}
