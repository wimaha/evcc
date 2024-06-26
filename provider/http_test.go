package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/evcc-io/evcc/util"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type httpHandler struct {
	val string
	req *http.Request
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.req = req
	h.val = lo.RandomString(16, lo.LettersCharset)
	_, _ = w.Write([]byte(h.val))
}

func TestHttpGet(t *testing.T) {
	h := new(httpHandler)
	srv := httptest.NewServer(h)
	defer srv.Close()

	uri := srv.URL + "/foo/bar"
	p := NewHTTP(util.NewLogger("foo"), http.MethodGet, uri, false, 1, 0)

	uriUrl, _ := url.Parse(uri)

	g, err := p.StringGetter()
	require.NoError(t, err)

	res, err := g()
	require.NoError(t, err)
	assert.Equal(t, uriUrl.Path, h.req.URL.Path)
	assert.Equal(t, h.val, res)
}

func TestHttpSet(t *testing.T) {
	h := new(httpHandler)
	srv := httptest.NewServer(h)
	defer srv.Close()

	uri := srv.URL + "/foo/bar?baz={{.baz}}"
	p := NewHTTP(util.NewLogger("foo"), http.MethodGet, uri, false, 1, 0)

	uriUrl, _ := url.Parse(uri)

	s, err := p.StringSetter("baz")
	require.NoError(t, err)

	err = s("4711")
	require.NoError(t, err)
	assert.Equal(t, uriUrl.Path, h.req.URL.Path)
	assert.Equal(t, "baz=4711", h.req.URL.RawQuery)
}

// templating in url path
func TestHttpSetTemplating(t *testing.T) {
	h := new(httpHandler)
	srv := httptest.NewServer(h)
	defer srv.Close()

	uri := srv.URL + "/foo/bar/{{.baz}}"
	p := NewHTTP(util.NewLogger("foo"), http.MethodGet, uri, false, 1, 0)

	s, err := p.StringSetter("baz")
	require.NoError(t, err)

	err = s("charge_start")
	require.NoError(t, err)

	fmt.Printf("Path: %s\n", h.req.URL.Path)

	assert.Equal(t, "/foo/bar/charge_start", h.req.URL.Path)
}

// test escaping in url path
func TestHttpGetEscaping(t *testing.T) {
	h := new(httpHandler)
	srv := httptest.NewServer(h)
	defer srv.Close()

	uri := srv.URL + "/foo/bar/cm?cmd=Status%208"
	p := NewHTTP(util.NewLogger("foo"), http.MethodGet, uri, false, 1, 0)

	g, err := p.StringGetter()
	require.NoError(t, err)

	res, err := g()
	require.NoError(t, err)

	assert.Equal(t, "/foo/bar/cm", h.req.URL.Path)
	assert.Equal(t, "cmd=Status%208", h.req.URL.RawQuery)
	assert.Equal(t, h.val, res)
}
