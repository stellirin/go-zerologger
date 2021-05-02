# Zerologger: logger middleware for Echo

[![reference](https://pkg.go.dev/badge/czechia.dev/zerologger.svg)](https://pkg.go.dev/czechia.dev/zerologger)
[![report](https://goreportcard.com/badge/czechia.dev/zerologger)](https://goreportcard.com/report/czechia.dev/zerologger)
[![tests](https://github.com/stellirin/go-zerologger/workflows/Go/badge.svg)](https://github.com/stellirin/go-zerologger/actions?query=workflow%3AGo)
[![coverage](https://codecov.io/gh/stellirin/go-zerologger/branch/main/graph/badge.svg?token=h5zC6Okqjz)](https://codecov.io/gh/stellirin/go-zerologger)

A simple package to use [Zerolog](https://github.com/rs/zerolog) as the Logger for Echo.

## ⚙️ Installation

```sh
go get -u czechia.dev/zerologger
```

## 📝 Format

Zerologger was inspired by the default Logger middleware in Fiber, replacing the string buffers and templates with Zerolog Events. The key differences are:

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
	"net/http"

	"czechia.dev/zerologger"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Use(zerologger.New(zerologger.Config{
		Format: []string{
			zerologger.TagTime,
			zerologger.TagStatus,
			zerologger.TagLatency,
			zerologger.TagMethod,
			zerologger.TagPath,
		},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World! 👋")
	})

	e.Start(":8080")
}
```

## ⏱ Benchmarks

Zerologger is faster than the default Echo logger and with fewer allocations. Zerologger significantly reduces the latency when logging with Timestamps. It also has the advantage that Zerologger can be configured to produce either structured logs or pretty logs without editing the custom Format string.

Below are some benchmarks with:

1. Minimal format
1. Default format, no time
1. Default format
1. **All** tags enabled, no time
1. **All** tags enabled

Despite some 'large' differences in the results between the two loggers, they both perform *great* and neither will have a noticable impact on your services (your business logic will be orders of magnitude more taxing than the actual logging).

### Results

```txt
goos: darwin
goarch: amd64
pkg: czechia.dev/zerologger
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz

Benchmark_Zerologger/Minimal-8           2927076               389.5 ns/op            91 B/op          2 allocs/op
Benchmark_Echo/Minimal-8                 2093095               561.6 ns/op           153 B/op          3 allocs/op

Benchmark_Zerologger/DefaultNoTime-8     1863060               621.1 ns/op           107 B/op          2 allocs/op
Benchmark_Echo/DefaultNoTime-8           2074165               581.6 ns/op           157 B/op          4 allocs/op

Benchmark_Zerologger/Default-8           1765065               665.6 ns/op           109 B/op          2 allocs/op
Benchmark_Echo/Default-8                 1250151               982.3 ns/op           182 B/op          5 allocs/op

Benchmark_Zerologger/MaximumNoTime-8      874166              1378   ns/op           206 B/op          7 allocs/op
Benchmark_Echo/MaximumNoTime-8            809390              1437   ns/op           266 B/op          9 allocs/op

Benchmark_Zerologger/Maximum-8            768734              1413   ns/op           186 B/op          7 allocs/op
Benchmark_Echo/Maximum-8                  564432              1845   ns/op           284 B/op         10 allocs/op

PASS
ok      czechia.dev/zerologger  14.658s
```
