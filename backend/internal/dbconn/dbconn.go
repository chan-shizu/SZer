package dbconn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func Open(ctx context.Context) (*sql.DB, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	databaseURL = withDefaultSSLMode(databaseURL)

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func withDefaultSSLMode(databaseURL string) string {
	if os.Getenv("ENV") != "development" {
		return databaseURL
	}

	u, err := url.Parse(databaseURL)
	if err != nil {
		// If it's not a URL (e.g. keyword/value conn string), don't mutate.
		return databaseURL
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return databaseURL
	}

	q := u.Query()
	if q.Get("sslmode") != "" {
		return databaseURL
	}

	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return fmt.Sprintf("%s", u.String())
}
