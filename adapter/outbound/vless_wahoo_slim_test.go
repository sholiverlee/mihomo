//go:build wahoo_slim

package outbound

import (
	"strings"
	"testing"
)

const wahooSlimTestUUID = "00000000-0000-0000-0000-000000000000"
const wahooSlimTestRealityPublicKey = "cFDk2RIjA_9yglRZ2nbHaK52lNPLIY7cQlhZ7Y-CsTg"

func TestWahooSlimVlessSupportsXHTTP(t *testing.T) {
	proxy, err := NewVless(VlessOption{
		Name:    "xhttp",
		Server:  "example.com",
		Port:    443,
		UUID:    wahooSlimTestUUID,
		TLS:     true,
		Network: "xhttp",
		XHTTPOpts: XHTTPOptions{
			Path: "/",
			Host: "example.com",
		},
	})
	if err != nil {
		t.Fatalf("NewVless(xhttp) returned error: %v", err)
	}
	if proxy.xhttpClient == nil {
		t.Fatal("NewVless(xhttp) did not initialize xhttp client")
	}
}

func TestWahooSlimVlessSupportsTCPReality(t *testing.T) {
	proxy, err := NewVless(VlessOption{
		Name:       "tcp-reality",
		Server:     "example.com",
		Port:       443,
		UUID:       wahooSlimTestUUID,
		TLS:        true,
		Network:    "tcp",
		ServerName: "example.com",
		RealityOpts: RealityOptions{
			PublicKey: wahooSlimTestRealityPublicKey,
			ShortID:   "0123456789abcdef",
		},
	})
	if err != nil {
		t.Fatalf("NewVless(tcp reality) returned error: %v", err)
	}
	if proxy.realityConfig == nil {
		t.Fatal("NewVless(tcp reality) did not parse reality config")
	}
	if proxy.xhttpClient != nil {
		t.Fatal("NewVless(tcp reality) should not initialize xhttp client")
	}
}

func TestWahooSlimVlessRejectsTCPWithoutReality(t *testing.T) {
	_, err := NewVless(VlessOption{
		Name:    "tcp",
		Server:  "example.com",
		Port:    443,
		UUID:    wahooSlimTestUUID,
		TLS:     true,
		Network: "tcp",
	})
	if err == nil || !strings.Contains(err.Error(), "tcp requires reality") {
		t.Fatalf("NewVless(tcp without reality) error = %v, want tcp requires reality", err)
	}
}

func TestWahooSlimVlessRejectsOtherNetworks(t *testing.T) {
	_, err := NewVless(VlessOption{
		Name:    "ws",
		Server:  "example.com",
		Port:    443,
		UUID:    wahooSlimTestUUID,
		TLS:     true,
		Network: "ws",
	})
	if err == nil || !strings.Contains(err.Error(), "xhttp and tcp reality only") {
		t.Fatalf("NewVless(ws) error = %v, want unsupported network error", err)
	}
}
