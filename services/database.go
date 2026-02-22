// Package services provides typed client wrappers for WAFFLE runtime
// capabilities. Each client sends structured messages through the Context.Send
// method and returns parsed results.
package services

import (
	"encoding/json"

	waffle "github.com/anthropics/waffle-guest-go"
)

// DatabaseClient provides typed access to the WAFFLE database capability.
// All operations are sent as messages through the context and handled by the
// runtime's database service.
type DatabaseClient struct {
	ctx *waffle.Context
}

// NewDatabaseClient creates a new DatabaseClient bound to the given context.
func NewDatabaseClient(ctx *waffle.Context) *DatabaseClient {
	return &DatabaseClient{ctx: ctx}
}

// Get retrieves a single record by collection and ID. The result is
// unmarshaled from the response data.
//
// Message kind: "svc.database.get"
// Meta: [["collection", collection], ["id", id]]
func (d *DatabaseClient) Get(collection, id string) (*waffle.Result, error) {
	msg := &waffle.Message{
		Kind: "svc.database.get",
		Meta: map[string]string{
			"collection": collection,
			"id":         id,
		},
	}
	result := d.ctx.Send(msg)
	if result.Action == waffle.ActionError && result.Err != nil {
		return result, result.Err
	}
	return result, nil
}

// GetInto retrieves a single record and unmarshals the response data into
// the provided value.
func (d *DatabaseClient) GetInto(collection, id string, v any) error {
	result, err := d.Get(collection, id)
	if err != nil {
		return err
	}
	if result.Response == nil || len(result.Response.Data) == 0 {
		return &waffle.WaffleError{
			Code:    "not_found",
			Message: "record not found in " + collection + ": " + id,
		}
	}
	return json.Unmarshal(result.Response.Data, v)
}

// List retrieves all records from a collection. The result contains the list
// of records in the response data.
//
// Message kind: "svc.database.list"
// Meta: [["collection", collection]]
func (d *DatabaseClient) List(collection string) (*waffle.Result, error) {
	msg := &waffle.Message{
		Kind: "svc.database.list",
		Meta: map[string]string{
			"collection": collection,
		},
	}
	result := d.ctx.Send(msg)
	if result.Action == waffle.ActionError && result.Err != nil {
		return result, result.Err
	}
	return result, nil
}

// ListInto retrieves all records from a collection and unmarshals the
// response data into the provided value (typically a slice pointer).
func (d *DatabaseClient) ListInto(collection string, v any) error {
	result, err := d.List(collection)
	if err != nil {
		return err
	}
	if result.Response == nil || len(result.Response.Data) == 0 {
		return nil
	}
	return json.Unmarshal(result.Response.Data, v)
}

// Create inserts a new record into a collection. The record is JSON-encoded
// and sent as the message data.
//
// Message kind: "svc.database.create"
// Meta: [["collection", collection]]
// Data: JSON-encoded record
func (d *DatabaseClient) Create(collection string, record any) (*waffle.Result, error) {
	data, err := json.Marshal(record)
	if err != nil {
		return nil, &waffle.WaffleError{
			Code:    "internal",
			Message: "failed to marshal record: " + err.Error(),
		}
	}
	msg := &waffle.Message{
		Kind: "svc.database.create",
		Data: data,
		Meta: map[string]string{
			"collection": collection,
		},
	}
	result := d.ctx.Send(msg)
	if result.Action == waffle.ActionError && result.Err != nil {
		return result, result.Err
	}
	return result, nil
}

// Update modifies an existing record in a collection. The record is
// JSON-encoded and sent as the message data.
//
// Message kind: "svc.database.update"
// Meta: [["collection", collection], ["id", id]]
// Data: JSON-encoded record
func (d *DatabaseClient) Update(collection, id string, record any) (*waffle.Result, error) {
	data, err := json.Marshal(record)
	if err != nil {
		return nil, &waffle.WaffleError{
			Code:    "internal",
			Message: "failed to marshal record: " + err.Error(),
		}
	}
	msg := &waffle.Message{
		Kind: "svc.database.update",
		Data: data,
		Meta: map[string]string{
			"collection": collection,
			"id":         id,
		},
	}
	result := d.ctx.Send(msg)
	if result.Action == waffle.ActionError && result.Err != nil {
		return result, result.Err
	}
	return result, nil
}

// Delete removes a record from a collection by ID.
//
// Message kind: "svc.database.delete"
// Meta: [["collection", collection], ["id", id]]
func (d *DatabaseClient) Delete(collection, id string) (*waffle.Result, error) {
	msg := &waffle.Message{
		Kind: "svc.database.delete",
		Meta: map[string]string{
			"collection": collection,
			"id":         id,
		},
	}
	result := d.ctx.Send(msg)
	if result.Action == waffle.ActionError && result.Err != nil {
		return result, result.Err
	}
	return result, nil
}
