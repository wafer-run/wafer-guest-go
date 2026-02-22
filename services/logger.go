package services

import (
	waffle "github.com/anthropics/waffle-guest-go"
)

// LoggerClient provides typed access to the WAFFLE logging capability. Log
// messages are sent to the runtime which handles output formatting and routing.
type LoggerClient struct {
	ctx *waffle.Context
}

// NewLoggerClient creates a new LoggerClient bound to the given context.
func NewLoggerClient(ctx *waffle.Context) *LoggerClient {
	return &LoggerClient{ctx: ctx}
}

// log sends a log message at the given level.
func (l *LoggerClient) log(level, message string) {
	msg := &waffle.Message{
		Kind: "svc.logger." + level,
		Data: []byte(message),
	}
	l.ctx.Send(msg)
}

// Debug sends a debug-level log message.
//
// Message kind: "svc.logger.debug"
// Data: message string
func (l *LoggerClient) Debug(message string) {
	l.log("debug", message)
}

// Info sends an info-level log message.
//
// Message kind: "svc.logger.info"
// Data: message string
func (l *LoggerClient) Info(message string) {
	l.log("info", message)
}

// Warn sends a warning-level log message.
//
// Message kind: "svc.logger.warn"
// Data: message string
func (l *LoggerClient) Warn(message string) {
	l.log("warn", message)
}

// Error sends an error-level log message.
//
// Message kind: "svc.logger.error"
// Data: message string
func (l *LoggerClient) Error(message string) {
	l.log("error", message)
}
