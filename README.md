go-gelf - GELF Library and Writer for Go
========================================

[GELF] (Graylog Extended Log Format) is an application-level logging
protocol that avoids many of the shortcomings of [syslog]. While it
can be run over any stream or datagram transport protocol, it has
special support ([chunking]) to allow long messages to be split over
multiple datagrams.

Versions
--------

This was forked from [Graylog2 go-gelf](https://github.com/Graylog2/go-gelf) to add the GELF HTTP transport protocol.

Currently sending GELF is supported via UDP, TCP and HTTP.
TLS is experimental in the tls branch.

The library provides an API that applications can use to log messages
directly to a Graylog server and an `io.Writer` that can be used to
redirect the standard library's log messages (`os.Stdout`) to a
Graylog server.

[GELF]: http://docs.graylog.org/en/2.2/pages/gelf.html
[syslog]: https://tools.ietf.org/html/rfc5424
[chunking]: http://docs.graylog.org/en/2.2/pages/gelf.html#chunked-gelf


Installing
----------

To install, run:

    go get xiam.li/gelf

Usage
-----

The easiest way to integrate graylog logging into your go app is by
having your `main` function (or even `init`) call `log.SetOutput()`.
By using an `io.MultiWriter`, we can log to both stdout and graylog -
giving us both centralized and local logs.  (Redundancy is nice).

```golang
package main

import (
	"flag"
	"io"
	"log"
	"os"
	"xiam.li/gelf"
)

func main() {
	var graylogAddr string

	flag.StringVar(&graylogAddr, "graylog", "", "graylog server addr")
	flag.Parse()

	if graylogAddr != "" {
		// If using UDP
		gelfWriter, err := gelf.NewUDPWriter(graylogAddr)
		// If using TCP
		//gelfWriter, err := gelf.NewTCPWriter(graylogAddr)
		if err != nil {
			log.Fatalf("gelf.NewWriter: %s", err)
		}
		// log to both stderr and graylog2
		log.SetOutput(io.MultiWriter(os.Stderr, gelfWriter))
		log.Printf("logging to stderr & graylog2@'%s'", graylogAddr)
	}

	// From here on out, any calls to log.Print* functions
	// will appear on stdout, and be sent over UDP or TCP to the
	// specified Graylog2 server.

	log.Printf("Hello gray World")

	// ...
}
```
The above program can be invoked as:

    go run test.go -graylog=localhost:12201

When using UDP messages may be dropped or re-ordered. However, Graylog
server availability will not impact application performance; there is
a small, fixed overhead per log call regardless of whether the target
server is reachable or not.


To Do
-----

- WriteMessage example

License
-------

go-gelf is offered under the MIT license, see LICENSE for details.
