package response

import (
	"encoding/json"
	"net/http"
)

// HTTPResponse represents an API response structure.
type HTTPResponse struct {
	Status string               `json:"status"`
	Data   any                  `json:"data,omitempty"`
	Errors []*HTTPResponseError `json:"errors,omitempty"`
}

// HTTPResponseError represents an API error structure.
type HTTPResponseError struct {
	Description string            `json:"description,omitempty"`
	Validations map[string]string `json:"validations,omitempty"`
}

type Response struct {
	httpResponse *HTTPResponse
	statusCode   int
	isError      bool
}

// Write sends the HTTP response with the appropriate headers and status code.
func (r *Response) Write(w http.ResponseWriter) {
	w.WriteHeader(r.statusCode)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(r.httpResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NewOK(data ...any) *Response { return newResponse(http.StatusOK, data...) }                        // NewOK creates a successful response with optional data.
func NewBadRequest() *Response    { return newErrorResponse(http.StatusBadRequest, ErrBadRequest) }     // NewBadRequest creates a bad request error response.
func NewForbidden() *Response     { return newErrorResponse(http.StatusForbidden, ErrForbidden) }       // NewForbidden creates a forbidden error response.
func NewUnauthorized() *Response  { return newErrorResponse(http.StatusUnauthorized, ErrUnauthorized) } // NewUnauthorized creates an unauthorized error response.

// NewRequestEntityTooLarge creates a request entity too large error response.
func NewRequestEntityTooLarge() *Response {
	return newErrorResponse(http.StatusRequestEntityTooLarge, ErrPayloadTooLarge)
}

// NewUnprocessableEntity creates an unprocessable entity error response.
func NewUnprocessableEntity() *Response {
	return newErrorResponse(http.StatusUnprocessableEntity, ErrUnprocessableEntity)
}

// NewError creates a custom error response with provided status code and errors.
func NewError(statusCode int, errs ...error) *Response {
	// If the status code is not of type error, rewrite to internal server error
	if !IsErrorStatusCode(statusCode) {
		statusCode = http.StatusInternalServerError
	}

	errors := make([]*HTTPResponseError, len(errs))
	for i, err := range errs {
		errors[i] = ErrToHTTPResponseErr(err)
	}

	return &Response{
		isError:    true,
		statusCode: statusCode,
		httpResponse: &HTTPResponse{
			Status: http.StatusText(statusCode),
			Errors: errors,
		},
	}
}

// newResponse creates a new generic response.
func newResponse(statusCode int, data ...any) *Response {
	resp := &Response{
		statusCode: statusCode,
		httpResponse: &HTTPResponse{
			Status: http.StatusText(statusCode),
		},
	}
	if len(data) > 0 {
		resp.httpResponse.Data = data[0]
	}
	return resp
}

// newErrorResponse creates a new error response with a single error.
func newErrorResponse(statusCode int, err error) *Response {
	return &Response{
		isError:    true,
		statusCode: statusCode,
		httpResponse: &HTTPResponse{
			Status: http.StatusText(statusCode),
			Errors: []*HTTPResponseError{
				ErrToHTTPResponseErr(err),
			},
		},
	}
}
