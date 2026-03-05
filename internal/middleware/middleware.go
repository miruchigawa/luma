package middleware

import (
	"luma/internal/command"
)

type HandlerFunc func(ctx *command.Context) error
type Middleware func(next HandlerFunc) HandlerFunc

// Chain builds a middleware chain. The first middleware in the list is the outermost.
//
// Key design decisions:
// - Standard decorator pattern for composability.
// - Allows cross-cutting concerns (logging, rate limits) without bloating commands.
func Chain(h HandlerFunc, middlewares ...Middleware) HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
