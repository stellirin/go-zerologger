package zerologger_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	. "czechia.dev/zerologger"
)

type stdout struct {
	Time    time.Time `json:"time"`
	Status  int       `json:"status"`
	Level   string    `json:"level"`
	Latency float32   `json:"latency"`
	IP      string    `json:"ip"`
	Method  string    `json:"method"`
	Path    string    `json:"path"`
	Agent   string    `json:"user-agent"`
	Message string    `json:"message"`
}

func Test_LoggerStatus(t *testing.T) {
	type args struct {
		out    *stdout
		status int
	}

	buf := new(bytes.Buffer)
	log.Logger = log.Output(buf)

	tests := []struct {
		name string
		args args
	}{
		{
			name: "StatusContinue",
			args: args{
				out:    new(stdout),
				status: fiber.StatusContinue,
			},
		},
		{
			name: "StatusOK",
			args: args{
				out:    new(stdout),
				status: fiber.StatusOK,
			},
		},
		{
			name: "StatusBadRequest",
			args: args{
				out:    new(stdout),
				status: fiber.StatusBadRequest,
			},
		},
		{
			name: "StatusInternalServerError",
			args: args{
				out:    new(stdout),
				status: fiber.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		app := fiber.New()
		app.Use(New())
		app.Get("/", func(ctx *fiber.Ctx) error {
			ctx.WriteString(tt.name)
			return fiber.NewError(tt.args.status, tt.name)
		})

		resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, tt.args.status, resp.StatusCode)

		data, _ := io.ReadAll(buf)
		json.Unmarshal(data, tt.args.out)
		utils.AssertEqual(t, StatusMessage[tt.args.status], tt.args.out.Message)

		app.Shutdown()
	}
}

func Test_LoggerConfig(t *testing.T) {
	type args struct {
		out    *stdout
		config []Config
		status int
	}

	buf := new(bytes.Buffer)
	log.Logger = log.Output(buf)

	tests := []struct {
		name string
		args args
	}{
		{
			name: "default",
			args: args{
				out:    new(stdout),
				config: []Config{ConfigDefault},
				status: fiber.StatusOK,
			},
		},
		{
			name: "false",
			args: args{
				out:    new(stdout),
				config: []Config{{Next: func(ctx *fiber.Ctx) bool { return false }}},
				status: fiber.StatusOK,
			},
		},
		{
			name: "true",
			args: args{
				out:    new(stdout),
				config: []Config{{Next: func(ctx *fiber.Ctx) bool { return true }}},
				status: 0,
			},
		},
		{
			name: "full",
			args: args{
				out: new(stdout),
				config: []Config{{
					Format: []string{TagPid, TagTime, TagReferer, TagProtocol, TagIP, TagIPs, TagHost, TagMethod, TagPath, TagURL, TagUA, TagLatency, TagStatus, TagResBody, TagQueryStringParams, TagBody, TagBytesSent, TagBytesReceived, TagRoute, TagError, "header:x-test", "query:q-test", "form:f-test", "cookie:c-test", "locals:l-bytes", "locals:l-string", "locals:l-uuid"},
				}},
				status: fiber.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		app := fiber.New()
		app.Use(New(tt.args.config...))
		app.Get("/", func(ctx *fiber.Ctx) error {
			ctx.Locals("l-bytes", []byte("test"))
			ctx.Locals("l-string", "test")
			ctx.Locals("l-uuid", uuid.Nil)
			ctx.WriteString(tt.name)
			return fiber.NewError(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", "/?q-test=test", nil)
		req.Header = map[string][]string{
			"x-test": {"test"},
		}

		resp, err := app.Test(req)
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

		data, _ := io.ReadAll(buf)
		json.Unmarshal(data, tt.args.out)
		utils.AssertEqual(t, StatusMessage[tt.args.status], tt.args.out.Message)

		app.Shutdown()
	}
}

func Benchmark_Logger(b *testing.B) {
	log.Logger = zerolog.New(io.Discard).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	h := app.Handler()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/")

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		h(ctx)
	}

	utils.AssertEqual(b, 200, ctx.Response.Header.StatusCode())
}
