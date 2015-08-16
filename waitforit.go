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
	// ErrMalformedURI is returned when the URI is malformed
	ErrMalformedURI = errors.New("Malformed URI")

	// ErrInvalidScheme is returned when the URI scheme is invalid
	ErrInvalidScheme = errors.New("Invalid URI scheme")

	// ErrMaxRetries is returned when the maximun number of retries has been
	// reached without success
	ErrMaxRetries = errors.New("Maximum retries reached")

	// ErrTimeout is returned when the waitable request timesout
	ErrTimeout = errors.New("Timeout")
)

// A Target represents a waitable endpoint
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
	waiter     waitable
}

type waitable interface {
	connect() error
	runTest() error
	cancel() error
}

func (t *Target) init() (err error) {
	if t.URI == "" {
		return ErrMalformedURI
	}

	t.url, err = url.Parse(t.URI)
	if err != nil {
		return
	}

	t.host, t.port, err = parseHostAndPort(t.url)
	if err != nil {
		return
	}

	t.waiter, err = makeWaitable(t.url.Scheme, t)
	if err != nil {
		return
	}

	if t.Retries == -1 {
		t.Retries = math.MaxInt32
	}

	return
}

// Elapsed returns the amount of time since Wait was called
func (t *Target) Elapsed() time.Duration {
	return t.elapsed
}

// Wait begins the blocking request to the endpoint
func (t *Target) Wait() (err error) {
	t.start = time.Now()

	defer func() {
		t.elapsed = time.Now().Sub(t.start)
	}()

	err = t.init()
	if err != nil {
		return
	}

	errCh := make(chan error)
	successCh := make(chan bool)

	run := func() {
		err = t.waiter.connect()
		if err != nil {
			errCh <- err
			return
		}

		err = t.waiter.runTest()
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
			cerr := t.waiter.cancel()
			if cerr != nil {
				log.Fatal(cerr)
			}
			break
		case <-successCh:
			return nil
		case <-time.After(t.Timeout):
			cerr := t.waiter.cancel()
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

func parseHostAndPort(url *url.URL) (host string, port int, err error) {
	tokens := strings.Split(url.Host, ":")
	switch len(tokens) {
	case 1:
		host = tokens[0]
	case 2:
		host = tokens[0]
		if len(tokens[1]) > 0 {
			port, err = strconv.Atoi(tokens[1])
		}
	default:
		// IPv6?
		err = ErrMalformedURI
	}
	return
}

func makeWaitable(scheme string, t *Target) (w waitable, err error) {
	switch t.url.Scheme {
	case "http", "https":
		w = &httpWaiter{target: t}
	case "redis":
		w = &redisWaiter{target: t}
	case "mysql":
		w = newMySQLWaiter(t)
	case "postgres":
		w = newPostgresWaiter(t)
	default:
		return nil, ErrInvalidScheme
	}

	return
}
