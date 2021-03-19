package zerologger_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"czechia.dev/zerologger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
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

func Test_Logger(t *testing.T) {
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
		app.Use(zerologger.New())
		app.Get("/", func(ctx *fiber.Ctx) error {
			ctx.WriteString(tt.name)
			return fiber.NewError(tt.args.status, tt.name)
		})

		resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, tt.args.status, resp.StatusCode)

		data, _ := io.ReadAll(buf)
		json.Unmarshal(data, tt.args.out)
		utils.AssertEqual(t, zerologger.StatusMessage[tt.args.status], tt.args.out.Message)

		app.Shutdown()
	}
}

func Test_LoggerConfig(t *testing.T) {
	type args struct {
		out    *stdout
		config []zerologger.Config
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
				config: []zerologger.Config{zerologger.ConfigDefault},
				status: fiber.StatusOK,
			},
		},
		{
			name: "false",
			args: args{
				out:    new(stdout),
				config: []zerologger.Config{{Next: func(ctx *fiber.Ctx) bool { return false }}},
				status: fiber.StatusOK,
			},
		},
		{
			name: "true",
			args: args{
				out:    new(stdout),
				config: []zerologger.Config{{Next: func(ctx *fiber.Ctx) bool { return true }}},
				status: 0,
			},
		},
	}

	for _, tt := range tests {
		app := fiber.New()
		app.Use(zerologger.New(tt.args.config...))
		app.Get("/", func(ctx *fiber.Ctx) error {
			ctx.WriteString(tt.name)
			return fiber.NewError(fiber.StatusOK)
		})

		resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

		data, _ := io.ReadAll(buf)
		json.Unmarshal(data, tt.args.out)
		utils.AssertEqual(t, zerologger.StatusMessage[tt.args.status], tt.args.out.Message)

		app.Shutdown()
	}
}

func Benchmark_Logger(b *testing.B) {
	log.Logger = zerolog.New(io.Discard).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(zerologger.New())

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
