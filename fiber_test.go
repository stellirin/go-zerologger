package zerologger_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	. "czechia.dev/zerologger"
)

func Test_Fiber(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{TagError},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return errors.New("some random error")
	})

	res, _ := app.Test(httptest.NewRequest("GET", "/", nil))
	data, _ := io.ReadAll(buf)

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
	require.Contains(t, string(data), `"error":"some random error"`)
}

func Test_Fiber_locals(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{"locals:demo"},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("demo", "johndoe")
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/int", func(c *fiber.Ctx) error {
		c.Locals("demo", 55)
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/bytes", func(c *fiber.Ctx) error {
		c.Locals("demo", []byte("55"))
		return c.SendStatus(fiber.StatusOK)
	})

	res, _ := app.Test(httptest.NewRequest("GET", "/", nil))
	data, _ := io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(data), `"demo":"johndoe"`)

	res, _ = app.Test(httptest.NewRequest("GET", "/int", nil))
	data, _ = io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(data), `"demo":"55"`)

	res, _ = app.Test(httptest.NewRequest("GET", "/bytes", nil))
	data, _ = io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(data), `"demo":"55"`)
}

func Test_Fiber_Next(t *testing.T) {
	app := fiber.New()
	app.Use(Fiber(Config{
		Next: func(_ *fiber.Ctx) bool {
			return true
		},
	}))

	res, _ := app.Test(httptest.NewRequest("GET", "/", nil))
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}

func Test_Fiber_ErrorTimeZone(t *testing.T) {
	app := fiber.New()
	app.Use(Fiber(Config{
		TimeZone: "invalid",
	}))

	res, _ := app.Test(httptest.NewRequest("GET", "/", nil))
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}

func Test_Fiber_All(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{TagPid, TagID, TagReferer, TagProtocol, TagIP, TagIPs, TagHost, TagMethod, TagPath, TagURL, TagUA, TagStatus, TagResBody, TagQueryStringParams, TagBody, TagBytesSent, TagBytesReceived, TagRoute, TagError, "header:test", "query:test", "form:test", "cookie:test"},
	}))

	res, _ := app.Test(httptest.NewRequest("GET", "/?foo=bar", nil))
	data, _ := io.ReadAll(buf)

	require.Equal(t, http.StatusNotFound, res.StatusCode)

	expected := fmt.Sprintf(`{"level":"warn","pid":"%d","id":"","referer":"","protocol":"http","ip":"0.0.0.0","ips":"","host":"example.com","method":"GET","path":"/","url":"/?foo=bar","ua":"","status":404,"resBody":"Cannot GET /","queryParams":"foo=bar","body":"","bytesSent":12,"bytesReceived":0,"route":"/","test":"","test":"","test":"","test":"","message":"NotFound"}`, os.Getpid())
	require.Equal(t, expected, strings.TrimSpace(string(data)))
}

func Test_Fiber_QueryParams(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{TagQueryStringParams},
	}))

	res, _ := app.Test(httptest.NewRequest("GET", "/?foo=bar&baz=moz", nil))
	data, _ := io.ReadAll(buf)

	require.Equal(t, http.StatusNotFound, res.StatusCode)
	require.Contains(t, string(data), `"queryParams":"foo=bar&baz=moz"`)
}

func Test_Fiber_Response_Body(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{TagResBody},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Sample response body")
	})

	app.Post("/test", func(c *fiber.Ctx) error {
		return c.Send([]byte("Post in test"))
	})

	_, err := app.Test(httptest.NewRequest("GET", "/", nil))
	data, _ := io.ReadAll(buf)

	require.Equal(t, nil, err)
	require.Contains(t, string(data), `"resBody":"Sample response body"`)

	_, err = app.Test(httptest.NewRequest("POST", "/test", nil))
	data, _ = io.ReadAll(buf)

	require.Equal(t, nil, err)
	require.Contains(t, string(data), `"resBody":"Post in test"`)
}

func Test_Fiber_AppendUint(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{
			TagBytesReceived, TagBytesSent, TagStatus,
		},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello")
	})

	res, _ := app.Test(httptest.NewRequest("GET", "/", nil))
	data, _ := io.ReadAll(buf)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(data), `"bytesReceived":0`)
	require.Contains(t, string(data), `"bytesSent":5`)
	require.Contains(t, string(data), `"status":200`)
}

func Test_Fiber_Data_Race(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello")
	})

	var (
		res1, res2 *http.Response
		err1, err2 error
	)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		res1, err1 = app.Test(httptest.NewRequest("GET", "/", nil))
		wg.Done()
	}()
	res2, err2 = app.Test(httptest.NewRequest("GET", "/", nil))
	wg.Wait()

	require.Equal(t, nil, err1)
	require.Equal(t, http.StatusOK, res1.StatusCode)
	require.Equal(t, nil, err2)
	require.Equal(t, http.StatusOK, res2.StatusCode)
}

func Test_Fiber_Redirect(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{
			TagBytesReceived, TagBytesSent, TagStatus,
		},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/", http.StatusPermanentRedirect)
	})

	res, _ := app.Test(httptest.NewRequest("GET", "/", nil))
	require.Equal(t, http.StatusPermanentRedirect, res.StatusCode)
}

func Benchmark_Fiber(b *testing.B) {
	Logger = zerolog.New(io.Discard)

	app := fiber.New()

	app.Use(Fiber(Config{
		Format: []string{
			TagBytesReceived, TagBytesSent, TagStatus,
		},
	}))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	h := app.Handler()

	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.SetMethod("GET")
	fctx.Request.SetRequestURI("/")

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		h(fctx)
	}

	require.Equal(b, 200, fctx.Response.Header.StatusCode())
}

// Dummy test for code coverage
// Initialize is a convenience function only.
// ALWAYS KEEP THIS TEST LAST, INITIALIZE MODIFIES THE GLOBAL STATE.
func Test_Initialize(t *testing.T) {
	Initialize(true, true)
}
