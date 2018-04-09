// inspired by github.com/minio/minio-go/transport.go
package hs3

import (
	"net"
	"net/http"
	"time"
)

// DefaultTransport - this default transport is similar to
// http.DefaultTransport but with additional param  DisableCompression
// is set to true to avoid decompressing content with 'gzip' encoding.
var DefaultTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	MaxIdleConnsPerHost:   100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	// Set this value so that the underlying transport round-tripper
	// doesn't try to auto decode the body of objects with
	// content-encoding set to `gzip`.
	//
	// Refer:
	//    https://golang.org/src/net/http/transport.go?h=roundTrip#L1843
	DisableCompression: true,
}
