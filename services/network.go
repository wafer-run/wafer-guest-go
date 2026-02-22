package services

import (
	"encoding/json"

	wafer "github.com/anthropics/wafer-sdk-go"
)

// NetworkClient provides typed access to the WAFER network capability for
// making outbound HTTP requests. All requests are proxied through the runtime
// for security and observability.
type NetworkClient struct {
	ctx *wafer.Context
}

// NewNetworkClient creates a new NetworkClient bound to the given context.
func NewNetworkClient(ctx *wafer.Context) *NetworkClient {
	return &NetworkClient{ctx: ctx}
}

// NetworkRequest describes an outbound HTTP request to be executed by the
// runtime.
type NetworkRequest struct {
	// Method is the HTTP method (GET, POST, PUT, DELETE, etc.).
	Method string `json:"method"`

	// URL is the fully qualified request URL.
	URL string `json:"url"`

	// Headers holds the HTTP request headers.
	Headers map[string]string `json:"headers,omitempty"`

	// Body is the raw request body.
	Body []byte `json:"body,omitempty"`
}

// NetworkResponse holds the result of an outbound HTTP request.
type NetworkResponse struct {
	// StatusCode is the HTTP response status code.
	StatusCode int `json:"status_code"`

	// Headers holds the HTTP response headers.
	Headers map[string]string `json:"headers,omitempty"`

	// Body is the raw response body.
	Body []byte `json:"body"`
}

// Do executes an outbound HTTP request through the runtime. The request is
// JSON-encoded and sent as the message data. Returns the parsed response.
//
// Message kind: "svc.network.do"
// Data: JSON-encoded NetworkRequest
func (n *NetworkClient) Do(req *NetworkRequest) (*NetworkResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, &wafer.WaferError{
			Code:    "internal",
			Message: "failed to marshal network request: " + err.Error(),
		}
	}
	msg := &wafer.Message{
		Kind: "svc.network.do",
		Data: data,
		Meta: map[string]string{
			"method": req.Method,
			"url":    req.URL,
		},
	}
	result := n.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return nil, result.Err
	}
	if result.Response == nil || len(result.Response.Data) == 0 {
		return nil, &wafer.WaferError{
			Code:    "internal",
			Message: "empty network response",
		}
	}
	var resp NetworkResponse
	if err := json.Unmarshal(result.Response.Data, &resp); err != nil {
		return nil, &wafer.WaferError{
			Code:    "internal",
			Message: "failed to unmarshal network response: " + err.Error(),
		}
	}
	return &resp, nil
}
