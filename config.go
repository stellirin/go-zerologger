package zerologger

import (
	"io"
	"time"

	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config defines the Zerologger config for the middleware.
type Config struct {

	// Skipper defines a function to skip this middleware when returned true.
	// This field is used only by Echo.
	//
	// Optional. Default: nil
	Skipper middleware.Skipper

	// Format defines the logging tags
	//
	// Optional. Default: []string{TagTime, TagStatus, TagLatency, TagMethod, TagPath}
	Format []string

	// TimeZone can be specified, such as "UTC" and "America/New_York" and "Asia/Chongqing", etc
	//
	// Optional. Default: "Local"
	TimeZone string

	// TimeFormat https://programming.guide/go/format-parse-string-time-date-example.html
	//
	// Optional. Default: time.RFC3339
	TimeFormat string

	// TimeInterval is the delay before the timestamp is updated
	//
	// Optional. Default: 500 * time.Millisecond, Minimum: 500 * time.Millisecond
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

// Helper function to set default values
func setConfig(config ...Config) (cfg Config) {
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set default values
	if cfg.Skipper == nil {
		cfg.Skipper = middleware.DefaultSkipper
	}

	if cfg.Format == nil {
		cfg.Format = []string{TagTime, TagStatus, TagLatency, TagMethod, TagPath}
	}

	if cfg.TimeZone == "" {
		cfg.TimeZone = "Local"
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}
	if cfg.TimeInterval < 500*time.Millisecond {
		cfg.TimeInterval = 500 * time.Millisecond
	}

	cfg.logger = log.Logger
	if cfg.Output != nil {
		cfg.logger = log.Logger.Output(cfg.Output)
	}

	return
}
