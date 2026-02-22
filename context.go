package waffle

import (
	"encoding/json"
	"unsafe"
)

// Host function imports provided by the WAFFLE runtime.
// These are the raw WASM import signatures using the ptr+len memory protocol.

//go:wasmimport waffle send
func hostSend(ptr uint32, len uint32) uint64

//go:wasmimport waffle capabilities
func hostCapabilities() uint64

//go:wasmimport waffle is_cancelled
func hostIsCancelled() uint32

// Context wraps the WASM host imports and provides a high-level interface for
// blocks to interact with the WAFFLE runtime. It is the guest-side counterpart
// of the runtime's Context interface.
type Context struct{}

// Send sends a message to the WAFFLE runtime and returns the result. This is
// the fundamental mechanism through which blocks access runtime capabilities
// such as logging, configuration, database access, and network requests.
//
// The message is serialized to JSON in the wire format, written to linear
// memory, and passed to the host via the waffle.send import. The host returns
// a packed i64 (ptr in high 32 bits, len in low 32 bits) pointing to the
// JSON-encoded result.
func (c *Context) Send(msg *Message) *Result {
	wmsg := messageToWasm(msg)
	data, err := json.Marshal(wmsg)
	if err != nil {
		return &Result{
			Action: ActionError,
			Err: &WaffleError{
				Code:    "internal",
				Message: "failed to marshal message: " + err.Error(),
			},
		}
	}

	ptr := uint32(uintptr(unsafe.Pointer(&data[0])))
	length := uint32(len(data))

	packed := hostSend(ptr, length)
	resultPtr := uint32(packed >> 32)
	resultLen := uint32(packed & 0xFFFFFFFF)

	if resultLen == 0 {
		return &Result{Action: Continue}
	}

	resultData := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(resultPtr))), resultLen)
	var wresult WasmResult
	if err := json.Unmarshal(resultData, &wresult); err != nil {
		return &Result{
			Action: ActionError,
			Err: &WaffleError{
				Code:    "internal",
				Message: "failed to unmarshal result: " + err.Error(),
			},
		}
	}

	return wasmToResult(&wresult)
}

// Capabilities returns the list of runtime capabilities available to this
// block. Each capability describes a message kind the runtime understands,
// along with human-readable documentation and JSON Schema definitions for
// input and output.
func (c *Context) Capabilities() []CapabilityInfo {
	packed := hostCapabilities()
	ptr := uint32(packed >> 32)
	length := uint32(packed & 0xFFFFFFFF)

	if length == 0 {
		return nil
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), length)
	var caps []CapabilityInfo
	if err := json.Unmarshal(data, &caps); err != nil {
		return nil
	}

	return caps
}

// IsCancelled checks whether the runtime context has been cancelled. Blocks
// should check this during long-running operations and return early when the
// context is cancelled (e.g., due to a timeout).
func (c *Context) IsCancelled() bool {
	return hostIsCancelled() != 0
}

// CapabilityInfo describes a single runtime capability available through
// Context.Send.
type CapabilityInfo struct {
	// Kind is the message kind that activates this capability, e.g. "log"
	// or "config.get".
	Kind string `json:"kind"`

	// Summary is a human-readable description of what this capability does.
	Summary string `json:"summary"`

	// Input is a JSON Schema describing the expected input format.
	Input json.RawMessage `json:"input,omitempty"`

	// Output is a JSON Schema describing the expected output format.
	Output json.RawMessage `json:"output,omitempty"`
}
