package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docopt/docopt-go"
	"github.com/kung-foo/waitforit"
)

// VERSION is set by the makefile
var VERSION = "0.0.0"

func main() {
	mainEx(os.Args[1:])
}

var usage = `
Usage:
    waitforit [options] <uri> [--exists=<item>]
    waitforit -h | --help | --version

Options:
    -h --help                Show this screen.
    -s, --silent             Silent
    -t, --timeout=<timeout>  Maximum time per attempt [default: 5s]
                             Valid time units are "ns", "us" (or "µs"), "ms",
                             "s", "m", "h"
    -r, --retry=<retry>      Number of times to retry after failure [default: 5]
                             Set to -1 to always retry
    --retry-delay=<delay>    Time to wait before retrying [default: 1s]
                             Valid time units are "ns", "us" (or "µs"), "ms",
                             "s", "m", "h"
    --exists=<item>          Wait for item to exist
                             For redis, this is waits for the key to exist
                             For DBs, this waits for the table to exist
    -k, --insecure           Disable SSL certificate validation
    --expand-env             Expand environmental variables in URI
    --version                Show version.

URI Examples:
    redis://127.0.0.1
    redis://MAH_SECRET@127.0.0.1/7
    http://username:password@somesite.com:8080/hello
    mysql://scott:tiger@127.0.0.1/ghost
    postgres://ghost:tiger@127.0.0.1/ghost
`

const debug = false

func mainEx(argv []string) {
	var err error
	args, err := docopt.Parse(usage, argv, true, VERSION, false)

	if debug {
		log.SetLevel(log.DebugLevel)

		sigChan := make(chan os.Signal)
		go func() {
			stacktrace := make([]byte, 8192)
			for _ = range sigChan {
				length := runtime.Stack(stacktrace, true)
				fmt.Println(string(stacktrace[:length]))
			}
		}()
		signal.Notify(sigChan, syscall.SIGQUIT)

		go func() {
			for {
				<-time.After(time.Second * 5)
				log.Debugf("NumGoroutine: %d", runtime.NumGoroutine())
			}
		}()
	}

	target := &waitforit.Target{}

	if args["--silent"].(bool) {
		log.SetOutput(ioutil.Discard)
	}

	target.Timeout, err = time.ParseDuration(args["--timeout"].(string))
	onErrorExit(err)

	target.RetryDelay, err = time.ParseDuration(args["--retry-delay"].(string))
	onErrorExit(err)

	target.Insecure = args["--insecure"].(bool)

	target.URI = args["<uri>"].(string)

	if args["--expand-env"].(bool) {
		target.URI = os.ExpandEnv(target.URI)
	}

	target.Retries, err = strconv.Atoi(args["--retry"].(string))
	onErrorExit(err)

	if args["--exists"] != nil {
		target.Exists = args["--exists"].(string)
	}

	if target.Retries > 0 {
		go func() {
			<-time.After(target.Timeout * time.Duration(target.Retries+1))
			log.Fatal("Failsafe timeout triggered. Something is not right.")
			os.Exit(1)
		}()
	}

	err = target.Wait()

	log.Infof("Elapsed time: %s", target.Elapsed())

	onErrorExit(err)
}

func onErrorExit(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
