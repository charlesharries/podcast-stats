package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golangcollege/sessions"
)

// newTestApplication generates a dummy application struct
// containing some mocked application dependencies.
func newTestApplication(t *testing.T) *application {
	session := sessions.New([]byte("S4fcFbWc5caesR3d6ddSbGxvyzy31IIf"))
	session.Lifetime = 12 * time.Hour

	return &application{
		errorLog: log.New(ioutil.Discard, "", 0),
		infoLog:  log.New(ioutil.Discard, "", 0),
		session:  session,
	}
}

// testServer embeds an httptest.Server instance to allow us to
// get and post to our handlers.
type testServer struct {
	*httptest.Server
}

// newTestServer generates an instance of our testServer type.
func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	ts.Client().Jar = jar

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

// get makes a GET request with our testServer, returning the status
// code, headers, and response body.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()

	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, body
}
