package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"keytiles-proxy/handler"
)

var DEAFULT_PORT = ":9999"

type config struct {
	port  string
	hosts []string

	// KT config
	ktScriptUpstream *url.URL
	ktAPIUpstreams   []*url.URL

	// TLS config
	cert string
	key  string
}

func main() {

	c := config{
		port:             readPort(),
		hosts:            readHosts("HOSTS"),
		ktScriptUpstream: hostToURL(os.Getenv("KT_SCRIPT_HOST")),
		ktAPIUpstreams:   hostsToURLs(readHosts("KT_API_HOSTS")),
		cert:             os.Getenv("TLS_CERT"),
		key:              os.Getenv("CERT_KEY"),
	}
	log.Printf("starting keytiles proxy server on port %v...", c.port)
	run(c)
}

func run(c config) {
	mux := http.NewServeMux()
	mux.Handle("/tracking/", handler.NewScriptHandler(c.hosts, c.ktScriptUpstream))
	mux.Handle("/", handler.NewAPIHandler(c.hosts, c.ktAPIUpstreams))

	var err error
	if len(c.cert) > 1 && len(c.key) > 1 {
		err = http.ListenAndServeTLS(c.port, c.cert, c.key, mux)
	} else {
		err = http.ListenAndServe(c.port, mux)
	}

	if err != nil {
		panic(err)
	}
}

func readHosts(env string) []string {
	hostnames := os.Getenv(env)
	hosts := strings.Split(hostnames, ",")

	if len(hosts) == 0 {
		panic("Atleast one host name is required.")
	}
	if len(hosts) == 1 {
		return []string{hosts[0], hosts[1]}
	}
	if len(hosts) > 2 {
		log.Println("More than two host names are not supported. Only first two host names would be used.")
	}

	return hosts[:2]
}

func hostsToURLs(hosts []string) []*url.URL {
	urls := make([]*url.URL, 0, len(hosts))
	for _, h := range hosts {
		urls = append(urls, hostToURL(h))
	}
	return urls
}

func hostToURL(host string) *url.URL {
	url, err := url.Parse(host)
	if err != nil {
		panic(err)
	}
	return url
}

func readPort() string {
	port := os.Getenv("PORT")
	if len(port) < 1 {
		return DEAFULT_PORT
	}

	return ":" + port
}
