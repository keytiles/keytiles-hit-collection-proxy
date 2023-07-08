package handler

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type APIHandler struct {
	proxy *httputil.ReverseProxy
	hosts []string
}

func NewAPIHandler(hosts []string, upstreams []*url.URL) http.Handler {
	if len(hosts) != len(upstreams) {
		log.Panic("number of hosts and kt upstreams does not match.")
	}

	hostToProxy := make(map[string]*url.URL, len(hosts))
	for i, h := range hosts {
		hostToProxy[h] = upstreams[i]
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
	proxy := &httputil.ReverseProxy{Director: director}

	return &APIHandler{
		proxy: proxy,
		hosts: hosts,
	}
}

func (ah *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	// r.Host = ah.Remote.Host
	ah.proxy.ServeHTTP(w, r)
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
	if !strings.Contains(host, ":") {
		return host, nil
	}
	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		return "", err
	}
	return hostname, nil
}
