package zerologger

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger is a new Zerolog Logger without timestamps.
// Zerologger will hande timestamps according to the Format.
var Logger zerolog.Logger

func init() {
	Logger = zerolog.New(os.Stdout)
}

// New creates a new zerolog middleware for Echo.
//
// The default Logger middleware from Echo uses buffers and templates and
// writes directly to os.Stderr. This strips out all of that and sends the
// log directly to Zerolog.
func New(config ...Config) echo.MiddlewareFunc {
	// Set default config
	cfg := setConfig(config)

	// Get timezone location
	tz, err := time.LoadLocation(cfg.TimeZone)
	if err != nil || tz == nil {
		cfg.timeZoneLocation = time.Local
	} else {
		cfg.timeZoneLocation = tz
	}

	// Check if format contains latency
	for _, tag := range cfg.Format {
		if tag == TagLatency {
			cfg.enableLatency = true
			break
		}
	}

	// Create correct timeformat
	var timestamp atomic.Value
	timestamp.Store(time.Now().In(cfg.timeZoneLocation).Format(cfg.TimeFormat))

	// Update date/time in a separate go routine
	for _, tag := range cfg.Format {
		if tag == TagTime {
			go func() {
				for {
					time.Sleep(cfg.TimeInterval)
					timestamp.Store(time.Now().In(cfg.timeZoneLocation).Format(cfg.TimeFormat))
				}
			}()
			break
		}
	}

	// Set PID once
	pid := strconv.Itoa(os.Getpid())

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// Don't execute the middleware if Next returns true
			if cfg.Skipper(ctx) {
				return next(ctx)
			}

			var start, stop time.Time

			// Set latency start time
			if cfg.enableLatency {
				start = time.Now()
			}

			// Handle request, store err for logging
			chainErr := next(ctx)
			if chainErr != nil {
				ctx.Error(chainErr)
			}

			// Set latency stop time
			if cfg.enableLatency {
				stop = time.Now()
			}

			req := ctx.Request()
			res := ctx.Response()

			status := res.Status

			var event *zerolog.Event
			switch {
			case status == http.StatusOK:
				event = cfg.logger.Info()
			case status >= http.StatusBadRequest && status < http.StatusInternalServerError:
				event = cfg.logger.Warn()
			case status >= http.StatusInternalServerError:
				event = cfg.logger.Error()
			default:
				event = cfg.logger.Debug()
			}

			for _, tag := range cfg.Format {
				switch tag {
				case TagTime:
					event = event.Str(TagTime, timestamp.Load().(string))
				case TagReferer:
					event = event.Str(TagReferer, req.Referer())
				case TagProtocol:
					event = event.Str(TagProtocol, req.Proto)
				case TagPid:
					event = event.Str(TagPid, pid)
				case TagID:
					event = event.Str(TagID, req.Header.Get(echo.HeaderXRequestID))
				case TagIP:
					event = event.Str(TagIP, ctx.RealIP())
				case TagIPs:
					event = event.Str(TagIPs, req.Header.Get(echo.HeaderXForwardedFor))
				case TagHost:
					event = event.Str(TagHost, req.Host)
				case TagPath:
					event = event.Str(TagPath, req.URL.Path)
				case TagURL:
					event = event.Str(TagURL, req.URL.String())
				case TagUA:
					event = event.Str(TagUA, req.UserAgent())
				case TagLatency:
					if cfg.PrettyLatency {
						event = event.Str(TagLatency, stop.Sub(start).String())
					} else {
						event = event.Dur(TagLatency, stop.Sub(start))
					}
				case TagBody:
					// NOOP - Echo doesn't support it
				case TagBytesReceived:
					cl := req.Header.Get(echo.HeaderContentLength)
					if cl == "" {
						event = event.Int(TagBytesReceived, 0)
						continue
					}
					i, _ := strconv.ParseInt(cl, 10, 64)
					event = event.Int64(TagBytesReceived, i)
				case TagBytesSent:
					event = event.Int64(TagBytesSent, res.Size)
				case TagRoute:
					event = event.Str(TagRoute, ctx.Path())
				case TagStatus:
					event = event.Int(TagStatus, status)
				case TagResBody:
					// NOOP - Echo doesn't support it
				case TagQueryStringParams:
					event = event.Str(TagQueryStringParams, req.URL.RawQuery)
				case TagMethod:
					event = event.Str(TagMethod, req.Method)
				case TagError:
					if chainErr != nil {
						event = event.Err(chainErr)
					}
				default:
					// Check if we have a value tag i.e.: "header:x-key"
					switch {
					case strings.HasPrefix(tag, TagHeader):
						event = event.Str(tag[7:], req.Header.Get(tag[7:]))
					case strings.HasPrefix(tag, TagQuery):
						event = event.Str(tag[6:], ctx.QueryParam(tag[6:]))
					case strings.HasPrefix(tag, TagForm):
						event = event.Str(tag[5:], ctx.FormValue(tag[5:]))
					case strings.HasPrefix(tag, TagCookie):
						cookie, err := ctx.Cookie(tag[7:])
						if err == nil {
							event = event.Str(tag[7:], cookie.Value)
						}
					case strings.HasPrefix(tag, TagLocals):
						switch v := ctx.Get(tag[7:]).(type) {
						case []byte:
							event = event.Bytes(tag[7:], v)
						case string:
							event = event.Str(tag[7:], v)
						case nil:
							// NOOP
						default:
							event = event.Str(tag[7:], fmt.Sprintf("%v", v))
						}
					}
				}
			}

			event.Msg(statusMessage[status])

			// End chain
			return nil
		}
	}
}

// Initialize is a convenience function to configure Zerolog with some useful defaults.
func Initialize(debug bool, pretty bool) {
	zerolog.LevelFieldName = "severity"

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if pretty {
		Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	log.Logger = Logger.With().Timestamp().Logger()
}

// Logger variables
const (
	TagPid               = "pid"
	TagTime              = "time"
	TagReferer           = "referer"
	TagProtocol          = "protocol"
	TagID                = "id"
	TagIP                = "ip"
	TagIPs               = "ips"
	TagHost              = "host"
	TagMethod            = "method"
	TagPath              = "path"
	TagURL               = "url"
	TagUA                = "ua"
	TagLatency           = "latency"
	TagStatus            = "status"
	TagResBody           = "resBody"
	TagQueryStringParams = "queryParams"
	TagBody              = "body"
	TagBytesSent         = "bytesSent"
	TagBytesReceived     = "bytesReceived"
	TagRoute             = "route"
	TagError             = "error"
	TagHeader            = "header:"
	TagLocals            = "locals:"
	TagQuery             = "query:"
	TagForm              = "form:"
	TagCookie            = "cookie:"
)

var statusMessage = map[int]string{
	http.StatusContinue:                      "Continue",                        // RFC 7231, 6.2.1
	http.StatusSwitchingProtocols:            "Switching Protocols",             // RFC 7231, 6.2.2
	http.StatusProcessing:                    "Processing",                      // RFC 2518, 10.1
	http.StatusEarlyHints:                    "Early Hints",                     // RFC 8297
	http.StatusOK:                            "OK",                              // RFC 7231, 6.3.1
	http.StatusCreated:                       "Created",                         // RFC 7231, 6.3.2
	http.StatusAccepted:                      "Accepted",                        // RFC 7231, 6.3.3
	http.StatusNonAuthoritativeInfo:          "Non-Authoritative Information",   // RFC 7231, 6.3.4
	http.StatusNoContent:                     "No Content",                      // RFC 7231, 6.3.5
	http.StatusResetContent:                  "Reset Content",                   // RFC 7231, 6.3.6
	http.StatusPartialContent:                "Partial Content",                 // RFC 7233, 4.1
	http.StatusMultiStatus:                   "Multi-Status",                    // RFC 4918, 11.1
	http.StatusAlreadyReported:               "Already Reported",                // RFC 5842, 7.1
	http.StatusIMUsed:                        "IM Used",                         // RFC 3229, 10.4.1
	http.StatusMultipleChoices:               "Multiple Choices",                // RFC 7231, 6.4.1
	http.StatusMovedPermanently:              "Moved Permanently",               // RFC 7231, 6.4.2
	http.StatusFound:                         "Found",                           // RFC 7231, 6.4.3
	http.StatusSeeOther:                      "See Other",                       // RFC 7231, 6.4.4
	http.StatusNotModified:                   "Not Modified",                    // RFC 7232, 4.1
	http.StatusUseProxy:                      "Use Proxy",                       // RFC 7231, 6.4.5
	http.StatusTemporaryRedirect:             "Temporary Redirect",              // RFC 7231, 6.4.7
	http.StatusPermanentRedirect:             "Permanent Redirect",              // RFC 7538, 3
	http.StatusBadRequest:                    "Bad Request",                     // RFC 7231, 6.5.1
	http.StatusUnauthorized:                  "Unauthorized",                    // RFC 7235, 3.1
	http.StatusPaymentRequired:               "Payment Required",                // RFC 7231, 6.5.2
	http.StatusForbidden:                     "Forbidden",                       // RFC 7231, 6.5.3
	http.StatusNotFound:                      "Not Found",                       // RFC 7231, 6.5.4
	http.StatusMethodNotAllowed:              "Method Not Allowed",              // RFC 7231, 6.5.5
	http.StatusNotAcceptable:                 "Not Acceptable",                  // RFC 7231, 6.5.6
	http.StatusProxyAuthRequired:             "Proxy Auth Required",             // RFC 7235, 3.2
	http.StatusRequestTimeout:                "Request Timeout",                 // RFC 7231, 6.5.7
	http.StatusConflict:                      "Conflict",                        // RFC 7231, 6.5.8
	http.StatusGone:                          "Gone",                            // RFC 7231, 6.5.9
	http.StatusLengthRequired:                "Length Required",                 // RFC 7231, 6.5.10
	http.StatusPreconditionFailed:            "Precondition Failed",             // RFC 7232, 4.2
	http.StatusRequestEntityTooLarge:         "Request Entity TooLarge",         // RFC 7231, 6.5.11
	http.StatusRequestURITooLong:             "Request URI Too Long",            // RFC 7231, 6.5.12
	http.StatusUnsupportedMediaType:          "Unsupported Media Type",          // RFC 7231, 6.5.13
	http.StatusRequestedRangeNotSatisfiable:  "Requested Range Not Satisfiable", // RFC 7233, 4.4
	http.StatusExpectationFailed:             "Expectation Failed",              // RFC 7231, 6.5.14
	http.StatusTeapot:                        "Teapot",                          // RFC 7168, 2.3.3
	http.StatusMisdirectedRequest:            "Misdirected Request",             // RFC 7540, 9.1.2
	http.StatusUnprocessableEntity:           "Unprocessable Entity",            // RFC 4918, 11.2
	http.StatusLocked:                        "Locked",                          // RFC 4918, 11.3
	http.StatusFailedDependency:              "Failed Dependency",               // RFC 4918, 11.4
	http.StatusTooEarly:                      "Too Early",                       // RFC 8470, 5.2.
	http.StatusUpgradeRequired:               "Upgrade Required",                // RFC 7231, 6.5.15
	http.StatusPreconditionRequired:          "Precondition Required",           // RFC 6585, 3
	http.StatusTooManyRequests:               "Too Many Requests",               // RFC 6585, 4
	http.StatusRequestHeaderFieldsTooLarge:   "Request Header Fields Too Large", // RFC 6585, 5
	http.StatusUnavailableForLegalReasons:    "Unavailable For Legal Reasons",   // RFC 7725, 3
	http.StatusInternalServerError:           "Internal Server Error",           // RFC 7231, 6.6.1
	http.StatusNotImplemented:                "Not Implemented",                 // RFC 7231, 6.6.2
	http.StatusBadGateway:                    "Bad Gateway",                     // RFC 7231, 6.6.3
	http.StatusServiceUnavailable:            "Service Unavailable",             // RFC 7231, 6.6.4
	http.StatusGatewayTimeout:                "Gateway Timeout",                 // RFC 7231, 6.6.5
	http.StatusHTTPVersionNotSupported:       "HTTP Version Not Supported",      // RFC 7231, 6.6.6
	http.StatusVariantAlsoNegotiates:         "Variant Also Negotiates",         // RFC 2295, 8.1
	http.StatusInsufficientStorage:           "Insufficient Storage",            // RFC 4918, 11.5
	http.StatusLoopDetected:                  "Loop Detected",                   // RFC 5842, 7.2
	http.StatusNotExtended:                   "Not Extended",                    // RFC 2774, 7
	http.StatusNetworkAuthenticationRequired: "Network Authentication Required", // RFC 6585, 6
}
