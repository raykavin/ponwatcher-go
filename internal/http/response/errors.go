package response

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Erros expostos para acesso direto
var (
	ErrBadRequest             error
	ErrUnprocessableEntity    error
	ErrUnauthorized           error
	ErrForbidden              error
	ErrNotFound               error
	ErrInvalidClient          error
	ErrConflict               error
	ErrGone                   error
	ErrPayloadTooLarge        error
	ErrInternalServer         error
	ErrBadGateway             error
	ErrStatusFailedDependency error
	ErrGatewayTimeout         error

	onceErrors       sync.Once
	onceStatusCodes  sync.Once
	onceStatusErrMap sync.Once
	errStatusCodes   map[error]int
	statusCodeErr    map[int]error
)

func StatusCodeToErr(statusCode int) error {
	if err, ok := getStatusCodeErrMap()[statusCode]; ok {
		return err
	}
	return fmt.Errorf("the HTTP status code: %d was not implemented", statusCode)
}

func ErrToStatusCode(err error) int {
	if statusCode, ok := getStatusCodesMap()[err]; ok {
		return statusCode
	}

	return http.StatusInternalServerError
}

func ErrToHTTPResponseErr(err error) *HTTPResponseError {
	titleCaseConverter := cases.Title(language.AmericanEnglish)
	description := titleCaseConverter.String(err.Error())

	return &HTTPResponseError{
		Description: description,
	}
}

func IsErrorStatusCode(statusCode int) bool {
	return !(statusCode >= http.StatusOK && statusCode <= http.StatusIMUsed)
}

// One-time initialization of errors for direct use
func initializeErrors() {
	onceErrors.Do(func() {
		ErrBadRequest = errors.New("the server could not process the request because the request is malformed")
		ErrUnprocessableEntity = errors.New("unable to process your request, please check for any errors occurred and try again")
		ErrUnauthorized = errors.New("the request you made requires valid authentication credentials, which were not provided or were invalid")
		ErrForbidden = errors.New("you do not have permission to access the resource, please contact your system administrator")
		ErrNotFound = errors.New("the requested resource was not found")
		ErrInvalidClient = errors.New("invalid oauth2 client application")
		ErrConflict = errors.New("a conflict occurred with the request, please review and try again")
		ErrGone = errors.New("the requested resource is no longer available on the server and no forwarding address is known")
		ErrPayloadTooLarge = errors.New("the request exceeds the maximum payload size limit")
		ErrInternalServer = errors.New("an internal server error occurred, please contact support")
		ErrBadGateway = errors.New("unable to process your request at this time. The server received an invalid response from the upstream server it accessed to fulfill your request")
		ErrStatusFailedDependency = errors.New("the operation could not be completed due to a failure of one or more dependencies")
		ErrGatewayTimeout = errors.New("the server could not receive a response from the upstream server within the expected timeframe. This may be due to a temporary network issue or a problem with the upstream server")
	})
}

// Initialize the status code map
func getStatusCodesMap() map[error]int {
	onceStatusCodes.Do(func() {
		initializeErrors() // Ensures that errors have been initialized
		errStatusCodes = map[error]int{
			ErrBadRequest:             http.StatusBadRequest,
			ErrUnauthorized:           http.StatusUnauthorized,
			ErrForbidden:              http.StatusForbidden,
			ErrNotFound:               http.StatusNotFound,
			ErrStatusFailedDependency: http.StatusFailedDependency,
			ErrGone:                   http.StatusGone,
			ErrPayloadTooLarge:        http.StatusRequestEntityTooLarge,
			ErrUnprocessableEntity:    http.StatusUnprocessableEntity,
			ErrInternalServer:         http.StatusInternalServerError,
			ErrBadGateway:             http.StatusBadGateway,
			ErrConflict:               http.StatusConflict,
			ErrGatewayTimeout:         http.StatusGatewayTimeout,
			ErrInvalidClient:          http.StatusUnauthorized,
		}
	})
	return errStatusCodes
}

// Initialize the error map by status code
func getStatusCodeErrMap() map[int]error {
	onceStatusErrMap.Do(func() {
		initializeErrors() // Ensures that errors have been initialized
		statusCodeErr = map[int]error{
			http.StatusBadRequest:            ErrBadRequest,
			http.StatusUnauthorized:          ErrUnauthorized,
			http.StatusForbidden:             ErrForbidden,
			http.StatusNotFound:              ErrNotFound,
			http.StatusGone:                  ErrGone,
			http.StatusRequestEntityTooLarge: ErrPayloadTooLarge,
			http.StatusUnprocessableEntity:   ErrUnprocessableEntity,
			http.StatusInternalServerError:   ErrInternalServer,
			http.StatusBadGateway:            ErrBadGateway,
			http.StatusGatewayTimeout:        ErrGatewayTimeout,
			http.StatusConflict:              ErrConflict,
		}
	})
	return statusCodeErr
}
