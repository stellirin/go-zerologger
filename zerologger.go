package zerologger

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
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

		// Manually call error handler
		if chainErr != nil {
			if err := errHandler(ctx, chainErr); err != nil {
				_ = ctx.SendStatus(fiber.StatusInternalServerError)
			}
		}

		// Set latency stop time
		stop := time.Now()

		status := ctx.Response().StatusCode()

		var event *zerolog.Event
		switch {
		case status == fiber.StatusOK:
			event = log.Info()
		case status >= fiber.StatusBadRequest && status < fiber.StatusInternalServerError:
			event = log.Warn()
		case status >= fiber.StatusInternalServerError:
			event = log.Error()
		default:
			event = log.Debug()
		}

		event.
			Int("status", status).
			Dur("latency", stop.Sub(start)).
			Str("method", ctx.Method()).
			Str("path", ctx.Path()).
			Msg(StatusMessage[status])

		// End chain
		return nil
	}
}

var StatusMessage = map[int]string{
	fiber.StatusContinue:                      "Continue",                      // RFC 7231, 6.2.1
	fiber.StatusSwitchingProtocols:            "SwitchingProtocols",            // RFC 7231, 6.2.2
	fiber.StatusProcessing:                    "Processing",                    // RFC 2518, 10.1
	fiber.StatusEarlyHints:                    "EarlyHints",                    // RFC 8297
	fiber.StatusOK:                            "OK",                            // RFC 7231, 6.3.1
	fiber.StatusCreated:                       "Created",                       // RFC 7231, 6.3.2
	fiber.StatusAccepted:                      "Accepted",                      // RFC 7231, 6.3.3
	fiber.StatusNonAuthoritativeInformation:   "NonAuthoritativeInformation",   // RFC 7231, 6.3.4
	fiber.StatusNoContent:                     "NoContent",                     // RFC 7231, 6.3.5
	fiber.StatusResetContent:                  "ResetContent",                  // RFC 7231, 6.3.6
	fiber.StatusPartialContent:                "PartialContent",                // RFC 7233, 4.1
	fiber.StatusMultiStatus:                   "MultiStatus",                   // RFC 4918, 11.1
	fiber.StatusAlreadyReported:               "AlreadyReported",               // RFC 5842, 7.1
	fiber.StatusIMUsed:                        "IMUsed",                        // RFC 3229, 10.4.1
	fiber.StatusMultipleChoices:               "MultipleChoices",               // RFC 7231, 6.4.1
	fiber.StatusMovedPermanently:              "MovedPermanently",              // RFC 7231, 6.4.2
	fiber.StatusFound:                         "Found",                         // RFC 7231, 6.4.3
	fiber.StatusSeeOther:                      "SeeOther",                      // RFC 7231, 6.4.4
	fiber.StatusNotModified:                   "NotModified",                   // RFC 7232, 4.1
	fiber.StatusUseProxy:                      "UseProxy",                      // RFC 7231, 6.4.5
	fiber.StatusTemporaryRedirect:             "TemporaryRedirect",             // RFC 7231, 6.4.7
	fiber.StatusPermanentRedirect:             "PermanentRedirect",             // RFC 7538, 3
	fiber.StatusBadRequest:                    "BadRequest",                    // RFC 7231, 6.5.1
	fiber.StatusUnauthorized:                  "Unauthorized",                  // RFC 7235, 3.1
	fiber.StatusPaymentRequired:               "PaymentRequired",               // RFC 7231, 6.5.2
	fiber.StatusForbidden:                     "Forbidden",                     // RFC 7231, 6.5.3
	fiber.StatusNotFound:                      "NotFound",                      // RFC 7231, 6.5.4
	fiber.StatusMethodNotAllowed:              "MethodNotAllowed",              // RFC 7231, 6.5.5
	fiber.StatusNotAcceptable:                 "NotAcceptable",                 // RFC 7231, 6.5.6
	fiber.StatusProxyAuthRequired:             "ProxyAuthRequired",             // RFC 7235, 3.2
	fiber.StatusRequestTimeout:                "RequestTimeout",                // RFC 7231, 6.5.7
	fiber.StatusConflict:                      "Conflict",                      // RFC 7231, 6.5.8
	fiber.StatusGone:                          "Gone",                          // RFC 7231, 6.5.9
	fiber.StatusLengthRequired:                "LengthRequired",                // RFC 7231, 6.5.10
	fiber.StatusPreconditionFailed:            "PreconditionFailed",            // RFC 7232, 4.2
	fiber.StatusRequestEntityTooLarge:         "RequestEntityTooLarge",         // RFC 7231, 6.5.11
	fiber.StatusRequestURITooLong:             "RequestURITooLong",             // RFC 7231, 6.5.12
	fiber.StatusUnsupportedMediaType:          "UnsupportedMediaType",          // RFC 7231, 6.5.13
	fiber.StatusRequestedRangeNotSatisfiable:  "RequestedRangeNotSatisfiable",  // RFC 7233, 4.4
	fiber.StatusExpectationFailed:             "ExpectationFailed",             // RFC 7231, 6.5.14
	fiber.StatusTeapot:                        "Teapot",                        // RFC 7168, 2.3.3
	fiber.StatusMisdirectedRequest:            "MisdirectedRequest",            // RFC 7540, 9.1.2
	fiber.StatusUnprocessableEntity:           "UnprocessableEntity",           // RFC 4918, 11.2
	fiber.StatusLocked:                        "Locked",                        // RFC 4918, 11.3
	fiber.StatusFailedDependency:              "FailedDependency",              // RFC 4918, 11.4
	fiber.StatusTooEarly:                      "TooEarly",                      // RFC 8470, 5.2.
	fiber.StatusUpgradeRequired:               "UpgradeRequired",               // RFC 7231, 6.5.15
	fiber.StatusPreconditionRequired:          "PreconditionRequired",          // RFC 6585, 3
	fiber.StatusTooManyRequests:               "TooManyRequests",               // RFC 6585, 4
	fiber.StatusRequestHeaderFieldsTooLarge:   "RequestHeaderFieldsTooLarge",   // RFC 6585, 5
	fiber.StatusUnavailableForLegalReasons:    "UnavailableForLegalReasons",    // RFC 7725, 3
	fiber.StatusInternalServerError:           "InternalServerError",           // RFC 7231, 6.6.1
	fiber.StatusNotImplemented:                "NotImplemented",                // RFC 7231, 6.6.2
	fiber.StatusBadGateway:                    "BadGateway",                    // RFC 7231, 6.6.3
	fiber.StatusServiceUnavailable:            "ServiceUnavailable",            // RFC 7231, 6.6.4
	fiber.StatusGatewayTimeout:                "GatewayTimeout",                // RFC 7231, 6.6.5
	fiber.StatusHTTPVersionNotSupported:       "HTTPVersionNotSupported",       // RFC 7231, 6.6.6
	fiber.StatusVariantAlsoNegotiates:         "VariantAlsoNegotiates",         // RFC 2295, 8.1
	fiber.StatusInsufficientStorage:           "InsufficientStorage",           // RFC 4918, 11.5
	fiber.StatusLoopDetected:                  "LoopDetected",                  // RFC 5842, 7.2
	fiber.StatusNotExtended:                   "NotExtended",                   // RFC 2774, 7
	fiber.StatusNetworkAuthenticationRequired: "NetworkAuthenticationRequired", // RFC 6585, 6
}
