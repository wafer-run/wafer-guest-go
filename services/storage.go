package services

import (
	"encoding/json"

	wafer "github.com/anthropics/wafer-sdk-go"
)

// StorageClient provides typed access to the WAFER storage capability for
// binary object storage operations. All operations are sent as messages
// through the context.
type StorageClient struct {
	ctx *wafer.Context
}

// NewStorageClient creates a new StorageClient bound to the given context.
func NewStorageClient(ctx *wafer.Context) *StorageClient {
	return &StorageClient{ctx: ctx}
}

// Put stores content in a bucket under the given key. The content is sent as
// the message data.
//
// Message kind: "svc.storage.put"
// Meta: [["bucket", bucket], ["key", key]]
// Data: content bytes
func (s *StorageClient) Put(bucket, key string, content []byte) (*wafer.Result, error) {
	msg := &wafer.Message{
		Kind: "svc.storage.put",
		Data: content,
		Meta: map[string]string{
			"bucket": bucket,
			"key":    key,
		},
	}
	result := s.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return result, result.Err
	}
	return result, nil
}

// Get retrieves content from a bucket by key. The content is returned in the
// response data.
//
// Message kind: "svc.storage.get"
// Meta: [["bucket", bucket], ["key", key]]
func (s *StorageClient) Get(bucket, key string) ([]byte, error) {
	msg := &wafer.Message{
		Kind: "svc.storage.get",
		Meta: map[string]string{
			"bucket": bucket,
			"key":    key,
		},
	}
	result := s.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return nil, result.Err
	}
	if result.Response == nil {
		return nil, nil
	}
	return result.Response.Data, nil
}

// Delete removes content from a bucket by key.
//
// Message kind: "svc.storage.delete"
// Meta: [["bucket", bucket], ["key", key]]
func (s *StorageClient) Delete(bucket, key string) (*wafer.Result, error) {
	msg := &wafer.Message{
		Kind: "svc.storage.delete",
		Meta: map[string]string{
			"bucket": bucket,
			"key":    key,
		},
	}
	result := s.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return result, result.Err
	}
	return result, nil
}

// StorageEntry represents a single item returned by List.
type StorageEntry struct {
	Key  string `json:"key"`
	Size int64  `json:"size,omitempty"`
}

// List returns all keys in a bucket. The response data contains a JSON array
// of StorageEntry objects.
//
// Message kind: "svc.storage.list"
// Meta: [["bucket", bucket]]
func (s *StorageClient) List(bucket string) ([]StorageEntry, error) {
	msg := &wafer.Message{
		Kind: "svc.storage.list",
		Meta: map[string]string{
			"bucket": bucket,
		},
	}
	result := s.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return nil, result.Err
	}
	if result.Response == nil || len(result.Response.Data) == 0 {
		return nil, nil
	}
	var entries []StorageEntry
	if err := json.Unmarshal(result.Response.Data, &entries); err != nil {
		return nil, &wafer.WaferError{
			Code:    "internal",
			Message: "failed to unmarshal storage list: " + err.Error(),
		}
	}
	return entries, nil
}
