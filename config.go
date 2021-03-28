package zerologger

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Config defines the Zerolog config for the Fiber middleware.
//
// We use the global Zerolog Logger so (currently) there is nothing to configure.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
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
	if cfg.Next == nil {
		cfg.Next = ConfigDefault.Next
	}

	return cfg
}
