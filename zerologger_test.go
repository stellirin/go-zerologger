package zerologger_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	. "czechia.dev/zerologger"
)

func testEcho(format ...string) (*bytes.Buffer, *echo.Echo) {
	buf := new(bytes.Buffer)
	e := echo.New()

	if len(format) <= 0 {
		e.Use(New())
		return buf, e
	}

	cfg := Config{
		Format: format,
		Output: buf,
	}

	e.Use(New(cfg))

	return buf, e
}

const echoURI = "http://example.com/info.html?test=true"

func Test_Echo(t *testing.T) {
	_, e := testEcho()

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func Test_TagPid(t *testing.T) {
	buf, e := testEcho(TagPid)
	pid := os.Getpid()

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%d"`, TagPid, pid))
}

func Test_TagTime(t *testing.T) {
	buf, e := testEcho(TagTime)
	// TODO: this format will cause false negatives
	format := time.RFC3339

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagTime, time.Now().Format(format)))
}

func Test_TagReferer(t *testing.T) {
	buf, e := testEcho(TagReferer)
	referer := "test"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set("Referer", referer)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagReferer, referer))
}

func Test_TagProtocol(t *testing.T) {
	buf, e := testEcho(TagProtocol)
	protocol := "HTTP/1.1"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagProtocol, protocol))
}

func Test_TagID(t *testing.T) {
	buf, e := testEcho(TagID)
	id := "test"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set(echo.HeaderXRequestID, id)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagID, id))
}

func Test_TagIP(t *testing.T) {
	buf, e := testEcho(TagIP)
	ip := "192.0.2.1"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagIP, ip))
}

func Test_TagIPs(t *testing.T) {
	buf, e := testEcho(TagIPs)
	ips := "1.2.3.4,5.6.7.8"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set(echo.HeaderXForwardedFor, ips)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagIPs, ips))
}

func Test_TagHost(t *testing.T) {
	buf, e := testEcho(TagHost)
	host := "example.com"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagHost, host))
}

func Test_TagMethod(t *testing.T) {
	buf, e := testEcho(TagMethod)
	method := http.MethodGet

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagMethod, method))
}

func Test_TagPath(t *testing.T) {
	buf, e := testEcho(TagPath)
	path := "/info.html"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagPath, path))
}

func Test_TagURL(t *testing.T) {
	buf, e := testEcho(TagURL)

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagURL, echoURI))
}

func Test_TagUA(t *testing.T) {
	buf, e := testEcho(TagUA)
	ua := "test"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set("User-Agent", ua)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagUA, ua))
}

func Test_TagLatency(t *testing.T) {
	buf, e := testEcho(TagLatency)

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":0.0`, TagLatency))
}

func Test_TagStatus(t *testing.T) {
	buf, e := testEcho(TagStatus)
	status := 404

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":%d`, TagStatus, status))
}

func Test_TagResBody(t *testing.T) {
	t.Skip("TagResBody not implemented on Echo.")

	buf, e := testEcho(TagResBody)
	body := "test"

	e.GET("/body", func(c echo.Context) error {
		return c.String(http.StatusOK, body)
	})

	req := httptest.NewRequest(http.MethodGet, "/body", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagResBody, body))
}

func Test_TagQueryStringParams(t *testing.T) {
	buf, e := testEcho(TagQueryStringParams)
	params := "test=true"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagQueryStringParams, params))
}

func Test_TagBody(t *testing.T) {
	t.Skip("TagBody not implemented on Echo.")

	buf, e := testEcho(TagBody)
	body := "test"

	e.POST("/info.html", func(c echo.Context) error {
		return c.String(http.StatusOK, body)
	})

	r := strings.NewReader("test")
	req := httptest.NewRequest(http.MethodPost, echoURI, r)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagBody, body))
}

func Test_TagBytesSent(t *testing.T) {
	buf, e := testEcho(TagBytesSent)
	status := 24

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":%d`, TagBytesSent, status))
}

func Test_TagBytesReceived(t *testing.T) {
	buf, e := testEcho(TagBytesReceived)
	body := "test"

	e.POST("/info.html", func(c echo.Context) error {
		return c.String(http.StatusOK, body)
	})

	r := strings.NewReader(body)

	req := httptest.NewRequest(http.MethodPost, echoURI, r)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":%d`, TagBytesReceived, 0))

	req = httptest.NewRequest(http.MethodPost, echoURI, r)
	req.Header.Set(echo.HeaderContentLength, fmt.Sprint(len(body)))
	res = httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ = io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":%d`, TagBytesReceived, len(body)))
}

func Test_TagRoute(t *testing.T) {
	buf, e := testEcho(TagRoute)
	route := "/info.html"

	e.GET(route, func(c echo.Context) error {
		return c.String(http.StatusOK, "info.html")
	})

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagRoute, route))
}

func Test_TagError(t *testing.T) {
	buf, e := testEcho(TagError)
	err := "test"

	e.GET("/info.html", func(c echo.Context) error {
		return errors.New(err)
	})

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagError, err))
}

func Test_TagHeader(t *testing.T) {
	header := "X-Test-Header"
	buf, e := testEcho(TagHeader + header)
	value := "test"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set(header, value)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, header, value))
}

func Test_TagLocals(t *testing.T) {
	local := "test"
	buf, e := testEcho(TagLocals + local)

	e.GET("/string", func(c echo.Context) error {
		c.Set("test", "55")
		return c.NoContent(http.StatusOK)
	})

	e.GET("/bytes", func(c echo.Context) error {
		c.Set("test", []byte("55"))
		return c.NoContent(http.StatusOK)
	})

	e.GET("/int", func(c echo.Context) error {
		c.Set("test", 55)
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/string", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"55"`, local))

	req = httptest.NewRequest(http.MethodGet, "/bytes", nil)
	res = httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ = io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"55"`, local))

	req = httptest.NewRequest(http.MethodGet, "/int", nil)
	res = httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ = io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"55"`, local))
}

func Test_TagQuery(t *testing.T) {
	query := "test"
	buf, e := testEcho(TagQuery + query)
	value := "true"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set(query, value)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, query, value))
}

func Test_TagForm(t *testing.T) {
	form := "test"
	value := "true"
	buf, e := testEcho(TagForm + form)

	e.POST("info.html", func(c echo.Context) error {
		c.FormValue(form)
		return c.String(http.StatusOK, "info.html")
	})

	uv := url.Values{}
	uv.Set(form, value)

	req := httptest.NewRequest(http.MethodPost, echoURI, strings.NewReader(uv.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(uv.Encode())))
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, form, value))
}

func Test_TagCookie(t *testing.T) {
	cookie := "test"
	buf, e := testEcho(TagCookie + cookie)
	value := "true"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.AddCookie(&http.Cookie{
		Name:  cookie,
		Value: value,
	})
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, cookie, value))
}

func Test_Next(t *testing.T) {
	e := echo.New()
	e.Use(New(Config{
		Skipper: func(_ echo.Context) bool {
			return true
		},
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func Test_ErrorTimeZone(t *testing.T) {
	e := echo.New()
	e.Use(New(Config{
		TimeZone: "invalid",
		Output:   io.Discard,
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}
func Test_Redirect(t *testing.T) {
	e := echo.New()
	e.Use(New(Config{
		PrettyLatency: true,
		Output:        io.Discard,
	}))

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusPermanentRedirect, "/")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	require.Equal(t, http.StatusPermanentRedirect, res.Code)
}

func Benchmark_Echo(b *testing.B) {
	e := echo.New()
	e.Use(New(Config{
		Format: []string{
			TagBytesReceived, TagBytesSent, TagStatus,
			// TagStatus, TagLatency, TagMethod, TagPath,
			// TagTime, TagStatus, TagLatency, TagMethod, TagPath,
			// TagPid, TagTime, TagReferer, TagProtocol, TagID, TagIP, TagIPs, TagHost, TagMethod, TagPath, TagURL, TagUA, TagLatency, TagStatus, TagResBody, TagQueryStringParams, TagBody, TagBytesSent, TagBytesReceived, TagRoute, TagError, "header:h-test", "locals:l-test", "query:q-test", "form:f-test", "cookie:c-test",
		},
		Output: io.Discard,
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		e.ServeHTTP(res, req)
	}

	require.Equal(b, http.StatusOK, res.Code)
}

// For coverage only
func Test_Initialize(t *testing.T) {
	Initialize(true, true)
}
