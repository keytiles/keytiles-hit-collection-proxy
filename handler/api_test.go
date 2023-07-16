package handler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIEndpoint(t *testing.T) {
	// given
	var (
		mu              sync.Mutex
		capturedHeaders http.Header
	)
	server1 := mockKeytilesAPIResponse(202, "server1", &mu, &capturedHeaders)
	defer server1.Close()
	server2 := mockKeytilesAPIResponse(202, "server2", &mu, &capturedHeaders)
	defer server2.Close()
	hosts := []string{"host1.com", "host2.com:8080"}
	upstreams1, err := url.Parse(server1.URL)
	assert.NoError(t, err)
	upstreams2, err := url.Parse(server2.URL)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/webhits/", strings.NewReader("{}"))
	req.Host = "host1.com"
	req.Header.Set("x-forwarded-for", "192.168.12.12")
	req.Header.Set("x-not-whitelisted-header", "some-value")

	res := httptest.NewRecorder()

	// when
	handler := NewAPIHandler(hosts, []*url.URL{upstreams1, upstreams2}, map[string]any{"x-forwarded-for": nil})
	handler.ServeHTTP(res, req)

	// then
	assertResponse(t, res, 202)

	// server 1 is hit since request host is host1.com.
	assert.Equal(t, "server1", res.Header().Get("x-server-id"))

	// assert header that should be remmoved and the ones should be allowed
	mu.Lock()
	assert.Equal(t, "", capturedHeaders.Get("x-not-whitelisted-header"))
	// last two digits have been masked
	assert.Equal(t, "192.168.12.0", capturedHeaders.Get("x-forwarded-for"))
	mu.Unlock()

	// second hit sent to another host
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/webhits/", strings.NewReader("{}"))
	req.Host = "host2.com:8080"
	res = httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	assertResponse(t, res, 202)
	// server 2 is hit since request host is host2.com:8080.
	assert.Equal(t, "server2", res.Header().Get("x-server-id"))

	// check other status codes
	for _, status := range []int{400, 500} {
		server1 = mockKeytilesAPIResponse(status, "server1", &mu, &capturedHeaders)
		defer server1.Close()

		upstreams1, err := url.Parse(server1.URL)
		assert.NoError(t, err)
		handler := NewAPIHandler([]string{"host1.com"}, []*url.URL{upstreams1}, nil)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/webhits/", strings.NewReader("{}"))
		req.Host = "host1.com"

		res = httptest.NewRecorder()
		handler.ServeHTTP(res, req)

		assertResponse(t, res, status)
	}

}

func assertResponse(t *testing.T, res *httptest.ResponseRecorder, status int) {
	actualRes := res.Result()
	defer actualRes.Body.Close()
	actualBody, err := ioutil.ReadAll(actualRes.Body)
	assert.NoError(t, err)
	assert.Equal(t, getKTAPIResponse(status), string(actualBody))
}

func mockKeytilesAPIResponse(statusCode int, id string, mu *sync.Mutex, capturedHeaders *http.Header) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch statusCode {
		case 202:
			{
				w.Header().Set("content-type", "application/javascript;charset=utf-8")
				w.Header().Set("x-server-id", id)
			}
		case 400:
			{
				w.WriteHeader(http.StatusBadRequest)
			}
		case 500:
			{
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		w.Write([]byte(getKTAPIResponse(statusCode)))
		mu.Lock()
		defer mu.Unlock()
		*capturedHeaders = r.Header
	}))
}

func getKTAPIResponse(status int) string {
	switch status {
	case 202:
		return `{
			"status": "ok",
			"message": "Hit is enqueued for processing"
		}`
	case 400, 500:
		return `{
			"status": "failed"
		}`
	}
	return ""
}
