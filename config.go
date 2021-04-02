package zerologger

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/labstack/echo/v4/middleware"
)

// Config defines the Zerolog config for the Fiber middleware.
//
// We use the global Zerolog Logger so (currently) there is nothing to configure.
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

	enableLatency    bool
	timeZoneLocation *time.Location
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Skipper:      middleware.DefaultSkipper,
	Next:         nil,
	Format:       []string{TagTime, TagStatus, TagLatency, TagMethod, TagPath},
	TimeFormat:   time.RFC3339,
	TimeZone:     "Local",
	TimeInterval: 500 * time.Millisecond,
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	// Set default values
	if cfg.Skipper == nil {
		cfg.Skipper = ConfigDefault.Skipper
	}
	if cfg.Next == nil {
		cfg.Next = ConfigDefault.Next
	}
	if cfg.Format == nil {
		cfg.Format = ConfigDefault.Format
	}
	if cfg.TimeZone == "" {
		cfg.TimeZone = ConfigDefault.TimeZone
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = ConfigDefault.TimeFormat
	}
	if int(cfg.TimeInterval) <= 0 {
		cfg.TimeInterval = ConfigDefault.TimeInterval
	}

	return cfg
}
