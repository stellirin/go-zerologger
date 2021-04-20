# Zerologger: logger middleware for Echo and Fiber

[![coverage](https://codecov.io/gh/stellirin/go-zerologger/branch/main/graph/badge.svg?token=h5zC6Okqjz)](https://codecov.io/gh/stellirin/go-zerologger)
[![tests](https://github.com/stellirin/go-zerologger/workflows/Go/badge.svg)](https://github.com/stellirin/go-zerologger/actions?query=workflow%3AGo)
[![report](https://goreportcard.com/badge/czechia.dev/zerologger)](https://goreportcard.com/report/czechia.dev/zerologger)

A simple package to use [Zerolog](https://github.com/rs/zerolog) as the Logger for Echo or Fiber.

## ⚙️ Installation

```sh
go get -u czechia.dev/zerologger
```

## 📝 Format

Zerologger was inspired by the default Logger middleware for Fiber, replacing the string buffers with Zerolog Events. The key differences are:

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

Some constants have a trailing semicolon. These can be used to extract data from the current context, so that `header:X-Test-Header` will add `"X-Test-Header": "test-value"` to the log.

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

Zerologger is _slightly_ slower than the default Fiber logger, but faster than the default Echo logger. Its main advantage over both is that Zerologger can be configured to produce either structured logs or pretty logs without editing the custom Format string.

Below are some benchmarks with:

1. Benchmark format
1. Default format without time
1. Default format with time
1. **All** tags enabled

Despite some 'large' differences in the results between the three loggers, they all perform *great* and none of them will have a noticable impact on your services (your business logic will be orders of magnitude more taxing than the actual logging).

### Zerologger

```txt
goos: darwin
goarch: amd64
pkg: czechia.dev/zerologger
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
Benchmark_Echo-8    	 3066121	       378.6 ns/op	      90 B/op	       2 allocs/op
Benchmark_Echo-8    	 1908175	       603.6 ns/op	     106 B/op	       2 allocs/op
Benchmark_Echo-8    	 1866398	       640.0 ns/op	     107 B/op	       2 allocs/op
Benchmark_Echo-8    	  729704	      1594   ns/op	     187 B/op	       7 allocs/op
PASS

Benchmark_Fiber-8   	 4939312	       245.8 ns/op	       0 B/op	       0 allocs/op
Benchmark_Fiber-8   	 2335804	       483.1 ns/op	       0 B/op	       0 allocs/op
Benchmark_Fiber-8   	 2293968	       515.6 ns/op	       0 B/op	       0 allocs/op
Benchmark_Fiber-8   	  935451	      1273   ns/op	       8 B/op	       1 allocs/op
PASS
```

### Echo Logger

```txt
goos: darwin
goarch: amd64
pkg: github.com/labstack/echo/v4/middleware
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
Benchmark_Logger-8   	 2232074	       537.0 ns/op	     151 B/op	       3 allocs/op
Benchmark_Logger-8   	 2129727	       565.8 ns/op	     156 B/op	       4 allocs/op
Benchmark_Logger-8   	 1255450	       919.6 ns/op	     182 B/op	       5 allocs/op
Benchmark_Logger-8   	  624400	      1827   ns/op	     280 B/op	      10 allocs/op
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
