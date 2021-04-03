package zerologger

import (
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

// Config defines the Zerologger config for the middleware.
type Config struct {

	// Skipper defines a function to skip this middleware when returned true.
	// This field is used only by Echo.
	//
	// Optional. Default: nil
	Skipper middleware.Skipper

	// Next defines a function to skip this middleware when returned true.
	// This field is used only by Fiber.
	//
	// Optional. Default: nil
	Next func(ctx *fiber.Ctx) bool

	// Format defines the logging tags
	//
	// Optional. Default: 'time status latency method path'
	Format []string

	// TimeFormat https://programming.guide/go/format-parse-string-time-date-example.html
	//
	// Optional. Default: time.RFC3339
	TimeFormat string

	// TimeZone can be specified, such as "UTC" and "America/New_York" and "Asia/Chongqing", etc
	//
	// Optional. Default: "Local"
	TimeZone string

	// TimeInterval is the delay before the timestamp is updated
	//
	// Optional. Default: 500 * time.Millisecond
	TimeInterval time.Duration

	// Output is an io.Writer where logs can be written. Zerologger will copy
	// the global Logger if Output is not set. Typically used in tests.
	//
	// Optional. Default: nil
	Output io.Writer

	// PrettyLatency prints the latency as a string instead of a number.
	//
	// Optional. Default: false
	PrettyLatency bool

	enableLatency    bool
	timeZoneLocation *time.Location
	logger           zerolog.Logger
}

// defaultConfig is the default config
var defaultConfig = Config{
	Skipper:      middleware.DefaultSkipper,
	Next:         nil,
	Format:       []string{TagTime, TagStatus, TagLatency, TagMethod, TagPath},
	TimeFormat:   time.RFC3339,
	TimeZone:     "Local",
	TimeInterval: 500 * time.Millisecond,
}

// Helper function to set default values
func setConfig(config []Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return defaultConfig
	}

	// Override default config
	cfg := config[0]

	// Set default values
	if cfg.Skipper == nil {
		cfg.Skipper = defaultConfig.Skipper
	}
	if cfg.Next == nil {
		cfg.Next = defaultConfig.Next
	}
	if cfg.Format == nil {
		cfg.Format = defaultConfig.Format
	}
	if cfg.TimeZone == "" {
		cfg.TimeZone = defaultConfig.TimeZone
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = defaultConfig.TimeFormat
	}
	if int(cfg.TimeInterval) <= 0 {
		cfg.TimeInterval = defaultConfig.TimeInterval
	}
	if cfg.Output != nil {
		cfg.logger = zerolog.New(cfg.Output)
	} else {
		cfg.logger = Logger
	}

	return cfg
}
