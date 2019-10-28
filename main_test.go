package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type bufferLogger struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *bufferLogger) Print(args ...interface{}) {
	l.mu.Lock()
	fmt.Fprint(&l.b, args...)
	l.mu.Unlock()
}

func (l *bufferLogger) Printf(format string, args ...interface{}) {
	l.mu.Lock()
	fmt.Fprintf(&l.b, format, args...)
	l.mu.Unlock()
}

func (l *bufferLogger) Println(args ...interface{}) {
	l.mu.Lock()
	fmt.Fprintln(&l.b, args...)
	l.mu.Unlock()
}

func (l *bufferLogger) String() string {
	return l.b.String()
}

type addr struct {
	value string
}

func newAddr(value string) net.Addr {
	return addr{value}
}

func (a addr) String() string {
	return a.value
}

func (a addr) Network() string {
	return "tcp"
}

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      string
		wantIP    string
		wantPorts string
		wantErr   bool
	}{
		{"Local/Single/v4", "127.0.0.1,8080", "127.0.0.1", "8080", false},
		{"Local/Multiple/v4", "127.0.0.1,8080,8081", "127.0.0.1", "8080,8081", false},
		{"Private/Single/v4", "192.168.0.1,8080", "192.168.0.1", "8080", false},
		{"Private/Multiple/v4", "192.168.0.1,8080,8081", "192.168.0.1", "8080,8081", false},
		{"Local/Single/v6", "::1,8080", "[::1]", "8080", false},
		{"Local/Multiple/v6", "::1,8080,8081", "[::1]", "8080,8081", false},
		{"Private/Single/v6", "fd30:3fac:747b::1,8080", "[fd30:3fac:747b::1]", "8080", false},
		{"Private/Multiple/v6", "fd30:3fac:747b::1,8080,8081", "[fd30:3fac:747b::1]", "8080,8081", false},
		{"Any/Single/v4", "0.0.0.0,8080", "0.0.0.0", "8080", false},
		{"Any/Multiple/v4", "0.0.0.0,8080,8081", "0.0.0.0", "8080,8081", false},
		{"Any/Single/v6", "::,8080", "[::]", "8080", false},
		{"Any/Multiple/v6", "::,8080,8081", "[::]", "8080,8081", false},
		{"None/Single", "8080", "0.0.0.0", "8080", false},
		{"None/Multiple", "8080,8081", "0.0.0.0", "8080,8081", false},
		{"Negative/TooFewArgs", "", "", "", true},
		{"Negative/InvalidIPRange", "123.456.789", "", "", true},
		{"Negative/InvalidIPDomain", "example.com", "", "", true},
		{"Negative/InvalidPortRange", "127.0.0.1,99999", "", "", true},
		{"Negative/InvalidPortName", "127.0.0.1,http", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare arguments and expected results
			args := strings.Split(tt.args, ",")
			if tt.args == "" {
				args = nil
			}
			wantPorts := strings.Split(tt.wantPorts, ",")
			if tt.wantPorts == "" {
				wantPorts = nil
			}

			ip, ports, err := validateArgs(args)

			// Check the returned error
			if (err != nil) != tt.wantErr {
				expectErr := "expected"
				if !tt.wantErr {
					expectErr = "unexpected"
				}
				t.Errorf("error is %s, got error: %v", expectErr, err)
				return
			}
			// Check validated IP address
			if ip != tt.wantIP {
				t.Errorf("invalid IP: got %q, want %q", ip, tt.wantIP)
			}
			// Check validated port list
			if !reflect.DeepEqual(ports, wantPorts) {
				t.Errorf("invalid ports: got %v, want %v", ports, wantPorts)
			}
		})
	}
}

func TestDumpHandlerGetWithHeaders(t *testing.T) {
	// Prepare expected results
	wantStatus := http.StatusOK
	wantBody := "GET / HTTP/1.1\r\nHost: 127.0.0.1:1234\r\nAccept: */*\r\n\r\n"

	// Prepare request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "127.0.0.1:1234"
	req.Header.Set("Accept", "*/*")

	// Send request to handler
	rr := httptest.NewRecorder()
	handler := dumpHandler(nil)
	handler.ServeHTTP(rr, req)

	// Check response status
	if status := rr.Code; status != wantStatus {
		t.Errorf("invalid status code: got %v, want %v", status, wantStatus)
	}
	// Check response body
	if body := rr.Body.String(); body != wantBody {
		t.Errorf("invalid body:\ngot  %q\nwant %q", body, wantBody)
	}
}

func TestDumpHandlerPostWithBody(t *testing.T) {
	// Prepare expected results
	wantStatus := http.StatusOK
	wantBody := "POST /form HTTP/1.1\r\n\r\nhello=world&http=echo"

	// Prepare request
	data := url.Values{}
	data.Add("hello", "world")
	data.Add("http", "echo")
	req, err := http.NewRequest("POST", "/form", bytes.NewBufferString(data.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	// Send request to handler
	rr := httptest.NewRecorder()
	handler := dumpHandler(nil)
	handler.ServeHTTP(rr, req)

	// Check response status
	if status := rr.Code; status != wantStatus {
		t.Errorf("invalid status code: got %v, want %v", status, wantStatus)
	}
	// Check response body
	if body := rr.Body.String(); body != wantBody {
		t.Errorf("invalid body:\ngot  %q\nwant %q", body, wantBody)
	}
}

func TestLogHandler(t *testing.T) {
	tests := []struct {
		name   string
		method string
		remote string
		local  string
		host   string
	}{
		{"LoopbackHost", "GET", "127.0.0.1", "127.0.0.1:1234", ""},
		{"PrivateHost", "POST", "192.168.0.2", "192.168.0.1:1234", ""},
		{"DomainHost", "HEAD", "127.0.0.1", "127.0.0.1:1234", "localhost:1234"},
	}

	// Prepare expected results
	wantStatus := http.StatusTeapot
	wantBody := "tea time!"
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, wantBody, wantStatus)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantOutput := fmt.Sprintf("%s request from %s on %s\n", tt.method, tt.remote, tt.local)

			// Prepare context to mimic behaviour of the real server
			ctx := context.WithValue(context.Background(), http.LocalAddrContextKey, newAddr(tt.local))
			// Prepare request
			req, err := http.NewRequestWithContext(ctx, tt.method, "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.host == "" {
				req.Host = tt.local
			} else {
				req.Host = tt.host
			}
			req.RemoteAddr = tt.remote

			// Send request to handler
			rr := httptest.NewRecorder()
			log := &bufferLogger{}
			handler := logHandler(log, wrappedHandler)
			handler.ServeHTTP(rr, req)

			// Check handler wrapping
			if status := rr.Code; status != wantStatus {
				t.Errorf("invalid status code: got %v, want %v", status, wantStatus)
			}
			if body := strings.TrimRight(rr.Body.String(), "\n"); body != wantBody {
				t.Errorf("invalid body:\ngot  %q\nwant %q", body, wantBody)
			}
			// Check logged message
			if output := log.String(); output != wantOutput {
				t.Errorf("invalid log message:\ngot  %q\nwant %q", output, wantOutput)
			}
		})
	}
}
