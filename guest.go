// Package waffle provides the guest SDK for writing WAFFLE blocks that compile
// to WebAssembly (GOOS=wasip1 GOARCH=wasm). Blocks built with this SDK run
// inside the WAFFLE runtime and communicate with the host through a well-defined
// memory protocol using JSON serialization.
//
// Build blocks with:
//
//	GOOS=wasip1 GOARCH=wasm go build -o block.wasm .
//
// A minimal block implementation looks like:
//
//	package main
//
//	import waffle "github.com/anthropics/waffle-guest-go"
//
//	type MyBlock struct{}
//
//	func (b *MyBlock) Info() waffle.BlockInfo {
//	    return waffle.BlockInfo{
//	        Name:      "@example/myblock",
//	        Version:   "1.0.0",
//	        Interface: "processor@v1",
//	        Summary:   "A simple message processor.",
//	    }
//	}
//
//	func (b *MyBlock) Handle(ctx *waffle.Context, msg *waffle.Message) *waffle.Result {
//	    return waffle.ContinueResult()
//	}
//
//	func (b *MyBlock) Lifecycle(ctx *waffle.Context, event waffle.LifecycleEvent) error {
//	    return nil
//	}
//
//	func main() {
//	    waffle.Register(&MyBlock{})
//	}
package waffle

import (
	"encoding/json"
	"unsafe"
)

// Block is the interface that every WAFFLE guest block must implement. The
// runtime calls these methods through the exported WASM functions.
type Block interface {
	// Info returns the block's identity, interface declaration, and instance
	// lifecycle configuration.
	Info() BlockInfo

	// Handle processes a message and returns a result that tells the runtime
	// what to do next. The Context provides access to runtime capabilities
	// such as logging, configuration, database operations, and network requests.
	Handle(ctx *Context, msg *Message) *Result

	// Lifecycle handles lifecycle events from the runtime. Init is called with
	// the block's configuration when the chain loads, Start is called before
	// the first message, and Stop is called during shutdown. Return a non-nil
	// error to signal failure (e.g., during Init to prevent the chain from
	// starting).
	Lifecycle(ctx *Context, event LifecycleEvent) error
}

// registeredBlock holds the globally registered block instance. Only one block
// can be registered per WASM module since each module is a single block.
var registeredBlock Block

// pinnedData holds references to data returned to the host, preventing garbage
// collection until the next export call. Each exported function clears this
// slice at the start of its execution.
var pinnedData [][]byte

// Register stores a Block implementation as the global block for this WASM
// module. This must be called from main() before the runtime invokes any
// exported functions. Only one block may be registered per module; subsequent
// calls overwrite the previous registration.
//
// Example usage:
//
//	func main() {
//	    waffle.Register(&MyBlock{})
//	}
func Register(block Block) {
	registeredBlock = block
}

// Exported WASM functions. These are called by the WAFFLE runtime host.
// They use the ptr+len memory protocol: input data is written by the host
// into memory allocated by malloc, and output data is returned as a packed
// i64 with the pointer in the high 32 bits and the length in the low 32 bits.

// info is called by the host to retrieve the block's identity information.
// It returns a packed i64 pointing to the JSON-encoded WasmBlockInfo.
//
//go:wasmexport info
func info() uint64 {
	pinnedData = pinnedData[:0]

	if registeredBlock == nil {
		return 0
	}

	blockInfo := registeredBlock.Info()
	wasmInfo := blockInfoToWasm(&blockInfo)

	data, err := json.Marshal(wasmInfo)
	if err != nil {
		return 0
	}

	return packPtrLen(data)
}

// handle is called by the host to process a message. The host writes the
// JSON-encoded WasmMessage into guest memory at the given ptr+len location.
// The function returns a packed i64 pointing to the JSON-encoded WasmResult.
//
//go:wasmexport handle
func handle(ptr uint32, length uint32) uint64 {
	pinnedData = pinnedData[:0]

	if registeredBlock == nil {
		errResult := &WasmResult{
			Action: "error",
			Error: &WasmError{
				Code:    "internal",
				Message: "no block registered",
			},
		}
		data, _ := json.Marshal(errResult)
		return packPtrLen(data)
	}

	// Read the message data from memory.
	msgData := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), length)

	var wmsg WasmMessage
	if err := json.Unmarshal(msgData, &wmsg); err != nil {
		errResult := &WasmResult{
			Action: "error",
			Error: &WasmError{
				Code:    "internal",
				Message: "failed to unmarshal message: " + err.Error(),
			},
		}
		data, _ := json.Marshal(errResult)
		return packPtrLen(data)
	}

	msg := wasmToMessage(&wmsg)
	ctx := &Context{}
	result := registeredBlock.Handle(ctx, msg)

	wasmResult := resultToWasm(result)
	data, err := json.Marshal(wasmResult)
	if err != nil {
		errResult := &WasmResult{
			Action: "error",
			Error: &WasmError{
				Code:    "internal",
				Message: "failed to marshal result: " + err.Error(),
			},
		}
		data, _ = json.Marshal(errResult)
		return packPtrLen(data)
	}

	return packPtrLen(data)
}

// lifecycle is called by the host to deliver lifecycle events to the block.
// The host writes the JSON-encoded WasmLifecycleEvent into guest memory at
// the given ptr+len location. Returns a packed i64 pointing to a JSON-encoded
// result: {"ok": true} on success or {"ok": false, "error": "..."} on failure.
//
//go:wasmexport lifecycle
func lifecycle(ptr uint32, length uint32) uint64 {
	pinnedData = pinnedData[:0]

	if registeredBlock == nil {
		return packErrorResponse("no block registered")
	}

	eventData := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), length)

	var wevent WasmLifecycleEvent
	if err := json.Unmarshal(eventData, &wevent); err != nil {
		return packErrorResponse("failed to unmarshal lifecycle event: " + err.Error())
	}

	event := wasmToLifecycleEvent(&wevent)
	ctx := &Context{}

	if err := registeredBlock.Lifecycle(ctx, *event); err != nil {
		return packErrorResponse(err.Error())
	}

	return packOkResponse()
}

// malloc allocates memory from the Go heap for the host to write into. The
// host calls this before invoking handle or lifecycle to reserve space for the
// input data.
//
//go:wasmexport malloc
func malloc(size uint32) uint32 {
	buf := make([]byte, size)
	return uint32(uintptr(unsafe.Pointer(&buf[0])))
}

// packPtrLen packs a byte slice reference into a uint64 with the pointer in
// the high 32 bits and the length in the low 32 bits. The data is pinned to
// prevent garbage collection before the host reads it.
func packPtrLen(data []byte) uint64 {
	if len(data) == 0 {
		return 0
	}
	pinnedData = append(pinnedData, data)
	ptr := uint64(uintptr(unsafe.Pointer(&data[0])))
	length := uint64(len(data))
	return (ptr << 32) | length
}

// lifecycleResponse is the JSON structure returned by the lifecycle export.
type lifecycleResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// packOkResponse returns a packed ptr+len for a successful lifecycle response.
func packOkResponse() uint64 {
	data, _ := json.Marshal(lifecycleResponse{OK: true})
	return packPtrLen(data)
}

// packErrorResponse returns a packed ptr+len for a failed lifecycle response.
func packErrorResponse(msg string) uint64 {
	data, _ := json.Marshal(lifecycleResponse{OK: false, Error: msg})
	return packPtrLen(data)
}
