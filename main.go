package main

import (
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
)

func printUsage() {
	fmt.Println("usage: httpecho [ip] <port> [port]...")
	os.Exit(1)
}

func validateArgs(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("not enough arguments")
	}

	ip := args[0]
	ports := args[1:]

	// Validate IP address
	if parsedIP := net.ParseIP(strings.TrimSpace(ip)); parsedIP != nil {
		ip = parsedIP.String()
		// Enclose IPv6 addresses in brackets
		if parsedIP.To4() == nil {
			ip = fmt.Sprintf("[%s]", ip)
		}
	} else {
		// IP is invalid, maybe it's a port number
		if _, err := strconv.ParseUint(ip, 10, 16); err != nil {
			return "", nil, fmt.Errorf("ip address %q is invalid", ip)
		}

		// Listen on all interfaces if IP is unspecified
		ip = "0.0.0.0"
		ports = args[0:]
	}

	// Validate port numbers
	for _, port := range ports {
		if _, err := strconv.ParseUint(strings.TrimSpace(port), 10, 16); err != nil {
			return "", nil, fmt.Errorf("port number %q is invalid", port)
		}
	}

	return ip, ports, nil
}

type logger interface {
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}

func dumpHandler(log logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(dump); err != nil {
			log.Printf("Unable to write response: %v", err)
		}
	})
}

func logHandler(log logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		remote, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			remote = r.RemoteAddr
		}
		localaddr, _ := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
		local := localaddr.String()

		log.Printf("%s request from %s on %s\n", method, remote, local)

		h.ServeHTTP(w, r)
	})
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
	}

	log := stdlog.New(os.Stderr, "", stdlog.LstdFlags)

	ip, ports, err := validateArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("Error while validating arguments: %v", err)
	}

	// Prepare request handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", dumpHandler(log))
	logMux := logHandler(log, mux)

	// Listen on specified ports
	for _, port := range ports[1:] {
		port := port
		go func() {
			log.Printf("Listening on %s:%s", ip, port)
			if err := http.ListenAndServe(fmt.Sprintf("%s:%s", ip, port), logMux); err != nil {
				log.Fatalf("Error while serving requests: %v", err)
			}
		}()
	}
	log.Printf("Listening on %s:%s", ip, ports[0])
	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", ip, ports[0]), logMux); err != nil {
		log.Fatalf("Error while serving requests: %v", err)
	}
}
