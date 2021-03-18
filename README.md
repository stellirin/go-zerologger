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
