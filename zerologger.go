package zerologger

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// New creates a new zerolog handler for Fiber.
//
// The default Logger middleware from Fiber uses buffers and templates and
// writes directly to os.Stderr. This strips out all of that and sends the
// log directly to Zerolog.
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Set variables
	var (
		once       sync.Once
		errHandler fiber.ErrorHandler
	)

	// Return new handler
	return func(ctx *fiber.Ctx) error {
		// Don't execute the middleware if Next returns true
		if cfg.Next != nil && cfg.Next(ctx) {
			return ctx.Next()
		}

		// Set error handler once
		once.Do(func() {
			// override error handler
			errHandler = ctx.App().Config().ErrorHandler
		})

		// Set latency start time
		start := time.Now()

		// Handle request, store err for logging
		chainErr := ctx.Next()

		msg := "OK"

		// Manually call error handler
		if chainErr != nil {
			if err := errHandler(ctx, chainErr); err != nil {
				_ = ctx.SendStatus(fiber.StatusInternalServerError)
			}
			msg = chainErr.Error()
		}

		// Set latency stop time
		stop := time.Now()

		zerolog := log.Logger.With().
			Int("status", ctx.Response().StatusCode()).
			Dur("latency", stop.Sub(start)).
			Str("ip", ctx.IP()).
			Str("method", ctx.Method()).
			Str("path", ctx.Path()).
			Str("user-agent", ctx.Get(fiber.HeaderUserAgent)).
			Logger()

		switch {
		case ctx.Response().StatusCode() == fiber.StatusOK:
			zerolog.Info().Msg(msg)
		case ctx.Response().StatusCode() >= fiber.StatusBadRequest && ctx.Response().StatusCode() < fiber.StatusInternalServerError:
			zerolog.Warn().Msg(msg)
		case ctx.Response().StatusCode() >= fiber.StatusInternalServerError:
			zerolog.Error().Msg(msg)
		default:
			zerolog.Debug().Msg(msg)
		}

		// End chain
		return nil
	}
}
