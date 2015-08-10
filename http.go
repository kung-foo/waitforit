package waitforit

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/franela/goreq"
)

var _ = log.Print

const maxRedirects = 5

type HTTPWaiter struct {
	target *Target
	res    *goreq.Response
}

func (w *HTTPWaiter) Connect() (err error) {
	goreq.SetConnectTimeout(w.target.Timeout)

	req := goreq.Request{
		Uri:          w.target.url.String(),
		Timeout:      w.target.Timeout,
		MaxRedirects: maxRedirects,
		Insecure:     w.target.Insecure,
		UserAgent:    fmt.Sprintf("waitforit"),
		Accept:       "*/*",
	}

	if w.target.url.User != nil {
		req.BasicAuthUsername = w.target.url.User.Username()
		req.BasicAuthPassword, _ = w.target.url.User.Password()
	}

	w.res, err = req.Do()

	if err != nil {
		if serr, ok := err.(*goreq.Error); ok {
			if serr.Timeout() {
				return ErrTimeout
			}
		}
		return err
	}

	body := w.res.Body
	if body != nil {
		defer body.Close()
	}

	if w.res.StatusCode != 200 {
		return fmt.Errorf(w.res.Status)
	}

	return nil
}

func (w *HTTPWaiter) RunTest() error {
	// nothing implemented yet
	return nil
}

func (w *HTTPWaiter) Cancel() error {
	if w.res != nil {
		w.res.CancelRequest()
	}
	return nil
}
