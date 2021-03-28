# Zerolog middleware for Fiber

[![codecov](https://codecov.io/gh/stellirin/fiber-zerologger/branch/main/graph/badge.svg?token=3FRCIF5YDW)](https://codecov.io/gh/stellirin/fiber-zerologger)
[![Test Action Status](https://github.com/stellirin/fiber-zerologger/workflows/Go/badge.svg)](https://github.com/stellirin/fiber-zerologger/actions?query=workflow%3AGo)

A simple package to use Zerolog as the Logger for Fiber.

## ⚙️ Installation

```sh
go get -u czechia.dev/zerologger
```

## ⏱ Benchmarks

Zerolog middleware for Fiber is slower than the default Fiber logger, but its main advantage is that it can produce both structured and pretty logs.

Below are some benchmarks with:

1. Default format without time
1. Default format with time
1. All tags enabled

### Zerologger

```txt
goos: darwin
goarch: amd64
pkg: czechia.dev/zerologger
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
Benchmark_Logger-8   	 2296683	       498.0 ns/op	       0 B/op	       0 allocs/op
Benchmark_Logger-8   	 1478088	       807.3 ns/op	       0 B/op	       0 allocs/op
Benchmark_Logger-8   	  749281	      1615   ns/op	       8 B/op	       1 allocs/op
PASS
ok  	czechia.dev/zerologger	1.314s
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

## 👀 Example

```go
package main

import (
	"os"
	"time"

	"czechia.dev/zerologger"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	app := fiber.New()
	app.Use(zerologger.New())

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello, World! 👋")
	})

	app.Listen(":8080")
}
```

## 📝 Format

Zerolog middleware for Fiber is heavily based on the default Logger middleware for Fiber. The key differences are:

* uses a slice of strings to define the log format
* no color output options, zerolog does not support it
* no time format options, uses the global Zerolog time format

The recommended method is to pass in a slice using the provided constants:

```go
app.Use(zerologger.New(zerologger.Config{
	Format: []string{
		zerologger.TagTime,
		zerologger.TagStatus,
		zerologger.TagLatency,
		zerologger.TagMethod,
		zerologger.TagPath,
	},
}))
```

Some constants have a trailing semicolon. These can be used to extract data from the current context, so that `header:x-test-header` will add `"x-test-header": "test-value"` to the log.

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
