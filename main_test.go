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
		{"v4Localhost_singlePort", "127.0.0.1,8080", "127.0.0.1", "8080", false},
		{"v4Localhost_multiPort", "127.0.0.1,8080,8081", "127.0.0.1", "8080,8081", false},
		{"v4Internal_singlePort", "192.168.0.1,8080", "192.168.0.1", "8080", false},
		{"v4Internal_multiPort", "192.168.0.1,8080,8081", "192.168.0.1", "8080,8081", false},
		{"v6Localhost_singlePort", "::1,8080", "[::1]", "8080", false},
		{"v6Localhost_multiPort", "::1,8080,8081", "[::1]", "8080,8081", false},
		{"v6Internal_singlePort", "fd30:3fac:747b::1,8080", "[fd30:3fac:747b::1]", "8080", false},
		{"v6Internal_multiPort", "fd30:3fac:747b::1,8080,8081", "[fd30:3fac:747b::1]", "8080,8081", false},
		{"v4AllIfaces_singlePort", "0.0.0.0,8080", "0.0.0.0", "8080", false},
		{"v4AllIfaces_multiPort", "0.0.0.0,8080,8081", "0.0.0.0", "8080,8081", false},
		{"v6AllIfaces_singlePort", "::,8080", "[::]", "8080", false},
		{"v6AllIfaces_multiPort", "::,8080,8081", "[::]", "8080,8081", false},
		{"noIfaces_singlePort", "8080", "0.0.0.0", "8080", false},
		{"noIfaces_multiPort", "8080,8081", "0.0.0.0", "8080,8081", false},
		{"negative_tooFewArgs", "", "", "", true},
		{"negative_invalidIPRange", "123.456.789", "", "", true},
		{"negative_invalidIPDomain", "example.com", "", "", true},
		{"negative_invalidPortRange", "127.0.0.1,99999", "", "", true},
		{"negative_invalidPortName", "127.0.0.1,http", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := strings.Split(tt.args, ",")
			if tt.args == "" {
				args = nil
			}
			wantPorts := strings.Split(tt.wantPorts, ",")
			if tt.wantPorts == "" {
				wantPorts = nil
			}

			gotIP, gotPorts, err := validateArgs(args)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIP != tt.wantIP {
				t.Errorf("validateArgs() got IP = %v, want %v", gotIP, tt.wantIP)
			}
			if !reflect.DeepEqual(gotPorts, wantPorts) {
				t.Errorf("validateArgs() got ports = %v, want %v", gotPorts, wantPorts)
			}
		})
	}
}
