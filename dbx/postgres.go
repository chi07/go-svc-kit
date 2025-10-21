package dbx

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
)

// OptionsFromDSN: parse DSN Postgres (postgresql://user:pass@host:port/db?sslmode=require)
// và tạo *pg.Options dùng chung cho mọi service.

func OptionsFromDSN(dsn string, poolSize int, sslInsecure bool) (*pg.Options, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	if u.Scheme != "postgresql" && u.Scheme != "postgres" {
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	host, port, err := splitHostPort(u.Host)
	if err != nil {
		return nil, err
	}
	db := strings.TrimPrefix(u.Path, "/")
	if db == "" {
		return nil, fmt.Errorf("database name missing in DSN path")
	}

	user := ""
	pass := ""
	if u.User != nil {
		user = u.User.Username()
		pw, _ := u.User.Password()
		pass = pw
	}

	var tlsConf *tls.Config
	switch strings.ToLower(u.Query().Get("sslmode")) {
	case "", "disable":
		tlsConf = nil
	case "require", "verify-ca", "verify-full":
		tlsConf = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: sslInsecure, // chỉ nên true ở dev/staging
		}
	}

	return &pg.Options{
		Addr:         net.JoinHostPort(host, port),
		User:         user,
		Password:     pass,
		Database:     db,
		PoolSize:     poolSize,
		IdleTimeout:  5 * time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		DialTimeout:  3 * time.Second,
		MaxConnAge:   30 * time.Minute,
		MinIdleConns: 2,
		TLSConfig:    tlsConf,
	}, nil
}

func splitHostPort(h string) (string, string, error) {
	host, port, err := net.SplitHostPort(h)
	if err == nil {
		return host, port, nil
	}
	if strings.Contains(err.Error(), "missing port in address") {
		return h, "5432", nil
	}
	return "", "", err
}

func ApplySetMap(q *pg.Query, fields map[string]any) *pg.Query {
	for k, v := range fields {
		q = q.Set(fmt.Sprintf("%s = ?", k), v)
	}
	return q
}
