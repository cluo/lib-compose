package composition

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func Test_FetchDefinition_NewFetchDefinitionFromRequest(t *testing.T) {
	a := assert.New(t)

	reader := strings.NewReader("the body")
	r, err := http.NewRequest("POST", "https://example.com/content?foo=bar", reader)
	a.NoError(err)

	r.Header = http.Header{
		"Content-Type":     {"text/html"},
		"Cookie":           {"aa=bb;"},
		"X-Feature-Toggle": {"true"},
		"Accept-Encoding":  {"gzip"}, // should not be copied
		"X-Correlation-Id": {"foobar123"},
	}

	fd := NewFetchDefinition("http://upstream:8080/")
	fd.FromRequest(r)
	a.Equal("http://upstream:8080/content?foo=bar", fd.URL)
	a.Equal(10*time.Second, fd.Timeout)
	a.Equal(true, fd.Required)

	a.Equal("text/html", fd.Header.Get("Content-Type"))
	a.Equal("aa=bb;", fd.Header.Get("Cookie"))
	a.Equal("true", fd.Header.Get("X-Feature-Toggle"))
	a.Equal("", fd.Header.Get("Accept-Encoding"))
	a.Equal("foobar123", fd.Header.Get("X-Correlation-Id"))

	a.Equal("POST", fd.Method)
	b, err := ioutil.ReadAll(fd.Body)
	a.NoError(err)
	a.Equal("the body", string(b))
}

func Test_FetchDefinition_use_DefaultErrorHandler_if_not_set(t *testing.T) {
	a := assert.New(t)

	fd := NewFetchDefinition("http://upstream:8080/")
	a.Equal(NewDefaultErrorHandler(), fd.ErrHandler)
}

func Test_FetchDefinition_NewFunctions_have_default_priority(t *testing.T) {
	a := assert.New(t)
	request := &http.Request{}
	request.URL = &url.URL{}

	fd1 := NewFetchDefinition("foo")
	fd2 := NewFetchDefinition("baa").FromRequest(request)
	fd3 := NewFetchDefinition("blub").WithResponseProcessor(nil).FromRequest(request)

	r, err := http.NewRequest("POST", "https://example.com/content?foo=bar", nil)
	a.NoError(err)

	fd4 := NewFetchDefinition("bla").FromRequest(r).WithResponseProcessor(nil)

	a.Equal(fd1.Priority, DefaultPriority)
	a.Equal(fd2.Priority, DefaultPriority)
	a.Equal(fd3.Priority, DefaultPriority)
	a.Equal(fd4.Priority, DefaultPriority)
}

func Test_FetchDefinition_NewFunctions_have_parameter_priority(t *testing.T) {
	a := assert.New(t)
	request := &http.Request{}
	request.URL = &url.URL{}

	fd1 := NewFetchDefinition("foo").WithPriority(42)
	fd2 := NewFetchDefinition("baa").WithPriority(54).FromRequest(request)
	fd3 := NewFetchDefinition("blub").WithResponseProcessor(nil).WithPriority(74).FromRequest(request)

	r, err := http.NewRequest("POST", "https://example.com/content?foo=bar", nil)
	a.NoError(err)

	fd4 := NewFetchDefinition("bla").WithResponseProcessor(nil).WithPriority(90).FromRequest(r)
	fd5 := NewFetchDefinition("faa").FromRequest(r).WithPriority(2014)

	a.Equal(fd1.Priority, 42)
	a.Equal(fd2.Priority, 54)
	a.Equal(fd3.Priority, 74)
	a.Equal(fd4.Priority, 90)
	a.Equal(fd5.Priority, 2014)
}
