package services

import (
	waffle "github.com/anthropics/waffle-guest-go"
)

// ConfigClient provides typed access to the WAFFLE configuration capability.
// Configuration values are string key-value pairs managed by the runtime.
type ConfigClient struct {
	ctx *waffle.Context
}

// NewConfigClient creates a new ConfigClient bound to the given context.
func NewConfigClient(ctx *waffle.Context) *ConfigClient {
	return &ConfigClient{ctx: ctx}
}

// Get retrieves a configuration value by key. Returns the value and true if
// the key exists, or an empty string and false if it does not.
//
// Message kind: "svc.config.get"
// Meta: [["key", key]]
func (c *ConfigClient) Get(key string) (string, bool) {
	msg := &waffle.Message{
		Kind: "svc.config.get",
		Meta: map[string]string{
			"key": key,
		},
	}
	result := c.ctx.Send(msg)
	if result.Action == waffle.ActionError || result.Response == nil {
		return "", false
	}
	return string(result.Response.Data), true
}

// GetDefault retrieves a configuration value by key, returning the provided
// default value if the key does not exist.
//
// Message kind: "svc.config.get"
// Meta: [["key", key]]
func (c *ConfigClient) GetDefault(key, defaultValue string) string {
	value, ok := c.Get(key)
	if !ok {
		return defaultValue
	}
	return value
}

// Set sets a configuration value for the given key.
//
// Message kind: "svc.config.set"
// Meta: [["key", key]]
// Data: value string
func (c *ConfigClient) Set(key, value string) error {
	msg := &waffle.Message{
		Kind: "svc.config.set",
		Data: []byte(value),
		Meta: map[string]string{
			"key": key,
		},
	}
	result := c.ctx.Send(msg)
	if result.Action == waffle.ActionError && result.Err != nil {
		return result.Err
	}
	return nil
}
