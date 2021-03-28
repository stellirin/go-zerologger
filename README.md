# Zerolog middleware for Fiber

[![codecov](https://codecov.io/gh/stellirin/fiber-zerologger/branch/main/graph/badge.svg?token=3FRCIF5YDW)](https://codecov.io/gh/stellirin/fiber-zerologger)
[![Test Action Status](https://github.com/stellirin/fiber-zerologger/workflows/Go/badge.svg)](https://github.com/stellirin/fiber-zerologger/actions?query=workflow%3AGo)

A simple package to use Zerolog as the Logger for Fiber.

## ⚙️ Installation

```sh
go get -u czechia.dev/zerologger
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
