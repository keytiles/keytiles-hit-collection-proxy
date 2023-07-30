package header

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnonymiseIP(t *testing.T) {
	for _, tc := range []struct {
		name       string
		remoteAddr string
		reqHeaders http.Header
		expectedIp string
	}{
		{
			name:       "IPV4 present in request headers",
			remoteAddr: "",
			reqHeaders: map[string][]string{"X-Forwarded-For": {"62.216.217.159"}},
			expectedIp: "62.216.217.0",
		},
		{
			name:       "IPV4 present in request headers with port",
			remoteAddr: "",
			reqHeaders: map[string][]string{"X-Forwarded-For": {"62.216.217.159:8080"}},
			expectedIp: "62.216.217.0",
		},
		{
			name:       "Mulitple IPV4 addresses present in request headers",
			remoteAddr: "",
			reqHeaders: map[string][]string{"X-Forwarded-For": {"62.216.217.159,162.216.217.159"}},
			expectedIp: "62.216.217.0",
		},
		{
			name:       "IPV4 not present in request headers",
			remoteAddr: "62.216.217.159",
			reqHeaders: map[string][]string{},
			expectedIp: "62.216.217.0",
		},
		{
			name:       "IPV6 present in request headers",
			remoteAddr: "",
			reqHeaders: map[string][]string{"X-Forwarded-For": {"[2001:a61:3b1a:4d01:14ca:b270:9d2f:ba4b]"}},
			expectedIp: "2001:a61:3b1a:4d01::",
		},
		{
			name:       "IPV6 present in request headers with port",
			remoteAddr: "",
			reqHeaders: map[string][]string{"X-Forwarded-For": {"[2001:a61:3b1a:4d01:14ca:b270:9d2f:ba4b]:8080"}},
			expectedIp: "2001:a61:3b1a:4d01::",
		},
		{
			name:       "Multiple IPV6 addresses present in request headers",
			remoteAddr: "",
			reqHeaders: map[string][]string{"X-Forwarded-For": {"[2001:a61:3b1a:4d01:14ca:b270:9d2f:ba4b],[2002:a61:3b1a:4d01:14ca:b270:9d2f:ba4b]"}},
			expectedIp: "2001:a61:3b1a:4d01::",
		},
		{
			name:       "IPV6 not present in request headers",
			remoteAddr: "[2001:a61:3b1a:4d01:14ca:b270:9d2f:ba4b]",
			reqHeaders: map[string][]string{},
			expectedIp: "2001:a61:3b1a:4d01::",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ip := AnonymiseIP(tc.reqHeaders, tc.remoteAddr)

			assert.Equal(t, tc.expectedIp, ip)

		})
	}
}
