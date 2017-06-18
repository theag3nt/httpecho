package main

import (
	"reflect"
	"strings"
	"testing"
)

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
