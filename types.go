package wafer

// Message represents a message flowing through a WAFER chain. A message
// contains a kind identifier, binary payload data, and string metadata.
type Message struct {
	// Kind identifies the type of message, e.g. "user.create" or "order.process".
	// Used for conditional routing via match patterns in chain configuration.
	Kind string

	// Data holds the message payload, typically JSON-encoded.
	Data []byte

	// Meta holds string key-value metadata such as headers, trace IDs, and
	// authentication context.
	Meta map[string]string
}

// GetMeta returns the metadata value for the given key, or an empty string
// if the key is not present.
func (m *Message) GetMeta(key string) string {
	if m.Meta == nil {
		return ""
	}
	return m.Meta[key]
}

// SetMeta sets a metadata key-value pair on the message. It initializes the
// Meta map if it is nil.
func (m *Message) SetMeta(key, value string) {
	if m.Meta == nil {
		m.Meta = make(map[string]string)
	}
	m.Meta[key] = value
}

// Action indicates what the runtime should do after a block processes a message.
type Action int

const (
	// Continue tells the runtime to pass the message to the next block in the chain.
	Continue Action = iota

	// Respond tells the runtime to short-circuit the chain and return the
	// response to the caller immediately.
	Respond

	// Drop tells the runtime to end the chain silently with no response.
	Drop

	// ActionError tells the runtime to short-circuit the chain and return an error
	// to the caller immediately.
	ActionError
)

// String returns the wire-format string representation of the Action.
func (a Action) String() string {
	switch a {
	case Continue:
		return "continue"
	case Respond:
		return "respond"
	case Drop:
		return "drop"
	case ActionError:
		return "error"
	default:
		return "continue"
	}
}

// ParseAction converts a wire-format action string to an Action value.
// Unrecognized strings default to Continue.
func ParseAction(s string) Action {
	switch s {
	case "continue":
		return Continue
	case "respond":
		return Respond
	case "drop":
		return Drop
	case "error":
		return ActionError
	default:
		return Continue
	}
}

// Response holds the data returned when a block short-circuits the chain
// with a Respond action.
type Response struct {
	// Data is the response payload, typically JSON-encoded.
	Data []byte

	// Meta holds string key-value metadata for the response.
	Meta map[string]string
}

// WaferError represents an error returned by a block. It contains a machine-
// readable code, a human-readable message, and optional metadata.
type WaferError struct {
	// Code is a machine-readable error code, e.g. "invalid_argument" or
	// "not_found". See the WAFER specification for recommended codes.
	Code string

	// Message is a human-readable description of the error.
	Message string

	// Meta holds optional string key-value metadata about the error.
	Meta map[string]string
}

// Error implements the error interface so WaferError can be used as a
// standard Go error.
func (e *WaferError) Error() string {
	return e.Code + ": " + e.Message
}

// Result is the outcome of processing a message. It tells the runtime what
// action to take next and optionally carries a response or error payload.
type Result struct {
	// Action indicates what the runtime should do next.
	Action Action

	// Response holds the response data when Action is Respond.
	Response *Response

	// Err holds the error data when Action is ActionError.
	Err *WaferError
}

// InstanceMode controls how many instances of a block are created and when.
type InstanceMode int

const (
	// PerNode creates one instance per chain node. This is the default mode.
	PerNode InstanceMode = iota

	// Singleton creates one instance shared across all chains.
	Singleton

	// PerChain creates one instance per chain, shared across nodes within that chain.
	PerChain

	// PerExecution creates a new instance for every message.
	PerExecution
)

// String returns the wire-format string representation of the InstanceMode.
func (m InstanceMode) String() string {
	switch m {
	case PerNode:
		return "per-node"
	case Singleton:
		return "singleton"
	case PerChain:
		return "per-chain"
	case PerExecution:
		return "per-execution"
	default:
		return "per-node"
	}
}

// ParseInstanceMode converts a wire-format instance mode string to an
// InstanceMode value. Unrecognized strings default to PerNode.
func ParseInstanceMode(s string) InstanceMode {
	switch s {
	case "per-node":
		return PerNode
	case "singleton":
		return Singleton
	case "per-chain":
		return PerChain
	case "per-execution":
		return PerExecution
	default:
		return PerNode
	}
}

// BlockInfo declares the identity, interface, and instance lifecycle of a block.
type BlockInfo struct {
	// Name is the block identifier, e.g. "@example/myblock".
	Name string

	// Version is the semantic version of the block, e.g. "2.1.0".
	Version string

	// Interface declares what contract the block implements, e.g. "database@v1".
	Interface string

	// Summary is a brief human-readable description of this block implementation.
	Summary string

	// InstanceMode is the default instance lifecycle for this block.
	InstanceMode InstanceMode

	// AllowedModes lists the instance modes this block supports. If empty, all
	// modes are permitted.
	AllowedModes []InstanceMode
}

// LifecycleEvent represents a lifecycle event delivered to a block by the runtime.
type LifecycleEvent struct {
	// Type identifies the lifecycle event kind.
	Type LifecycleType

	// Data holds event-specific payload. For Init events this is the block's
	// configuration JSON.
	Data []byte
}

// LifecycleType identifies the kind of lifecycle event.
type LifecycleType int

const (
	// Init indicates the block is being initialized. Data contains the
	// block's configuration JSON.
	Init LifecycleType = iota

	// Start indicates the chain is starting and is about to begin
	// processing messages.
	Start

	// Stop indicates the chain is shutting down.
	Stop
)

// String returns the wire-format string representation of the LifecycleType.
func (t LifecycleType) String() string {
	switch t {
	case Init:
		return "init"
	case Start:
		return "start"
	case Stop:
		return "stop"
	default:
		return "init"
	}
}

// ParseLifecycleType converts a wire-format lifecycle type string to a
// LifecycleType value. Unrecognized strings default to Init.
func ParseLifecycleType(s string) LifecycleType {
	switch s {
	case "init":
		return Init
	case "start":
		return Start
	case "stop":
		return Stop
	default:
		return Init
	}
}
