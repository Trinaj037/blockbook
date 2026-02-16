//go:build unittest
// +build unittest

package server

import (
	"net/http"
	"testing"
)

func TestCheckOriginAllowAll(t *testing.T) {
	s := &WebsocketServer{}
	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{
			name: "no origin",
			want: true,
		},
		{
			name:   "valid origin",
			origin: "https://example.com",
			want:   true,
		},
		{
			name:   "invalid origin",
			origin: "://bad",
			want:   true,
		},
		{
			name:   "null origin",
			origin: "null",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{Header: make(http.Header)}
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			got := s.checkOrigin(r)
			if got != tt.want {
				t.Fatalf("checkOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckOriginAllowlist(t *testing.T) {
	allowedOrigins := make(map[string]struct{})
	for _, origin := range []string{"https://example.com", "http://localhost:3000"} {
		normalizedOrigin, ok := normalizeOrigin(origin)
		if !ok {
			t.Fatalf("normalizeOrigin(%q) failed", origin)
		}
		allowedOrigins[normalizedOrigin] = struct{}{}
	}
	s := &WebsocketServer{allowedOrigins: allowedOrigins}

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{
			name: "no origin",
			want: true,
		},
		{
			name:   "allowed origin",
			origin: "https://example.com",
			want:   true,
		},
		{
			name:   "allowed origin different case",
			origin: "HTTP://LOCALHOST:3000",
			want:   true,
		},
		{
			name:   "disallowed origin",
			origin: "https://evil.com",
			want:   false,
		},
		{
			name:   "invalid origin",
			origin: "://bad",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{Header: make(http.Header)}
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			got := s.checkOrigin(r)
			if got != tt.want {
				t.Fatalf("checkOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAllowedOrigins(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want []string
	}{
		{
			name: "empty",
			env:  "",
			want: nil,
		},
		{
			name: "valid entries",
			env:  "https://example.com,http://localhost:3000",
			want: []string{"https://example.com", "http://localhost:3000"},
		},
		{
			name: "trims and normalizes",
			env:  " HTTPS://Example.com:9130 , http://LOCALHOST:3000 ",
			want: []string{"https://example.com:9130", "http://localhost:3000"},
		},
		{
			name: "invalid filtered",
			env:  "https://example.com,://bad,",
			want: []string{"https://example.com"},
		},
		{
			name: "all invalid",
			env:  "://bad,not-a-url",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAllowedOrigins("FAKE_WS_ALLOWED_ORIGINS", tt.env)
			if len(got) != len(tt.want) {
				t.Fatalf("parseAllowedOrigins() len = %d, want %d", len(got), len(tt.want))
			}
			for _, origin := range tt.want {
				if _, ok := got[origin]; !ok {
					t.Fatalf("parseAllowedOrigins() missing %q", origin)
				}
			}
		})
	}
}
