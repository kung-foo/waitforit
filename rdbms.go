package waitforit

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var _ = log.Print

type dbWaiter struct {
	target *Target
	db     *sql.DB
	driver string
}

func newPostgresWaiter(target *Target) *dbWaiter {
	return &dbWaiter{
		target: target,
		driver: "postgres",
	}
}

func newMySQLWaiter(target *Target) *dbWaiter {
	return &dbWaiter{
		target: target,
		driver: "mysql",
	}
}

func (w *dbWaiter) connect() (err error) {
	u := w.target.url
	values := u.Query()

	switch w.driver {
	case "postgres":
		values.Set("connect_timeout", strconv.Itoa(int(w.target.Timeout.Seconds())))
		if w.target.Insecure {
			values.Set("sslmode", "disable")
		}
	case "mysql":
		u.Scheme = ""
		u.Host = fmt.Sprintf("tcp(%s:%d)", w.target.host, w.target.port)
		values.Set("timeout", fmt.Sprintf("%ds", int(w.target.Timeout.Seconds())))
	}

	u.RawQuery = values.Encode()

	dsn := u.String()

	if strings.HasPrefix(dsn, "//") {
		dsn = dsn[2:]
	}

	w.db, err = sql.Open(w.driver, dsn)
	if err != nil {
		return
	}
	return w.db.Ping()
}

func (w *dbWaiter) runTest() (err error) {
	if w.target.Exists != "" {
		var ok string
		q := `select exists (select 1 from %s limit 1)`
		err = w.db.QueryRow(fmt.Sprintf(q, w.target.Exists)).Scan(&ok)
		if err != nil {
			return
		}
	}
	return
}

func (w *dbWaiter) cancel() (err error) {
	if w.db != nil {
		return w.db.Close()
	}
	return
}
