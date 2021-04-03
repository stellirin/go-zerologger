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

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	. "czechia.dev/zerologger"
)

func newFiber(format ...string) (*bytes.Buffer, *fiber.App) {
	buf := new(bytes.Buffer)
	app := fiber.New()

	if len(format) <= 0 {
		app.Use(Fiber())
		return buf, app
	}

	cfg := Config{
		Format: format,
		Output: buf,
	}

	app.Use(Fiber(cfg))

	return buf, app
}

const fiberURI = "http://example.com/info.html?test=true"

func Test_Fiber(t *testing.T) {
	_, app := newFiber()

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	res, _ := app.Test(req)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}

func Test_Fiber_TagPid(t *testing.T) {
	buf, app := newFiber(TagPid)
	pid := os.Getpid()

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%d"`, TagPid, pid))
}

func Test_Fiber_TagTime(t *testing.T) {
	buf, app := newFiber(TagTime)
	// TODO: this format will cause false negatives
	format := time.RFC3339

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagTime, time.Now().Format(format)))
}

func Test_Fiber_TagReferer(t *testing.T) {
	buf, app := newFiber(TagReferer)
	referer := "test"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	req.Header.Set(fiber.HeaderReferer, referer)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagReferer, referer))
}

func Test_Fiber_TagProtocol(t *testing.T) {
	buf, app := newFiber(TagProtocol)
	protocol := "http"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagProtocol, protocol))
}

func Test_Fiber_TagID(t *testing.T) {
	buf, app := newFiber(TagID)
	id := "test"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	req.Header.Set(fiber.HeaderXRequestID, id)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagID, id))
}

func Test_Fiber_TagIP(t *testing.T) {
	buf, app := newFiber(TagIP)
	ip := "0.0.0.0"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagIP, ip))
}

func Test_Fiber_TagIPs(t *testing.T) {
	buf, app := newFiber(TagIPs)
	ips := "1.2.3.4,5.6.7.8"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	req.Header.Set(fiber.HeaderXForwardedFor, ips)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagIPs, ips))
}

func Test_Fiber_TagHost(t *testing.T) {
	buf, app := newFiber(TagHost)
	host := "example.com"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagHost, host))
}

func Test_Fiber_TagMethod(t *testing.T) {
	buf, app := newFiber(TagMethod)
	method := http.MethodGet

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagMethod, method))
}

func Test_Fiber_TagPath(t *testing.T) {
	buf, app := newFiber(TagPath)
	path := "/info.html"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagPath, path))
}

func Test_Fiber_TagURL(t *testing.T) {
	buf, app := newFiber(TagURL)

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagURL, fiberURI))
}

func Test_Fiber_TagUA(t *testing.T) {
	buf, app := newFiber(TagUA)
	ua := "test"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	req.Header.Set(fiber.HeaderUserAgent, ua)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagUA, ua))
}

func Test_Fiber_TagLatency(t *testing.T) {
	buf, app := newFiber(TagLatency)

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":0.0`, TagLatency))
}

func Test_Fiber_TagStatus(t *testing.T) {
	buf, app := newFiber(TagStatus)
	status := 404

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":%d`, TagStatus, status))
}

func Test_Fiber_TagResBody(t *testing.T) {
	buf, app := newFiber(TagResBody)
	body := "test"

	app.Get("/body", func(c *fiber.Ctx) error {
		return c.SendString(body)
	})

	req := httptest.NewRequest(http.MethodGet, "/body", nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagResBody, body))
}

func Test_Fiber_TagQueryStringParams(t *testing.T) {
	buf, app := newFiber(TagQueryStringParams)
	params := "test=true"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagQueryStringParams, params))
}

func Test_Fiber_TagBody(t *testing.T) {
	buf, app := newFiber(TagBody)
	body := "test"

	app.Post("/info.html", func(c *fiber.Ctx) error {
		return c.SendString(body)
	})

	r := strings.NewReader("test")
	req := httptest.NewRequest(http.MethodPost, fiberURI, r)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagBody, body))
}

func Test_Fiber_TagBytesSent(t *testing.T) {
	buf, app := newFiber(TagBytesSent)
	status := 21

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":%d`, TagBytesSent, status))
}

func Test_Fiber_TagBytesReceived(t *testing.T) {
	buf, app := newFiber(TagBytesReceived)
	body := "test"

	app.Post("/info.html", func(c *fiber.Ctx) error {
		return c.SendString(body)
	})

	r := strings.NewReader(body)
	req := httptest.NewRequest(http.MethodPost, fiberURI, r)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":%d`, TagBytesReceived, len(body)))
}

func Test_Fiber_TagRoute(t *testing.T) {
	buf, app := newFiber(TagRoute)
	route := "/info.html"

	app.Get(route, func(c *fiber.Ctx) error {
		return c.SendString("info.html")
	})

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagRoute, route))
}

func Test_Fiber_TagError(t *testing.T) {
	buf, app := newFiber(TagError)
	err := "test"

	app.Get("/info.html", func(c *fiber.Ctx) error {
		return errors.New(err)
	})

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	res, _ := app.Test(req)
	data, _ := io.ReadAll(buf)

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, TagError, err))
}

func Test_Fiber_TagHeader(t *testing.T) {
	header := "X-Test-Header"
	buf, app := newFiber(TagHeader + header)
	value := "test"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	req.Header.Set(header, value)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, header, value))
}

func Test_Fiber_TagLocals(t *testing.T) {
	local := "test"
	buf, app := newFiber(TagLocals + local)

	app.Get("/string", func(c *fiber.Ctx) error {
		c.Locals("test", "55")
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/bytes", func(c *fiber.Ctx) error {
		c.Locals("test", []byte("55"))
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/int", func(c *fiber.Ctx) error {
		c.Locals("test", 55)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/string", nil)
	res, _ := app.Test(req)
	data, _ := io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"55"`, local))

	req = httptest.NewRequest(http.MethodGet, "/bytes", nil)
	res, _ = app.Test(req)
	data, _ = io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"55"`, local))

	req = httptest.NewRequest(http.MethodGet, "/int", nil)
	res, _ = app.Test(req)
	data, _ = io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"55"`, local))
}

func Test_Fiber_TagQuery(t *testing.T) {
	query := "test"
	buf, app := newFiber(TagQuery + query)
	value := "true"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	req.Header.Set(query, value)
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, query, value))
}

func Test_Fiber_TagForm(t *testing.T) {
	form := "test"
	value := "true"
	buf, app := newFiber(TagForm + form)

	app.Post("info.html", func(c *fiber.Ctx) error {
		c.FormValue(form)
		return c.SendString("info.html")
	})

	uv := url.Values{}
	uv.Set(form, value)

	req := httptest.NewRequest(http.MethodPost, fiberURI, strings.NewReader(uv.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(uv.Encode())))
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, form, value))
}

func Test_Fiber_TagCookie(t *testing.T) {
	cookie := "test"
	buf, app := newFiber(TagCookie + cookie)
	value := "true"

	req := httptest.NewRequest(http.MethodGet, fiberURI, nil)
	req.AddCookie(&http.Cookie{
		Name:  cookie,
		Value: value,
	})
	app.Test(req)
	data, _ := io.ReadAll(buf)
	require.Contains(t, string(data), fmt.Sprintf(`"%s":"%s"`, cookie, value))
}

func Test_Fiber_Next(t *testing.T) {
	app := fiber.New()
	app.Use(Fiber(Config{
		Next: func(_ *fiber.Ctx) bool {
			return true
		},
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res, _ := app.Test(req)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}

func Test_Fiber_ErrorTimeZone(t *testing.T) {
	app := fiber.New()
	app.Use(Fiber(Config{
		TimeZone: "invalid",
		Output:   io.Discard,
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res, _ := app.Test(req)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}
func Test_Fiber_Redirect(t *testing.T) {
	app := fiber.New()
	app.Use(Fiber(Config{
		PrettyLatency: true,
		Output:        io.Discard,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/", http.StatusPermanentRedirect)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res, _ := app.Test(req)
	require.Equal(t, http.StatusPermanentRedirect, res.StatusCode)
}

func Benchmark_Fiber(b *testing.B) {
	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{
			TagBytesReceived, TagBytesSent, TagStatus,
		},
		Output: io.Discard,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.SetMethod(http.MethodGet)
	fctx.Request.SetRequestURI("/")

	h := app.Handler()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		h(fctx)
	}

	require.Equal(b, http.StatusOK, fctx.Response.Header.StatusCode())
}

// Dummy test for code coverage
// Initialize is a convenience function only.
// ALWAYS KEEP THIS TEST LAST, INITIALIZE MODIFIES THE GLOBAL STATE.
func Test_Initialize(t *testing.T) {
	Initialize(true, true)
}
