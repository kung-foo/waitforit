package waitforit

import (
	"math"
	"net/url"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTarget(t *testing.T) {
	Convey("Creating a target with", t, func() {
		Convey("no URI should fail", func() {
			wt := &Target{}
			So(wt.init(), ShouldEqual, ErrMalformedURI)
		})

		Convey("an invalid scheme should fail", func() {
			wt := &Target{URI: "htp://w.com:-1"}
			So(wt.init().Error(), ShouldEqual, "Invalid URI scheme")
		})

		Convey("-1 retries should goto maxint", func() {
			wt := &Target{
				URI:     "http://foo",
				Retries: -1,
			}

			So(wt.init(), ShouldBeNil)
			So(wt.Retries, ShouldEqual, math.MaxInt32)
		})
	})

	Convey("Parsing URIs", t, func() {
		Convey("with an invalid port should fail", func() {
			u, _ := url.Parse("http://foo.com:hello/")
			host, port, err := parseHostAndPort(u)

			So(err, ShouldHaveSameTypeAs, &strconv.NumError{})
			So(host, ShouldEqual, "foo.com")
			So(port, ShouldEqual, 0)
		})

		Convey("with a valid URI should succeed", func() {
			var u *url.URL
			u, _ = url.Parse("http://foo.com/")
			host, port, err := parseHostAndPort(u)

			So(err, ShouldBeNil)
			So(host, ShouldEqual, "foo.com")
			So(port, ShouldEqual, 0)

			u, _ = url.Parse("http://foo.com:8080/")
			host, port, err = parseHostAndPort(u)

			So(err, ShouldBeNil)
			So(host, ShouldEqual, "foo.com")
			So(port, ShouldEqual, 8080)

			u, _ = url.Parse("http://foo.com:/")
			host, port, err = parseHostAndPort(u)

			So(err, ShouldBeNil)
			So(host, ShouldEqual, "foo.com")
			So(port, ShouldEqual, 0)
		})

		// TODO: remove this test when IPv6 is supported
		Convey("with an IPv6 literal should fail", func() {
			u, _ := url.Parse("http://[2a02:1788:4fd:cd::c742:cde2]/")
			host, port, err := parseHostAndPort(u)

			So(err, ShouldEqual, ErrMalformedURI)
			So(host, ShouldEqual, "")
			So(port, ShouldEqual, 0)
		})
	})
}

func TestHTTPTarget(t *testing.T) {
	Convey("An HTTP target should", t, func() {
	})
}
