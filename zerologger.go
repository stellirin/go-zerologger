package zerologger

import (
	"fmt"
	"io"
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

// New creates a new zerolog middleware for Echo.
//
// The default Logger middleware from Echo uses buffers and templates and
// writes directly to os.Stderr. This strips out all of that and sends the
// log directly to Zerolog.
func New(config ...Config) echo.MiddlewareFunc {
	// Set default config
	cfg := setConfig(config...)

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

			event.Msg(http.StatusText(status))

			// End chain
			return nil
		}
	}
}

// Initialize is a convenience function to configure Zerolog with some useful defaults.
func Initialize(level string, pretty bool) error {
	Level, err := zerolog.ParseLevel(level)
	if err != nil {
		return err
	}

	if Level == zerolog.NoLevel {
		Level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(Level)

	var w io.Writer = os.Stdout
	if pretty {
		w = zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339}
	}

	log.Logger = zerolog.New(w).With().Timestamp().Logger()

	// GCP Cloud Logging
	zerolog.LevelFieldName = "severity"

	return nil
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
