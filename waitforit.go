package waitforit

import (
	"errors"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	ErrMalformedURI  = errors.New("Malformed URI")
	ErrInvalidScheme = errors.New("Invalid URI scheme")
	ErrMaxRetries    = errors.New("Maximum retries reached")
	ErrTimeout       = errors.New("Timeout")
)

type Target struct {
	Timeout    time.Duration
	Retries    int
	RetryDelay time.Duration
	URI        string
	Exists     string
	Insecure   bool
	elapsed    time.Duration
	start      time.Time
	url        *url.URL
	host       string
	port       int
	waiter     Waitable
}

type Waitable interface {
	Connect() error
	RunTest() error
	Cancel() error
}

func (t *Target) Elapsed() time.Duration {
	if t.elapsed == 0 {
		return time.Now().Sub(t.start)
	}
	return t.elapsed
}

func (t *Target) Wait() (err error) {
	defer func() {
		t.elapsed = time.Now().Sub(t.start)
	}()

	t.start = time.Now()

	t.url, err = url.Parse(t.URI)
	if err != nil {
		return
	}

	if t.Retries == -1 {
		t.Retries = math.MaxInt32
	}

	switch t.url.Scheme {
	case "http", "https":
		t.waiter = &HTTPWaiter{target: t}
	case "redis":
		t.waiter = &RedisWaiter{target: t}
	case "mysql":
		t.waiter = NewMySQLWaiter(t)
	case "postgres":
		t.waiter = NewPostgresWaiter(t)
	default:
		return ErrInvalidScheme
	}

	tokens := strings.Split(t.url.Host, ":")
	switch len(tokens) {
	case 1:
		t.host = tokens[0]
	case 2:
		t.host = tokens[0]
		t.port, err = strconv.Atoi(tokens[1])
		if err != nil {
			return
		}
	default:
		// IPv6?
		return ErrMalformedURI
	}

	errCh := make(chan error)
	successCh := make(chan bool)

	run := func() {
		err = t.waiter.Connect()
		if err != nil {
			errCh <- err
			return
		}

		err = t.waiter.RunTest()
		if err != nil {
			errCh <- err
			return
		}
		successCh <- true
	}

	for {
		start := time.Now()
		go run()

		select {
		case err = <-errCh:
			log.Warn(err)
			cerr := t.waiter.Cancel()
			if cerr != nil {
				log.Fatal(cerr)
			}
			break
		case <-successCh:
			return nil
		case <-time.After(t.Timeout):
			cerr := t.waiter.Cancel()
			if cerr != nil {
				log.Fatal(cerr)
			} else {
				err = ErrTimeout
			}
		}

		t.Retries--

		if t.Retries < 0 {
			log.Errorf("Last error: %v", err)
			return ErrMaxRetries
		}

		time.Sleep(t.RetryDelay - time.Now().Sub(start))
	}
}
