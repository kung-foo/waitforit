package waitforit

import (
	"fmt"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/mediocregopher/radix.v2/redis"
)

var _ = log.Print

var (
	reRedisPath = regexp.MustCompile(`^/(\d{1,2})$`)
)

type RedisWaiter struct {
	target *Target
	db     *redis.Client
}

func (w *RedisWaiter) Connect() (err error) {
	port := w.target.port
	if port == 0 {
		port = 6379
	}

	w.db, err = redis.DialTimeout(
		"tcp",
		fmt.Sprintf("%s:%d", w.target.host, port), w.target.Timeout)

	if err != nil {
		return
	}

	if w.target.url.User != nil {
		auth := w.target.url.User.Username()
		var ok string
		ok, err = w.db.Cmd("AUTH", auth).Str()
		if err != nil {
			return
		}
		if ok != "OK" {
			return fmt.Errorf(ok)
		}
	}

	pong, err := w.db.Cmd("PING").Str()
	if err != nil {
		return
	}

	if pong != "PONG" {
		return fmt.Errorf("Invalid PING response: %s", pong)
	}

	return
}

func (w *RedisWaiter) RunTest() error {
	dbNum := "0"

	if w.target.Exists != "" {
		m := reRedisPath.FindStringSubmatch(w.target.url.Path)
		if m != nil && len(m) == 2 {
			dbNum = m[1]
			ok, err := w.db.Cmd("SELECT", dbNum).Str()
			if err != nil {
				return err
			}
			if ok != "OK" {
				return fmt.Errorf(ok)
			}
		}
		count, err := w.db.Cmd("EXISTS", w.target.Exists).Int()
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("redis key %s does not exist in db %s", w.target.Exists, dbNum)
		}
	}
	return nil
}

func (w *RedisWaiter) Cancel() error {
	if w.db != nil {
		return w.db.Close()
	}
	return nil
}
