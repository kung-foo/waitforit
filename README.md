# waitforit

**waitforit** is a command line utility and library for delaying a script until a requisite service is available.

For example, when using docker-compose, it is common that entrypoint scripts need to wait until the database container has started and been initialized.

Initializing a nodejs server container that depends on a Postrgres instance with a database called *ghost*:

```sh
#!/usr/bin/env sh
set -ex
waitforit -k -r 60 postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@db:5432/ghost
exec npm start --production
```

## Usage

```
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
```
