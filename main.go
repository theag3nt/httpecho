package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
)

func validateArgs(args []string) (string, []string, error) {
	ip := args[1]
	ports := args[2:]

	// Validate IP address
	if parsedIP := net.ParseIP(ip); parsedIP == nil {
		// IP is invalid, maybe it's a port number
		if _, err := strconv.ParseUint(ip, 10, 16); err != nil {
			return "", nil, fmt.Errorf("ip address %#v is invalid", ip)
		}

		// Listen on all interfaces if IP is unspecified
		ip = "0.0.0.0"
		ports = args[1:]
	}

	// Validate port numbers
	for _, port := range ports {
		if _, err := strconv.ParseUint(port, 10, 16); err != nil {
			return "", nil, fmt.Errorf("port number %#v is invalid", port)
		}
	}

	return ip, ports, nil
}

func dumpHandler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(dump); err != nil {
		log.Printf("Unable to write response: %v", err)
	}
}

func logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		remote, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			remote = r.RemoteAddr
		}
		local := r.Host

		log.Printf("%s request from %s on %s\n", method, remote, local)

		h.ServeHTTP(w, r)
	})
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: httpecho [ip] <port> [port]...")
		os.Exit(1)
	}

	ip, ports, err := validateArgs(os.Args)
	if err != nil {
		log.Fatalf("Error while validating arguments: %v", err)
	}

	// Prepare request handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", dumpHandler)
	logMux := logHandler(mux)

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
