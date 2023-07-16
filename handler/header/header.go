package header

import (
	"net"
	"net/http"
	"strings"
)

const (
	XForwardedFor = "X-Forwarded-For"
)

var (
	ipv4Mask = net.IPv4Mask(255, 255, 255, 0)                                                     // /24
	ipv6Mask = net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0} // /64
)

func WhitelistHeaders(reqHeaders http.Header, allowedHeaders map[string]any) {
	for h := range reqHeaders {
		if _, ok := allowedHeaders[strings.ToLower(h)]; !ok {
			reqHeaders.Del(h)
		}
	}
}

func AnonymiseIP(reqHeaders http.Header, remoteAddr string) string {
	forwardedIPs := reqHeaders.Get(XForwardedFor)
	if len(forwardedIPs) > 0 {
		firstIp := strings.Split(forwardedIPs, ",")[0]
		return maskIp(firstIp)
	}

	if strings.Contains(remoteAddr, ":") {
		// get host if remoteAddr is host:port
		remoteAddr = strings.Split(remoteAddr, ":")[0]
	}

	return maskIp(remoteAddr)

}

func maskIp(ip string) string {
	// copied from net.ParseIP()
	for i := 0; i < len(ip); i++ {
		switch ip[i] {
		case '.':
			// ipv4 address
			ip4 := net.ParseIP(ip)
			if ip4 == nil {
				return ""
			}

			return ip4.Mask(ipv4Mask).String()
		case ':':
			// ipv6 address
			ipv6 := net.ParseIP(ip)
			if ipv6 == nil {
				return ""
			}

			return ipv6.Mask(ipv6Mask).String()
		}
	}

	return ""
}
