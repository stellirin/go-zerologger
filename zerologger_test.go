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

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"

	. "czechia.dev/zerologger"
)

type StdOut struct {
	Severity      string `json:"severity"`
	Pid           string `json:"pid"`
	Time          string `json:"time"`
	Referer       string `json:"referer"`
	Protocol      string `json:"protocol"`
	IP            string `json:"ip"`
	IPs           string `json:"ips"`
	Host          string `json:"host"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	URL           string `json:"url"`
	UA            string `json:"ua"`
	Latency       string `json:"latency"`
	Status        int    `json:"status"`
	ResBody       string `json:"resBody"`
	QueryParams   string `json:"queryParams"`
	Body          string `json:"body"`
	BytesSent     int    `json:"bytesSent"`
	BytesReceived int    `json:"bytesReceived"`
	Route         string `json:"route"`
	Error         string `json:"error"`

	LocalsDemo string `json:"demo"`
}

func Test_Logger(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{TagError},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return errors.New("some random error")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	data, _ := io.ReadAll(buf)

	out := new(StdOut)
	json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusInternalServerError, resp.StatusCode)
	utils.AssertEqual(t, "some random error", out.Error)
}

func Test_Logger_locals(t *testing.T) {
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

	app.Get("/empty", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	data, _ := io.ReadAll(buf)

	out := new(StdOut)
	json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "johndoe", out.LocalsDemo)

	resp, err = app.Test(httptest.NewRequest("GET", "/int", nil))
	data, _ = io.ReadAll(buf)

	out = new(StdOut)
	json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "55", out.LocalsDemo)

	resp, err = app.Test(httptest.NewRequest("GET", "/bytes", nil))
	data, _ = io.ReadAll(buf)

	out = new(StdOut)
	json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "55", out.LocalsDemo)

	resp, err = app.Test(httptest.NewRequest("GET", "/empty", nil))
	data, _ = io.ReadAll(buf)

	out = new(StdOut)
	json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "", out.LocalsDemo)
}

func Test_Logger_Next(t *testing.T) {
	app := fiber.New()
	app.Use(Fiber(Config{
		Next: func(_ *fiber.Ctx) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusNotFound, resp.StatusCode)
}

func Test_Logger_ErrorTimeZone(t *testing.T) {
	app := fiber.New()
	app.Use(Fiber(Config{
		TimeZone: "invalid",
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusNotFound, resp.StatusCode)
}

func Test_Logger_All(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{TagPid, TagTime, TagReferer, TagProtocol, TagIP, TagIPs, TagHost, TagMethod, TagPath, TagURL, TagUA, TagStatus, TagResBody, TagQueryStringParams, TagBody, TagBytesSent, TagBytesReceived, TagRoute, TagError, "header:test", "query:test", "form:test", "cookie:test"},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/?foo=bar", nil))
	data, _ := io.ReadAll(buf)

	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusNotFound, resp.StatusCode)

	expected := fmt.Sprintf(`{"level":"warn","pid":"%d","time":"","referer":"","protocol":"http","ip":"0.0.0.0","ips":"","host":"example.com","method":"GET","path":"/","url":"/?foo=bar","ua":"","status":404,"resBody":"Cannot GET /","queryParams":"foo=bar","body":"","bytesSent":12,"bytesReceived":0,"route":"/","test":"","test":"","test":"","test":"","message":"NotFound"}`, os.Getpid())
	utils.AssertEqual(t, expected, strings.TrimSpace(string(data)))
}

func Test_Query_Params(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{TagQueryStringParams},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/?foo=bar&baz=moz", nil))
	data, _ := io.ReadAll(buf)

	out := new(StdOut)
	json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusNotFound, resp.StatusCode)

	expected := "foo=bar&baz=moz"
	utils.AssertEqual(t, expected, out.QueryParams)
}

func Test_Response_Body(t *testing.T) {
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

	out := new(StdOut)
	json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)

	expectedGetResponse := "Sample response body"
	utils.AssertEqual(t, expectedGetResponse, out.ResBody)

	_, err = app.Test(httptest.NewRequest("POST", "/test", nil))
	data, _ = io.ReadAll(buf)

	out = new(StdOut)
	json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)

	expectedPostResponse := "Post in test"
	utils.AssertEqual(t, expectedPostResponse, out.ResBody)
}

func Test_Logger_AppendUint(t *testing.T) {
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

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	data, _ := io.ReadAll(buf)

	out := new(StdOut)
	json.Unmarshal(data, out)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, 0, out.BytesReceived)
	utils.AssertEqual(t, 5, out.BytesSent)
	utils.AssertEqual(t, 200, out.Status)
}

func Test_Logger_Data_Race(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello")
	})

	var (
		resp1, resp2 *http.Response
		err1, err2   error
	)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		resp1, err1 = app.Test(httptest.NewRequest("GET", "/", nil))
		wg.Done()
	}()
	resp2, err2 = app.Test(httptest.NewRequest("GET", "/", nil))
	wg.Wait()

	utils.AssertEqual(t, nil, err1)
	utils.AssertEqual(t, fiber.StatusOK, resp1.StatusCode)
	utils.AssertEqual(t, nil, err2)
	utils.AssertEqual(t, fiber.StatusOK, resp2.StatusCode)
}

func Test_Logger_Redirect(t *testing.T) {
	buf := new(bytes.Buffer)
	Logger = zerolog.New(buf)

	app := fiber.New()
	app.Use(Fiber(Config{
		Format: []string{
			TagBytesReceived, TagBytesSent, TagStatus,
		},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/", fiber.StatusContinue)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusContinue, resp.StatusCode)
}

func Benchmark_Logger(b *testing.B) {
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

	utils.AssertEqual(b, 200, fctx.Response.Header.StatusCode())
}

// Dummy test for code coverage
// Initialize is a convenience function only.
// ALWAYS KEEP THIS TEST LAST, INITIALIZE MODIFIES THE GLOBAL STATE.
func Test_Initialize(t *testing.T) {
	Initialize(true, true)
}
