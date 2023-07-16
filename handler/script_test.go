package handler

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpstreamScriptEndpoint(t *testing.T) {
	for _, tc := range []struct {
		name           string
		mockStatusCode int
		bodyCheck      func(actualBody []byte) bool
		headerCheck    func(header http.Header) bool
	}{
		{
			name:           "200 success",
			mockStatusCode: 200,
			bodyCheck: func(actualBody []byte) bool {
				expectedBody, err := os.ReadFile("testdata/expected-script")
				assert.NoError(t, err)
				return assert.Equal(t, expectedBody, actualBody)
			},
			headerCheck: func(header http.Header) bool {
				return assert.Equal(t, "application/javascript;charset=utf-8", header.Get("content-type"))
			},
		},
		{
			name:           "400 Bad Request",
			mockStatusCode: 400,
			bodyCheck: func(actualBody []byte) bool {
				return true
			},
			headerCheck: func(header http.Header) bool {
				return true
			},
		},
		{
			name:           "500 Internel server error",
			mockStatusCode: 500,
			bodyCheck: func(actualBody []byte) bool {
				return true
			},
			headerCheck: func(header http.Header) bool {
				return true
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given
			server := mockKeytilesScriptResponse(tc.mockStatusCode)
			defer server.Close()
			hosts := []string{"test1.host.com", "test2.host.com"}
			req, _ := http.NewRequest(http.MethodGet, "/tracking/", nil)
			res := httptest.NewRecorder()
			url, _ := url.Parse(server.URL)

			// when
			handler := NewScriptHandler(hosts, url, nil)
			handler.ServeHTTP(res, req)

			// then
			actualRes := res.Result()
			defer actualRes.Body.Close()
			actualBody, err := ioutil.ReadAll(actualRes.Body)
			assert.NoError(t, err)

			assert.Equal(t, tc.mockStatusCode, actualRes.StatusCode)
			assert.True(t, tc.bodyCheck(actualBody))
			assert.True(t, tc.headerCheck(res.Header()))
		})
	}
}

func getAddr() string {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	port := l.Addr().(*net.TCPAddr).Port
	addr := ":" + strconv.Itoa(port)

	return addr
}

func mockKeytilesScriptResponse(statusCode int) *httptest.Server {
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
	}))
}
