package wafer

import (
	"encoding/base64"
)

// Wire format types for the WASM boundary. These types define the JSON
// serialization format used to pass data between the guest WASM module and
// the WAFER host runtime. Binary data is base64-encoded and metadata is
// represented as arrays of two-element string arrays for cross-language
// compatibility.

// WasmMessage is the wire format for a Message. It is serialized to JSON
// and passed across the WASM boundary.
type WasmMessage struct {
	// Kind identifies the message type.
	Kind string `json:"kind"`

	// Data is the base64-encoded binary payload.
	Data string `json:"data"`

	// Meta is a list of key-value pairs represented as [key, value] arrays.
	Meta [][]string `json:"meta"`
}

// WasmResult is the wire format for a Result.
type WasmResult struct {
	// Action is the string representation of the action, e.g. "continue",
	// "respond", "drop", or "error".
	Action string `json:"action"`

	// Response holds the response payload when Action is "respond".
	Response *WasmResponse `json:"response,omitempty"`

	// Error holds the error payload when Action is "error".
	Error *WasmError `json:"error,omitempty"`
}

// WasmResponse is the wire format for a Response.
type WasmResponse struct {
	// Data is the base64-encoded binary payload.
	Data string `json:"data"`

	// Meta is a list of key-value pairs represented as [key, value] arrays.
	Meta [][]string `json:"meta"`
}

// WasmError is the wire format for a WaferError.
type WasmError struct {
	// Code is the machine-readable error code.
	Code string `json:"code"`

	// Message is the human-readable error description.
	Message string `json:"message"`

	// Meta is a list of key-value pairs represented as [key, value] arrays.
	Meta [][]string `json:"meta"`
}

// WasmBlockInfo is the wire format for a BlockInfo.
type WasmBlockInfo struct {
	// Name is the block identifier.
	Name string `json:"name"`

	// Version is the semantic version string.
	Version string `json:"version"`

	// Interface declares the contract this block implements.
	Interface string `json:"interface"`

	// Summary is a brief description of this block implementation.
	Summary string `json:"summary"`

	// InstanceMode is the default instance lifecycle as a string.
	InstanceMode string `json:"instance_mode"`

	// AllowedModes lists the supported instance modes as strings.
	AllowedModes []string `json:"allowed_modes"`
}

// WasmLifecycleEvent is the wire format for a LifecycleEvent.
type WasmLifecycleEvent struct {
	// Type is the lifecycle event type as a string, e.g. "init", "start", "stop".
	Type string `json:"type"`

	// Data is the base64-encoded event payload.
	Data string `json:"data"`
}

// Conversion functions between domain types and wire format types.

// messageToWasm converts a Message to its wire format representation.
func messageToWasm(m *Message) *WasmMessage {
	wm := &WasmMessage{
		Kind: m.Kind,
	}
	if len(m.Data) > 0 {
		wm.Data = base64.StdEncoding.EncodeToString(m.Data)
	}
	if len(m.Meta) > 0 {
		wm.Meta = mapToMeta(m.Meta)
	}
	return wm
}

// wasmToMessage converts a wire format WasmMessage to a Message.
func wasmToMessage(wm *WasmMessage) *Message {
	m := &Message{
		Kind: wm.Kind,
	}
	if wm.Data != "" {
		data, err := base64.StdEncoding.DecodeString(wm.Data)
		if err == nil {
			m.Data = data
		}
	}
	if len(wm.Meta) > 0 {
		m.Meta = metaToMap(wm.Meta)
	}
	return m
}

// resultToWasm converts a Result to its wire format representation.
func resultToWasm(r *Result) *WasmResult {
	wr := &WasmResult{
		Action: r.Action.String(),
	}
	if r.Response != nil {
		wr.Response = responseToWasm(r.Response)
	}
	if r.Err != nil {
		wr.Error = errorToWasm(r.Err)
	}
	return wr
}

// wasmToResult converts a wire format WasmResult to a Result.
func wasmToResult(wr *WasmResult) *Result {
	r := &Result{
		Action: ParseAction(wr.Action),
	}
	if wr.Response != nil {
		r.Response = wasmToResponse(wr.Response)
	}
	if wr.Error != nil {
		r.Err = wasmToError(wr.Error)
	}
	return r
}

// responseToWasm converts a Response to its wire format representation.
func responseToWasm(r *Response) *WasmResponse {
	wr := &WasmResponse{}
	if len(r.Data) > 0 {
		wr.Data = base64.StdEncoding.EncodeToString(r.Data)
	}
	if len(r.Meta) > 0 {
		wr.Meta = mapToMeta(r.Meta)
	}
	return wr
}

// wasmToResponse converts a wire format WasmResponse to a Response.
func wasmToResponse(wr *WasmResponse) *Response {
	r := &Response{}
	if wr.Data != "" {
		data, err := base64.StdEncoding.DecodeString(wr.Data)
		if err == nil {
			r.Data = data
		}
	}
	if len(wr.Meta) > 0 {
		r.Meta = metaToMap(wr.Meta)
	}
	return r
}

// errorToWasm converts a WaferError to its wire format representation.
func errorToWasm(e *WaferError) *WasmError {
	we := &WasmError{
		Code:    e.Code,
		Message: e.Message,
	}
	if len(e.Meta) > 0 {
		we.Meta = mapToMeta(e.Meta)
	}
	return we
}

// wasmToError converts a wire format WasmError to a WaferError.
func wasmToError(we *WasmError) *WaferError {
	e := &WaferError{
		Code:    we.Code,
		Message: we.Message,
	}
	if len(we.Meta) > 0 {
		e.Meta = metaToMap(we.Meta)
	}
	return e
}

// blockInfoToWasm converts a BlockInfo to its wire format representation.
func blockInfoToWasm(info *BlockInfo) *WasmBlockInfo {
	wi := &WasmBlockInfo{
		Name:         info.Name,
		Version:      info.Version,
		Interface:    info.Interface,
		Summary:      info.Summary,
		InstanceMode: info.InstanceMode.String(),
	}
	if len(info.AllowedModes) > 0 {
		wi.AllowedModes = make([]string, len(info.AllowedModes))
		for i, mode := range info.AllowedModes {
			wi.AllowedModes[i] = mode.String()
		}
	}
	return wi
}

// wasmToLifecycleEvent converts a wire format WasmLifecycleEvent to a
// LifecycleEvent.
func wasmToLifecycleEvent(we *WasmLifecycleEvent) *LifecycleEvent {
	e := &LifecycleEvent{
		Type: ParseLifecycleType(we.Type),
	}
	if we.Data != "" {
		data, err := base64.StdEncoding.DecodeString(we.Data)
		if err == nil {
			e.Data = data
		}
	}
	return e
}

// mapToMeta converts a map[string]string to the wire format [[key, value], ...].
func mapToMeta(m map[string]string) [][]string {
	meta := make([][]string, 0, len(m))
	for k, v := range m {
		meta = append(meta, []string{k, v})
	}
	return meta
}

// metaToMap converts the wire format [[key, value], ...] to a map[string]string.
func metaToMap(meta [][]string) map[string]string {
	m := make(map[string]string, len(meta))
	for _, pair := range meta {
		if len(pair) == 2 {
			m[pair[0]] = pair[1]
		}
	}
	return m
}
