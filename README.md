# Zerolog middleware for Fiber

[![codecov](https://codecov.io/gh/stellirin/fiber-zerologger/branch/main/graph/badge.svg?token=3FRCIF5YDW)](https://codecov.io/gh/stellirin/fiber-zerologger)
[![Test Action Status](https://github.com/stellirin/fiber-zerologger/workflows/Go/badge.svg)](https://github.com/stellirin/fiber-zerologger/actions?query=workflow%3AGo)

A simple package to use [Zerolog](https://github.com/rs/zerolog) as the Logger for Fiber.

## ⚙️ Installation

```sh
go get -u czechia.dev/zerologger
```

## 📝 Format

Zerolog middleware for Fiber is heavily based on the default Logger middleware for Fiber. The key differences are:

* uses a slice of strings to define the log format
* no color output options, zerolog does not support it

The recommended method is to pass in a slice using the provided constants:

```go
Format: []string{
	zerologger.TagTime,
	zerologger.TagStatus,
	zerologger.TagLatency,
	zerologger.TagMethod,
	zerologger.TagPath,
}
```

Some constants have a trailing semicolon. These can be used to extract data from the current context, so that `header:x-test-header` will add `"x-test-header": "test-value"` to the log.

## 👀 Example

```go
package main

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"czechia.dev/zerologger"
)

func main() {
	log.Logger = zerolog.New(os.Stdout)

	app := fiber.New()
	app.Use(zerologger.New(zerologger.Config{
		Format: []string{
			zerologger.TagTime,
			zerologger.TagStatus,
			zerologger.TagLatency,
			zerologger.TagMethod,
			zerologger.TagPath,
		},
		TimeFormat:   time.RFC3339,
		TimeZone:     "UTC",
		TimeInterval: 500 * time.Millisecond,
	}))

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello, World! 👋")
	})

	app.Listen(":8080")
}
```

## 🧬 Constants

```go
// Logger variables
const (
	TagPid               = "pid"
	TagTime              = "time"
	TagReferer           = "referer"
	TagProtocol          = "protocol"
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
```

## ⏱ Benchmarks

Zerolog middleware for Fiber is _slightly_ slower than the default Fiber logger. Its main advantage is that Zerolog can be configured to produce both structured and pretty logs.

Below are some benchmarks with:

1. Default format without time
1. Default format with time
1. **All** tags enabled

This shows that printing logs with the Zerolog middleware for Fiber takes approximately 25% longer than the default Logger middleware, with the gap closing as the number of fields increases.

This is however still **extremely** efficient, 500ns is a negligible part of processing a request.

### Zerologger

```txt
goos: darwin
goarch: amd64
pkg: czechia.dev/zerologger
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
Benchmark_Logger-8   	 2258288	       506.4 ns/op	       0 B/op	       0 allocs/op
Benchmark_Logger-8   	 2203069	       536.0 ns/op	       0 B/op	       0 allocs/op
Benchmark_Logger-8   	  862860	      1321   ns/op	       8 B/op	       1 allocs/op
PASS
ok  	czechia.dev/zerologger	1.282s
```

### Logger

```txt
goos: darwin
goarch: amd64
pkg: github.com/gofiber/fiber/v2/middleware/logger
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
Benchmark_Logger-8   	 2918078	       399.8 ns/op	       4 B/op	       1 allocs/op
Benchmark_Logger-8   	 2866551	       419.2 ns/op	       4 B/op	       1 allocs/op
Benchmark_Logger-8   	  938065	      1106   ns/op	      16 B/op	       2 allocs/op
PASS
ok  	github.com/gofiber/fiber/v2/middleware/logger	1.318s
```
