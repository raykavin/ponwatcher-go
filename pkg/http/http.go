// Package http provides interfaces for HTTP server and client operations.
package http

import (
	"context"
	"net/http"
	"time"
)

// RequestContext encapsulates the HTTP request/response cycle and provides
// methods for handling HTTP operations.
type RequestContext interface {
	// Context returns the current context associated with the request.
	Context() context.Context

	// Writer returns the http.ResponseWriter for writing the response.
	Writer() http.ResponseWriter

	// Request returns the original http.Request.
	Request() *http.Request

	// Set stores a value in the request context with the provided key.
	Set(key string, value any)

	// Get retrieves a value from the request context by key.
	// Returns the value and a boolean indicating if the key exists.
	Get(key string) (any, bool)

	// JSON sends a JSON response with the specified status code.
	JSON(statusCode int, data any)

	// BindJSON parses the request body as JSON into the provided struct pointer.
	BindJSON(outPtr any)

	// GetParam returns the value of the URL parameter with the specified key.
	GetParam(key string) string

	// GetQuery returns the value of the first query parameter with the specified key.
	GetQuery(key string) string

	// Redirect sends an HTTP redirect to the specified URL with the given status code.
	Redirect(statusCode int, to string)

	// SetCookie adds an HTTP cookie to the response.
	SetCookie(c *http.Cookie)

	// Abort stops the current request handling chain.
	Abort()

	// Next continues execution to the next handler in the chain.
	Next()
}

// HttpServer represents an HTTP server that can be started and shut down.
type HttpServer interface {
	// Listen starts the HTTP server and begins accepting connections.
	// This method blocks until the server is shut down or an error occurs.
	Listen() error

	// Shutdown gracefully shuts down the server without interrupting active connections.
	// It uses the provided context for timeout control.
	Shutdown(ctx context.Context) error
}

// HttpClient provides methods for making HTTP requests.
type HttpClient interface {
	// Returns the HTTP status code and any error that occurred.
	func(ctx context.Context, url, method string, payload any,
		headers map[string]string, outPtr any, requestTimeout ...time.Duration) (int, error)
}
