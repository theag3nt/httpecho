# httpecho

`httpecho` is a simple tool which echoes incoming HTTP requests back to its clients.

[![Build Status](https://travis-ci.org/theag3nt/httpecho.svg?branch=master)](https://travis-ci.org/theag3nt/httpecho)
[![Go Report Card](https://goreportcard.com/badge/github.com/theag3nt/httpecho)](https://goreportcard.com/report/github.com/theag3nt/httpecho)

It is useful for testing and debugging HTTP clients and reverse proxies, showing the exact request which arrives to the endpoint. It is inspired by [httpbin](https://github.com/kennethreitz/httpbin) but it is much simpler, and it was created with the intention to be easily runnable on multiple ports at the same time.

## Installation

`httpecho` can be installed with `go get`:

    go get github.com/theag3nt/httpecho

Prebuilt binaries will also be available from GitHub in the future.

## Usage

    httpecho [ip] <port> [port]...

`httpecho` will bind to the IP address and ports specified, echoing incoming HTTP requests (including the request line and headers) as text. If you do not specify an IP address (or the first argument is an invalid IP, but a valid port number) it assumes `0.0.0.0` and listens on all interfaces.

Incoming requests are also logged to the standard output in the following format:

    2006/01/02 15:04:05 GET request from 192.168.0.2 on 192.168.0.1:8080
                        ^^^              ^^^^^^^^^^^    ^^^^^^^^^^^^^^^^
                        request method   remote host    local host and port
## Examples

This will listen for requests on all network interfaces on the standard HTTP port:

    httpecho 80

This will listen for requests on the local machine on ports `8080` and `8081`:

    httpecho 127.0.0.1 8080 8081

### Output

The following is a sample output from `httpecho` running on two ports, receiving a request on each one:

    $ httpecho 8080 8081
    2006/01/02 15:04:05 Listening on 0.0.0.0:8080
    2006/01/02 15:04:05 Listening on 0.0.0.0:8081
    2006/01/02 15:04:10 GET request from 127.0.0.1 on localhost:8080
    2006/01/02 15:04:20 POST request from 127.0.0.1 on localhost:8081

While `httpecho` was running, we have sent two requests using `curl`:

    $ curl localhost:8080
    GET / HTTP/1.1
    Host: localhost:8080
    Accept: */*
    User-Agent: curl/7.47.0

    $ curl -X POST -d hello=world -d http=echo localhost:8081
    POST / HTTP/1.1
    Host: localhost:8081
    Accept: */*
    Content-Length: 21
    Content-Type: application/x-www-form-urlencoded
    User-Agent: curl/7.47.0

    hello=world&http=echo
