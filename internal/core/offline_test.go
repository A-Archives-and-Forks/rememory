package core

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
)

// drandHosts returns the set of hostnames from DrandEndpoints for allowlisting.
func drandHosts() map[string]bool {
	hosts := make(map[string]bool)
	for _, endpoint := range DrandEndpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			continue
		}
		hosts[u.Hostname()] = true
	}
	return hosts
}

// TestMain blocks all network access by default. This catches network calls
// even from libraries that create their own http.Client (like drand-client).
// Same principle as the Playwright offline-by-default fixture in e2e/fixtures.ts.
//
// When REMEMORY_TEST_TLOCK=1, only drand endpoints are allowed.
func TestMain(m *testing.M) {
	allowed := map[string]bool{}
	if os.Getenv("REMEMORY_TEST_TLOCK") == "1" {
		allowed = drandHosts()
	}

	blockDial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _, _ := net.SplitHostPort(addr)
		if allowed[host] {
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		}
		return nil, fmt.Errorf(
			"unexpected network connection: %s %s\n"+
				"Tests are offline by default. Allowed hosts: %v",
			network, addr, allowed,
		)
	}

	http.DefaultTransport = &http.Transport{
		DialContext: blockDial,
	}
	// Force Go's pure-Go DNS resolver so lookups go through our Dial function.
	// Without PreferGo, cgo-based resolution bypasses net.Resolver entirely.
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if len(allowed) > 0 {
				// Allow DNS lookups when we have an allowlist (the dial check
				// above will still block connections to non-allowed hosts).
				return (&net.Dialer{}).DialContext(ctx, network, addr)
			}
			return nil, fmt.Errorf(
				"unexpected DNS lookup: %s %s — tests are offline by default",
				network, addr,
			)
		},
	}
	os.Exit(m.Run())
}
