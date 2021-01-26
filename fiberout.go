package fbt

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Timeout struct {
	timeout  time.Duration
	handler  fiber.Handler
	response fiber.Handler
}

type Options func(*Timeout)

func defaultResp(ctx *fiber.Ctx) error {
	return ctx.Status(http.StatusRequestTimeout).SendString(http.StatusText(http.StatusRequestTimeout))
}

func Handler(handler fiber.Handler) Options {
	return func(t *Timeout) {
		t.handler = handler
	}
}

// New
func New(opts ...Options) fiber.Handler {
	const defaultTime = 10 * time.Second

	to := &Timeout{
		timeout:  defaultTime,
		handler:  nil,
		response: defaultResp,
	}

	for _, opt := range opts {
		opt(to)
	}

	if to.timeout <= 0 {
		return to.handler
	}

	return func(ctx *fiber.Ctx) error {
		ch := make(chan struct{}, 1)

		go func() {
			defer func() {
				_ = recover()
			}()
			_ = to.handler(ctx)
			ch <- struct{}{}
		}()

		select {
		case <-ch:
		case <-time.After(to.timeout):
			return fiber.ErrRequestTimeout
		}
		return nil

	}
}
