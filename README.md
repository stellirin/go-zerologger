# Zerologger: logger middleware for Echo and Fiber

[![codecov](https://codecov.io/gh/stellirin/go-zerologger/branch/main/graph/badge.svg?token=h5zC6Okqjz)](https://codecov.io/gh/stellirin/go-zerologger)
[![Test Action Status](https://github.com/stellirin/go-zerologger/workflows/Go/badge.svg)](https://github.com/stellirin/go-zerologger/actions?query=workflow%3AGo)

A simple package to use [Zerolog](https://github.com/rs/zerolog) as the Logger for Echo or Fiber.

## ⚙️ Installation

```sh
go get -u czechia.dev/zerologger
```

## 📝 Format

Zerologger is based on the default Logger middleware for Fiber. The key differences are:

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
	"czechia.dev/zerologger"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	app.Use(zerologger.Fiber(zerologger.Config{
		Format: []string{
			zerologger.TagTime,
			zerologger.TagStatus,
			zerologger.TagLatency,
			zerologger.TagMethod,
			zerologger.TagPath,
		},
	}))

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello, World! 👋")
	})

	app.Listen(":8080")
}
```

## ⏱ Benchmarks

Zerologger is _slightly_ slower than the default Fiber logger. Its main advantage is that Zerolog can be configured to produce both structured and pretty logs.

Below are some benchmarks with:

1. Benchmark format
1. Default format without time
1. Default format with time
1. **All** tags enabled

This shows that printing logs with Zerologger takes approximately 25% longer than the default Fiber Logger middleware, with the gap closing as the number of fields increases.

This is however still **extremely** efficient, 500ns is a negligible part of processing a request.

### Zerologger

```txt
goos: darwin
goarch: amd64
pkg: czechia.dev/zerologger
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
Benchmark_Logger-8   	 4527990	       264.4 ns/op	       0 B/op	       0 allocs/op
Benchmark_Logger-8   	 2258288	       506.4 ns/op	       0 B/op	       0 allocs/op
Benchmark_Logger-8   	 2203069	       536.0 ns/op	       0 B/op	       0 allocs/op
Benchmark_Logger-8   	  862860	      1321   ns/op	       8 B/op	       1 allocs/op
PASS
```

### Fiber Logger

```txt
goos: darwin
goarch: amd64
pkg: github.com/gofiber/fiber/v2/middleware/logger
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
Benchmark_Logger-8   	 5769978	       206.2 ns/op	       0 B/op	       0 allocs/op
Benchmark_Logger-8   	 2918078	       399.8 ns/op	       4 B/op	       1 allocs/op
Benchmark_Logger-8   	 2866551	       419.2 ns/op	       4 B/op	       1 allocs/op
Benchmark_Logger-8   	  938065	      1106   ns/op	      16 B/op	       2 allocs/op
PASS
```
