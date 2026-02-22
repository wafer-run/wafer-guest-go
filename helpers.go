package waffle

import (
	"encoding/json"
	"strconv"
)

// Convenience constructors for common Result values.

// ContinueResult returns a Result that tells the runtime to pass the message
// to the next block in the chain.
func ContinueResult() *Result {
	return &Result{Action: Continue}
}

// RespondResult returns a Result that short-circuits the chain and returns
// the given response to the caller.
func RespondResult(resp *Response) *Result {
	return &Result{Action: Respond, Response: resp}
}

// DropResult returns a Result that ends the chain silently with no response.
func DropResult() *Result {
	return &Result{Action: Drop}
}

// ErrorResult returns a Result that short-circuits the chain with the given error.
func ErrorResult(err *WaffleError) *Result {
	return &Result{Action: ActionError, Err: err}
}

// RespondData creates a Respond result with the given data and optional metadata.
func RespondData(data []byte, meta map[string]string) *Result {
	return &Result{
		Action: Respond,
		Response: &Response{
			Data: data,
			Meta: meta,
		},
	}
}

// JsonRespond creates a Respond result by JSON-encoding the given value. If
// marshaling fails, an internal error result is returned instead.
func JsonRespond(v any) *Result {
	data, err := json.Marshal(v)
	if err != nil {
		return ErrInternal("failed to marshal response: " + err.Error())
	}
	return &Result{
		Action: Respond,
		Response: &Response{
			Data: data,
			Meta: map[string]string{"content-type": "application/json"},
		},
	}
}

// Error creates an error Result with the given code and message.
func Error(code, message string) *Result {
	return &Result{
		Action: ActionError,
		Err: &WaffleError{
			Code:    code,
			Message: message,
		},
	}
}

// ErrorWithMeta creates an error Result with the given code, message, and metadata.
func ErrorWithMeta(code, message string, meta map[string]string) *Result {
	return &Result{
		Action: ActionError,
		Err: &WaffleError{
			Code:    code,
			Message: message,
			Meta:    meta,
		},
	}
}

// Convenience error constructors for common error codes defined in the WAFFLE
// specification. These follow the gRPC status code conventions for
// interoperability.

// ErrBadRequest returns an error result with code "invalid_argument".
func ErrBadRequest(message string) *Result {
	return Error("invalid_argument", message)
}

// ErrNotFound returns an error result with code "not_found".
func ErrNotFound(message string) *Result {
	return Error("not_found", message)
}

// ErrAlreadyExists returns an error result with code "already_exists".
func ErrAlreadyExists(message string) *Result {
	return Error("already_exists", message)
}

// ErrPermissionDenied returns an error result with code "permission_denied".
func ErrPermissionDenied(message string) *Result {
	return Error("permission_denied", message)
}

// ErrUnauthenticated returns an error result with code "unauthenticated".
func ErrUnauthenticated(message string) *Result {
	return Error("unauthenticated", message)
}

// ErrUnavailable returns an error result with code "unavailable".
func ErrUnavailable(message string) *Result {
	return Error("unavailable", message)
}

// ErrDeadlineExceeded returns an error result with code "deadline_exceeded".
func ErrDeadlineExceeded(message string) *Result {
	return Error("deadline_exceeded", message)
}

// ErrResourceExhausted returns an error result with code "resource_exhausted".
func ErrResourceExhausted(message string) *Result {
	return Error("resource_exhausted", message)
}

// ErrFailedPrecondition returns an error result with code "failed_precondition".
func ErrFailedPrecondition(message string) *Result {
	return Error("failed_precondition", message)
}

// ErrInternal returns an error result with code "internal".
func ErrInternal(message string) *Result {
	return Error("internal", message)
}

// RespondWithStatus creates a Respond result with the given HTTP status code,
// data, and content type.
func RespondWithStatus(status int, data []byte, contentType string) *Result {
	return &Result{
		Action: Respond,
		Response: &Response{
			Data: data,
			Meta: map[string]string{
				"resp.status":  strconv.Itoa(status),
				"content-type": contentType,
			},
		},
	}
}

// JsonRespondStatus creates a Respond result by JSON-encoding the given value
// with the specified HTTP status code.
func JsonRespondStatus(status int, v any) *Result {
	data, err := json.Marshal(v)
	if err != nil {
		return ErrInternal("failed to marshal response: " + err.Error())
	}
	return &Result{
		Action: Respond,
		Response: &Response{
			Data: data,
			Meta: map[string]string{
				"resp.status":  strconv.Itoa(status),
				"content-type": "application/json",
			},
		},
	}
}

// ErrorStatus creates an error Result with the given HTTP status code, error code,
// and message.
func ErrorStatus(status int, code, message string) *Result {
	return &Result{
		Action: ActionError,
		Err: &WaffleError{
			Code:    code,
			Message: message,
			Meta: map[string]string{
				"resp.status": strconv.Itoa(status),
			},
		},
	}
}

// ResponseBuilder provides a fluent API for constructing Response values with
// chained method calls.
//
// Example usage:
//
//	result := waffle.NewResponseBuilder().
//	    JSON(map[string]string{"status": "ok"}).
//	    Meta("x-request-id", "abc123").
//	    Respond()
type ResponseBuilder struct {
	data []byte
	meta map[string]string
}

// NewResponseBuilder creates a new ResponseBuilder with an initialized
// metadata map.
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		meta: make(map[string]string),
	}
}

// Data sets the response payload to raw bytes.
func (b *ResponseBuilder) Data(data []byte) *ResponseBuilder {
	b.data = data
	return b
}

// JSON sets the response payload by JSON-encoding the given value and sets
// the content-type metadata to "application/json". If marshaling fails, the
// data is set to nil.
func (b *ResponseBuilder) JSON(v any) *ResponseBuilder {
	data, err := json.Marshal(v)
	if err != nil {
		b.data = nil
		return b
	}
	b.data = data
	b.meta["content-type"] = "application/json"
	return b
}

// Meta sets a metadata key-value pair on the response.
func (b *ResponseBuilder) Meta(key, value string) *ResponseBuilder {
	b.meta[key] = value
	return b
}

// Build constructs the Response from the builder state.
func (b *ResponseBuilder) Build() *Response {
	return &Response{
		Data: b.data,
		Meta: b.meta,
	}
}

// Respond creates a Respond result from the builder state.
func (b *ResponseBuilder) Respond() *Result {
	return &Result{
		Action:   Respond,
		Response: b.Build(),
	}
}
