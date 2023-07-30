package header

import (
	"log"
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
	ip := reqHeaders.Get(XForwardedFor)
	if len(ip) > 0 {
		// get the first ip
		ip = strings.Split(ip, ",")[0]
	}

	if ip == "" {
		ip = remoteAddr
	}

	for i := 0; i < len(ip); i++ {
		switch ip[i] {
		case '.':
			// ipv4 address
			// if port is present
			var err error
			if strings.Contains(ip, ":") {
				ip, _, err = net.SplitHostPort(ip)
				if err != nil {
					log.Println("Error: could not extract host from ip")
					return ""
				}
			}

			ip4 := net.ParseIP(ip)
			if ip4 == nil {
				log.Println("Error: could not parse ip")
				return ""
			}

			return ip4.Mask(ipv4Mask).String()
		case ':':
			// ipv6 address
			// if port is present
			var err error
			if strings.Contains(ip, "]:") {
				ip, _, err = net.SplitHostPort(ip)
				if err != nil {
					log.Println("Error: could not extract host from ip")
					return ""
				}
			}
			// trim [] from ip string
			ip = strings.TrimPrefix(strings.TrimSuffix(ip, "]"), "[")

			ipv6 := net.ParseIP(ip)
			if ipv6 == nil {
				log.Println("Error: could not parse ip")
				return ""
			}

			return ipv6.Mask(ipv6Mask).String()
		}
	}
	return ""
}
