package hue

import "crypto/tls"
import "net"
import "net/http"
import "time"

// Use a global timeout for all client operations
const clientTimeout = 2 * time.Second

func newTimeoutClient() *http.Client {
	transport := http.Transport{
		Dial:                  timeoutDialer,
		DialTLS:               timeoutDialerTLS,
		TLSHandshakeTimeout:   clientTimeout,
		ResponseHeaderTimeout: clientTimeout,
	}

	return &http.Client{
		Transport: &transport,
		Timeout:   clientTimeout,
	}
}

func timeoutDialer(network, addr string) (net.Conn, error) {
	dialer := net.Dialer{Timeout: clientTimeout}
	return dialer.Dial(network, addr)
}

func timeoutDialerTLS(network, addr string) (net.Conn, error) {
	// The hue bridge uses a self-signed certificate
	conf := tls.Config{InsecureSkipVerify: true}
	return tls.Dial(network, addr, &conf)
}
