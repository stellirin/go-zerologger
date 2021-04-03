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

func newEcho(format ...string) (*bytes.Buffer, *echo.Echo) {
	buf := new(bytes.Buffer)
	app := echo.New()

	if len(format) <= 0 {
		app.Use(Echo())
		return buf, app
	}

	cfg := Config{
		Format: format,
		Output: buf,
	}

	app.Use(Echo(cfg))

	return buf, app
}

const echoURI = "http://example.com/info.html?test=true"

func Test_Echo(t *testing.T) {
	_, app := newEcho()

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func Test_Echo_TagPid(t *testing.T) {
	buf, app := newEcho(TagPid)
	pid := os.Getpid()

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%d"`, TagPid, pid))
}

func Test_Echo_TagTime(t *testing.T) {
	buf, app := newEcho(TagTime)
	// TODO: this format will cause false negatives
	format := time.RFC3339

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagTime, time.Now().Format(format)))
}

func Test_Echo_TagReferer(t *testing.T) {
	buf, app := newEcho(TagReferer)
	referer := "test"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set("Referer", referer)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagReferer, referer))
}

func Test_Echo_TagProtocol(t *testing.T) {
	buf, app := newEcho(TagProtocol)
	protocol := "HTTP/1.1"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagProtocol, protocol))
}

func Test_Echo_TagID(t *testing.T) {
	buf, app := newEcho(TagID)
	id := "test"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set(echo.HeaderXRequestID, id)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagID, id))
}

func Test_Echo_TagIP(t *testing.T) {
	buf, app := newEcho(TagIP)
	ip := "192.0.2.1"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagIP, ip))
}

func Test_Echo_TagIPs(t *testing.T) {
	buf, app := newEcho(TagIPs)
	ips := "1.2.3.4,5.6.7.8"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set(echo.HeaderXForwardedFor, ips)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagIPs, ips))
}

func Test_Echo_TagHost(t *testing.T) {
	buf, app := newEcho(TagHost)
	host := "example.com"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagHost, host))
}

func Test_Echo_TagMethod(t *testing.T) {
	buf, app := newEcho(TagMethod)
	method := http.MethodGet

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagMethod, method))
}

func Test_Echo_TagPath(t *testing.T) {
	buf, app := newEcho(TagPath)
	path := "/info.html"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagPath, path))
}

func Test_Echo_TagURL(t *testing.T) {
	buf, app := newEcho(TagURL)

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagURL, echoURI))
}

func Test_Echo_TagUA(t *testing.T) {
	buf, app := newEcho(TagUA)
	ua := "test"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set("User-Agent", ua)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagUA, ua))
}

func Test_Echo_TagLatency(t *testing.T) {
	buf, app := newEcho(TagLatency)

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":0.0`, TagLatency))
}

func Test_Echo_TagStatus(t *testing.T) {
	buf, app := newEcho(TagStatus)
	status := 404

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":%d`, TagStatus, status))
}

func Test_Echo_TagResBody(t *testing.T) {
	t.Skip("TagResBody not implemented on Echo.")

	buf, app := newEcho(TagResBody)
	body := "test"

	app.GET("/body", func(c echo.Context) error {
		return c.String(http.StatusOK, body)
	})

	req := httptest.NewRequest(http.MethodGet, "/body", nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagResBody, body))
}

func Test_Echo_TagQueryStringParams(t *testing.T) {
	buf, app := newEcho(TagQueryStringParams)
	params := "test=true"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagQueryStringParams, params))
}

func Test_Echo_TagBody(t *testing.T) {
	t.Skip("TagBody not implemented on Echo.")

	buf, app := newEcho(TagBody)
	body := "test"

	app.POST("/info.html", func(c echo.Context) error {
		return c.String(http.StatusOK, body)
	})

	r := strings.NewReader("test")
	req := httptest.NewRequest(http.MethodPost, echoURI, r)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagBody, body))
}

func Test_Echo_TagBytesSent(t *testing.T) {
	buf, app := newEcho(TagBytesSent)
	status := 24

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":%d`, TagBytesSent, status))
}

func Test_Echo_TagBytesReceived(t *testing.T) {
	buf, app := newEcho(TagBytesReceived)
	body := "test"

	app.POST("/info.html", func(c echo.Context) error {
		return c.String(http.StatusOK, body)
	})

	r := strings.NewReader(body)

	req := httptest.NewRequest(http.MethodPost, echoURI, r)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":%d`, TagBytesReceived, 0))

	req = httptest.NewRequest(http.MethodPost, echoURI, r)
	req.Header.Set(echo.HeaderContentLength, fmt.Sprint(len(body)))
	res = httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ = io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":%d`, TagBytesReceived, len(body)))
}

func Test_Echo_TagRoute(t *testing.T) {
	buf, app := newEcho(TagRoute)
	route := "/info.html"

	app.GET(route, func(c echo.Context) error {
		return c.String(http.StatusOK, "info.html")
	})

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagRoute, route))
}

func Test_Echo_TagError(t *testing.T) {
	buf, app := newEcho(TagError)
	err := "test"

	app.GET("/info.html", func(c echo.Context) error {
		return errors.New(err)
	})

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagError, err))
}

func Test_Echo_TagHeader(t *testing.T) {
	header := "X-Test-Header"
	buf, app := newEcho(TagHeader + header)
	value := "test"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set(header, value)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, header, value))
}

func Test_Echo_TagLocals(t *testing.T) {
	local := "test"
	buf, app := newEcho(TagLocals + local)

	app.GET("/string", func(c echo.Context) error {
		c.Set("test", "55")
		return c.NoContent(http.StatusOK)
	})

	app.GET("/bytes", func(c echo.Context) error {
		c.Set("test", []byte("55"))
		return c.NoContent(http.StatusOK)
	})

	app.GET("/int", func(c echo.Context) error {
		c.Set("test", 55)
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/string", nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"55"`, local))

	req = httptest.NewRequest(http.MethodGet, "/bytes", nil)
	res = httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ = io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"55"`, local))

	req = httptest.NewRequest(http.MethodGet, "/int", nil)
	res = httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ = io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"55"`, local))
}

func Test_Echo_TagQuery(t *testing.T) {
	query := "test"
	buf, app := newEcho(TagQuery + query)
	value := "true"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.Header.Set(query, value)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, query, value))
}

func Test_Echo_TagForm(t *testing.T) {
	form := "test"
	value := "true"
	buf, app := newEcho(TagForm + form)

	app.POST("info.html", func(c echo.Context) error {
		c.FormValue(form)
		return c.String(http.StatusOK, "info.html")
	})

	uv := url.Values{}
	uv.Set(form, value)

	req := httptest.NewRequest(http.MethodPost, echoURI, strings.NewReader(uv.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(uv.Encode())))
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, form, value))
}

func Test_Echo_TagCookie(t *testing.T) {
	cookie := "test"
	buf, app := newEcho(TagCookie + cookie)
	value := "true"

	req := httptest.NewRequest(http.MethodGet, echoURI, nil)
	req.AddCookie(&http.Cookie{
		Name:  cookie,
		Value: value,
	})
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, cookie, value))
}

func Test_Echo_Next(t *testing.T) {
	app := echo.New()
	app.Use(Echo(Config{
		Skipper: func(_ echo.Context) bool {
			return true
		},
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func Test_Echo_ErrorTimeZone(t *testing.T) {
	app := echo.New()
	app.Use(Echo(Config{
		TimeZone: "invalid",
		Output:   io.Discard,
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}
func Test_Echo_Redirect(t *testing.T) {
	app := echo.New()
	app.Use(Echo(Config{
		PrettyLatency: true,
		Output:        io.Discard,
	}))

	app.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusPermanentRedirect, "/")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	require.Equal(t, http.StatusPermanentRedirect, res.Code)
}

func Benchmark_Echo(b *testing.B) {
	app := echo.New()
	app.Use(Echo(Config{
		Format: []string{
			TagBytesReceived, TagBytesSent, TagStatus,
		},
		Output: io.Discard,
	}))

	app.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		app.ServeHTTP(res, req)
	}

	require.Equal(b, http.StatusOK, res.Code)
}
