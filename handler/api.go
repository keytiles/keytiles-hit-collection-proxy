package handler

import (
	"keytiles-proxy/handler/header"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type APIHandler struct {
	proxy *httputil.ReverseProxy
}

func NewAPIHandler(hosts []string, upstreams []*url.URL, allowedHeaders map[string]any) http.Handler {
	if len(hosts) != len(upstreams) {
		log.Panic("number of hosts and kt upstreams does not match.")
	}

	hostToProxy := make(map[string]*url.URL, len(hosts))
	for i, h := range hosts {
		hostname, err := extractHostname(h)
		if err != nil {
			log.Panicf("invalid host %v", h)
		}
		hostToProxy[hostname] = upstreams[i]
	}

	director := func(req *http.Request) {
		hostname, err := extractHostname(req.Host)
		if err != nil {
			log.Printf("Error: could not extract hostname from request Host %v due to %v", req.Host, err.Error())
			return
		}
		target, ok := hostToProxy[hostname]
		if !ok {
			log.Println("Error: could not find a matching upstream.")
			return
		}

		rewriteRequestURL(req, target)

		// anonymise IP address before forwarding to Keytiles.
		ip := header.AnonymiseIP(req.Header, req.RemoteAddr)
		req.Header.Set(header.XForwardedFor, ip)

		// allow only whitelisted headers to be forwarded.
		header.WhitelistHeaders(req.Header, allowedHeaders)
		log.Printf("Sending to host %v", req.URL.Host)
	}

	return &APIHandler{
		proxy: &httputil.ReverseProxy{Director: director},
	}
}

func (ah *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	ah.proxy.ServeHTTP(w, r)
}

// copied from reverseproxy.go in the std lib.
func rewriteRequestURL(req *http.Request, target *url.URL) {
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
}

// copied from reverseproxy.go in the std lib.
func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

// copied from reverseproxy.go in the std lib.
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// extracts hostname from host.
func extractHostname(host string) (string, error) {
	// handle http://blabla.com or http://blabla.com:8080 case
	if strings.HasPrefix(host, "http") {
		url, err := url.Parse(host)
		if err != nil {
			return "", err
		}
		host = url.Host
	}

	// handle blabla.com case
	if !strings.Contains(host, ":") {
		return host, nil
	}

	// handle blabla.com:8080 case
	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		return "", err
	}
	return hostname, nil
}
