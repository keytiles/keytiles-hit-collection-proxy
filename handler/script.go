package handler

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

const (
	keytilesHost1 = "api.keytiles.com"
	keytilesHost2 = "api2.keytiles.com"
)

type ScriptHandler struct {
	Proxy    *httputil.ReverseProxy
	Upstream *url.URL
}

func NewScriptHandler(hosts []string, upstream *url.URL) http.Handler {
	scriptProxy := httputil.NewSingleHostReverseProxy(upstream)
	scriptProxy.ModifyResponse = func(r *http.Response) error {
		if r.StatusCode == 200 {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				return err
			}
			err = r.Body.Close()
			if err != nil {
				return err
			}
			body = bytes.ReplaceAll(body, []byte(keytilesHost1), []byte(hosts[0]))
			body = bytes.ReplaceAll(body, []byte(keytilesHost2), []byte(hosts[1]))

			r.Body = ioutil.NopCloser(bytes.NewReader(body))
			r.ContentLength = int64(len(body))
			r.Header.Set("Content-Length", strconv.Itoa(len(body)))
		}
		return nil
	}

	return &ScriptHandler{
		Proxy:    scriptProxy,
		Upstream: upstream,
	}
}

func (sh *ScriptHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	//r.Host = sh.Upstream.Host
	sh.Proxy.ServeHTTP(w, r)
}
