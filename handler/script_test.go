package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpstreamScriptEndpoint(t *testing.T) {
	var (
		mu              sync.Mutex
		capturedHeaders http.Header
	)
	for _, tc := range []struct {
		name           string
		mockStatusCode int
		bodyCheck      func(actualBody string) bool
		headerCheck    func(header http.Header) bool
	}{
		{
			name:           "200 success",
			mockStatusCode: 200,
			bodyCheck: func(actualBody string) bool {
				expectedBody, err := os.ReadFile("testdata/expected-script")
				assert.NoError(t, err)
				return assert.Equal(t, string(expectedBody), actualBody)
			},
			headerCheck: func(header http.Header) bool {
				return assert.Equal(t, "application/javascript;charset=utf-8", header.Get("content-type"))
			},
		},
		{
			name:           "400 Bad Request",
			mockStatusCode: 400,
			bodyCheck: func(actualBody string) bool {
				return true
			},
			headerCheck: func(header http.Header) bool {
				return true
			},
		},
		{
			name:           "500 Internel server error",
			mockStatusCode: 500,
			bodyCheck: func(actualBody string) bool {
				return true
			},
			headerCheck: func(header http.Header) bool {
				return true
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given
			server := mockKeytilesScriptResponse(tc.mockStatusCode, &mu, &capturedHeaders)
			defer server.Close()
			hosts := []string{"test1.host.com", "test2.host.com"}
			req, _ := http.NewRequest(http.MethodGet, "/tracking/", nil)
			req.Header.Set("x-forwarded-for", "192.168.12.12")
			req.Header.Set("x-not-whitelisted-header", "some-value")

			res := httptest.NewRecorder()
			url, _ := url.Parse(server.URL)

			// when
			handler := NewScriptHandler(hosts, url, map[string]any{"x-forwarded-for": nil})
			handler.ServeHTTP(res, req)

			// then
			actualRes := res.Result()
			defer actualRes.Body.Close()
			actualBody, err := io.ReadAll(actualRes.Body)
			assert.NoError(t, err)

			assert.Equal(t, tc.mockStatusCode, actualRes.StatusCode)
			assert.True(t, tc.bodyCheck(string(actualBody)))
			assert.True(t, tc.headerCheck(res.Header()))

			// assert header that should be remmoved and the ones should be allowed
			mu.Lock()
			assert.Equal(t, "", capturedHeaders.Get("x-not-whitelisted-header"))
			// last two digits have been masked
			assert.Equal(t, "192.168.12.0", capturedHeaders.Get("x-forwarded-for"))
			mu.Unlock()
		})
	}
}

func mockKeytilesScriptResponse(statusCode int, mu *sync.Mutex, capturedHeaders *http.Header) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch statusCode {
		case 200:
			{
				fileBytes, err := os.ReadFile("testdata/upstream-script")
				if err != nil {
					panic(err)
				}
				w.Header().Set("content-type", "application/javascript;charset=utf-8")
				w.Write(fileBytes)
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
