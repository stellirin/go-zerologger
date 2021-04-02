package zerologger_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	. "czechia.dev/zerologger"
)

func Test_Echo(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	e := echo.New()
	e.Use(Echo(Config{
		Format: []string{TagError},
	}))

	e.GET("/", func(c echo.Context) error {
		return errors.New("some random error")
	})

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)

	out := new(StdOut)
	err := json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, http.StatusInternalServerError, res.Code)
	utils.AssertEqual(t, "some random error", out.Error)
}

func Test_Echo_locals(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	e := echo.New()
	e.Use(Echo(Config{
		Format: []string{"locals:demo"},
	}))

	e.GET("/", func(c echo.Context) error {
		c.Set("demo", "johndoe")
		return c.NoContent(http.StatusOK)
	})

	e.GET("/int", func(c echo.Context) error {
		c.Set("demo", 55)
		return c.NoContent(http.StatusOK)
	})

	e.GET("/bytes", func(c echo.Context) error {
		c.Set("demo", []byte("55"))
		return c.NoContent(http.StatusOK)
	})

	e.GET("/empty", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)

	out := new(StdOut)
	err := json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, http.StatusOK, res.Code)
	utils.AssertEqual(t, "johndoe", out.LocalsDemo)

	req = httptest.NewRequest("GET", "/int", nil)
	res = httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ = io.ReadAll(buf)

	out = new(StdOut)
	err = json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, http.StatusOK, res.Code)
	utils.AssertEqual(t, "55", out.LocalsDemo)

	req = httptest.NewRequest("GET", "/bytes", nil)
	res = httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ = io.ReadAll(buf)

	out = new(StdOut)
	err = json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, http.StatusOK, res.Code)
	utils.AssertEqual(t, "55", out.LocalsDemo)

	req = httptest.NewRequest("GET", "/empty", nil)
	res = httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ = io.ReadAll(buf)

	out = new(StdOut)
	err = json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, http.StatusOK, res.Code)
	utils.AssertEqual(t, "", out.LocalsDemo)
}

func Test_Echo_Next(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)
	e := echo.New()
	e.Use(Echo(Config{
		Skipper: func(_ echo.Context) bool {
			return true
		},
	}))

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	utils.AssertEqual(t, http.StatusNotFound, res.Code)
}

func Test_Echo_ErrorTimeZone(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	e := echo.New()
	e.Use(Echo(Config{
		TimeZone: "invalid",
	}))

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	utils.AssertEqual(t, http.StatusNotFound, res.Code)
}

func Test_Echo_All(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	e := echo.New()
	e.Use(Echo(Config{
		Format: []string{TagPid, TagID, TagReferer, TagProtocol, TagIP, TagIPs, TagHost, TagMethod, TagPath, TagURL, TagUA, TagStatus, TagResBody, TagQueryStringParams, TagBody, TagBytesSent, TagBytesReceived, TagRoute, TagError, "header:test", "query:test", "form:test", "cookie:test"},
	}))

	req := httptest.NewRequest("GET", "/?foo=bar", nil)
	req.Header.Set(echo.HeaderContentLength, "0")
	req.AddCookie(&http.Cookie{
		Name:  "test",
		Value: "test",
	})
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)

	out := new(StdOut)
	err := json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, http.StatusNotFound, res.Code)

	expected := fmt.Sprintf(`{"level":"warn","pid":"%d","id":"","referer":"","protocol":"HTTP/1.1","ip":"192.0.2.1","ips":"","host":"example.com","method":"GET","path":"/","url":"/?foo=bar","ua":"","status":404,"queryParams":"foo=bar","bytesSent":24,"bytesReceived":0,"route":"/","error":"code=404, message=Not Found","test":"","test":"","test":"","test":"test","message":"NotFound"}`, os.Getpid())
	utils.AssertEqual(t, expected, strings.TrimSpace(string(data)))
}

func Test_Echo_QueryParams(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	e := echo.New()
	e.Use(Echo(Config{
		Format: []string{TagQueryStringParams},
	}))

	req := httptest.NewRequest("GET", "/?foo=bar&baz=moz", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)

	out := new(StdOut)
	err := json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, http.StatusNotFound, res.Code)

	expected := "foo=bar&baz=moz"
	utils.AssertEqual(t, expected, out.QueryParams)
}

// func Test_Echo_ResponseBody(t *testing.T) {
// 	buf := new(bytes.Buffer)
// 	Logger = zerolog.New(buf)

// 	e := echo.New()
// 	e.Use(Echo(Config{
// 		Format: []string{TagResBody},
// 	}))

// 	e.GET("/", func(c echo.Context) error {
// 		return c.String(http.StatusOK, "Sample response body")
// 	})

// 	e.POST("/test", func(c echo.Context) error {
// 		return c.String(http.StatusOK, "Post in test")
// 	})

// 	req := httptest.NewRequest("GET", "/", nil)
// 	res := httptest.NewRecorder()
// 	e.ServeHTTP(res, req)
// 	data, _ := io.ReadAll(buf)

// 	out := new(StdOut)
// 	err := json.Unmarshal(data, out)
// 	utils.AssertEqual(t, nil, err)

// 	expectedGetResponse := "Sample response body"
// 	utils.AssertEqual(t, expectedGetResponse, out.ResBody)

// 	req = httptest.NewRequest("POST", "/test", nil)
// 	res = httptest.NewRecorder()
// 	e.ServeHTTP(res, req)
// 	data, _ = io.ReadAll(buf)

// 	out = new(StdOut)
// 	err = json.Unmarshal(data, out)
// 	utils.AssertEqual(t, nil, err)

// 	expectedPostResponse := "Post in test"
// 	utils.AssertEqual(t, expectedPostResponse, out.ResBody)
// }

func Test_Echo_AppendUint(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	e := echo.New()
	e.Use(Echo(Config{
		Format: []string{
			TagBytesReceived, TagBytesSent, TagStatus,
		},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello")
	})

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)

	out := new(StdOut)
	err := json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, http.StatusOK, res.Code)
	utils.AssertEqual(t, 0, out.BytesReceived)
	utils.AssertEqual(t, 5, out.BytesSent)
	utils.AssertEqual(t, 200, out.Status)
}

func Test_Echo_Data_Race(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	e := echo.New()
	e.Use(Echo())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello")
	})

	var (
		res1, res2 *httptest.ResponseRecorder
		err1, err2 error
	)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		req1 := httptest.NewRequest("GET", "/", nil)
		res1 = httptest.NewRecorder()
		e.ServeHTTP(res1, req1)
		wg.Done()
	}()
	req2 := httptest.NewRequest("GET", "/", nil)
	res2 = httptest.NewRecorder()
	e.ServeHTTP(res2, req2)
	wg.Wait()

	utils.AssertEqual(t, nil, err1)
	utils.AssertEqual(t, http.StatusOK, res1.Code)
	utils.AssertEqual(t, nil, err2)
	utils.AssertEqual(t, http.StatusOK, res2.Code)
}

func Test_Echo_Redirect(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	e := echo.New()
	e.Use(Echo(Config{
		Format: []string{
			TagBytesReceived, TagBytesSent, TagStatus,
		},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusPermanentRedirect, "/")
	})

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	utils.AssertEqual(t, http.StatusPermanentRedirect, res.Code)
}

// func Benchmark_Echo(b *testing.B) {
// 	Logger = zerolog.New(io.Discard)

// 	e := echo.New()
// 	e.Use(Echo(Config{
// 		Format: []string{
// 			TagBytesReceived, TagBytesSent, TagStatus,
// 		},
// 	}))
// 	e.GET("/", func(c echo.Context) error {
// 		return c.String(http.StatusOK, "Hello, World!")
// 	})

// 	h := e.Handler()

// 	fctx := &fasthttp.RequestCtx{}
// 	fctx.Request.Header.SetMethod("GET")
// 	fctx.Request.SetRequestURI("/")

// 	b.ReportAllocs()
// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {
// 		h(fctx)
// 	}

// 	utils.AssertEqual(b, 200, fctx.Response.Header.StatusCode())
// }
