package services

import (
	"encoding/json"

	wafer "github.com/anthropics/wafer-sdk-go"
)

// FilterOp represents a filter comparison operator.
type FilterOp string

const (
	OpEqual        FilterOp = "eq"
	OpNotEqual     FilterOp = "neq"
	OpGreater      FilterOp = "gt"
	OpGreaterEqual FilterOp = "gte"
	OpLess         FilterOp = "lt"
	OpLessEqual    FilterOp = "lte"
	OpLike         FilterOp = "like"
	OpIn           FilterOp = "in"
)

// Filter represents a single filter condition for database queries.
type Filter struct {
	Field    string      `json:"field"`
	Operator FilterOp    `json:"operator"`
	Value    interface{} `json:"value"`
}

// SortField defines a sort directive for database queries.
type SortField struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc"`
}

// ListOptions configures a List query with filtering, sorting, and pagination.
type ListOptions struct {
	Filters []Filter    `json:"filters,omitempty"`
	Sort    []SortField `json:"sort,omitempty"`
	Limit   int64       `json:"limit,omitempty"`
	Offset  int64       `json:"offset,omitempty"`
}

// Record matches the host's Record type.
type Record struct {
	ID   string                 `json:"id"`
	Data map[string]interface{} `json:"data"`
}

// RecordList matches the host's RecordList type.
type RecordList struct {
	Records    []Record `json:"records"`
	TotalCount int64    `json:"total_count"`
	Page       int64    `json:"page"`
	PageSize   int64    `json:"page_size"`
}

// ListWithOptions retrieves records from a collection with filtering, sorting,
// and pagination options. The ListOptions are JSON-encoded in the message data.
func (d *DatabaseClient) ListWithOptions(collection string, opts *ListOptions) (*RecordList, error) {
	var data []byte
	if opts != nil {
		var err error
		data, err = json.Marshal(opts)
		if err != nil {
			return nil, &wafer.WaferError{
				Code:    "internal",
				Message: "failed to marshal list options: " + err.Error(),
			}
		}
	}
	msg := &wafer.Message{
		Kind: "svc.database.list",
		Data: data,
		Meta: map[string]string{
			"collection": collection,
		},
	}
	result := d.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return nil, result.Err
	}
	if result.Response == nil || len(result.Response.Data) == 0 {
		return &RecordList{}, nil
	}
	var rl RecordList
	if err := json.Unmarshal(result.Response.Data, &rl); err != nil {
		return nil, &wafer.WaferError{
			Code:    "internal",
			Message: "failed to unmarshal record list: " + err.Error(),
		}
	}
	return &rl, nil
}

// ListRecords retrieves all records from a collection, returning typed Records.
func (d *DatabaseClient) ListRecords(collection string) ([]Record, error) {
	rl, err := d.ListWithOptions(collection, nil)
	if err != nil {
		return nil, err
	}
	return rl.Records, nil
}

// GetRecord retrieves a single record by collection and ID, returning a typed Record.
func (d *DatabaseClient) GetRecord(collection, id string) (*Record, error) {
	result, err := d.Get(collection, id)
	if err != nil {
		return nil, err
	}
	if result.Response == nil || len(result.Response.Data) == 0 {
		return nil, &wafer.WaferError{
			Code:    "not_found",
			Message: "record not found in " + collection + ": " + id,
		}
	}
	var r Record
	if err := json.Unmarshal(result.Response.Data, &r); err != nil {
		return nil, &wafer.WaferError{
			Code:    "internal",
			Message: "failed to unmarshal record: " + err.Error(),
		}
	}
	return &r, nil
}

// GetByField retrieves a single record where field equals value.
func (d *DatabaseClient) GetByField(collection, field, value string) (*Record, error) {
	rl, err := d.ListWithOptions(collection, &ListOptions{
		Filters: []Filter{{
			Field:    field,
			Operator: OpEqual,
			Value:    value,
		}},
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}
	if len(rl.Records) == 0 {
		return nil, &wafer.WaferError{
			Code:    "not_found",
			Message: "record not found in " + collection + " where " + field + " = " + value,
		}
	}
	return &rl.Records[0], nil
}

// Count returns the number of records in a collection matching the given filters.
func (d *DatabaseClient) Count(collection string, filters ...Filter) (int64, error) {
	var data []byte
	if len(filters) > 0 {
		var err error
		data, err = json.Marshal(filters)
		if err != nil {
			return 0, &wafer.WaferError{
				Code:    "internal",
				Message: "failed to marshal filters: " + err.Error(),
			}
		}
	}
	msg := &wafer.Message{
		Kind: "svc.database.count",
		Data: data,
		Meta: map[string]string{
			"collection": collection,
		},
	}
	result := d.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return 0, result.Err
	}
	if result.Response == nil || len(result.Response.Data) == 0 {
		return 0, nil
	}
	var count int64
	if err := json.Unmarshal(result.Response.Data, &count); err != nil {
		return 0, &wafer.WaferError{
			Code:    "internal",
			Message: "failed to unmarshal count: " + err.Error(),
		}
	}
	return count, nil
}

// QueryRaw executes a raw SELECT query and returns records.
func (d *DatabaseClient) QueryRaw(query string, args ...interface{}) ([]Record, error) {
	payload := map[string]interface{}{
		"query": query,
		"args":  args,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, &wafer.WaferError{
			Code:    "internal",
			Message: "failed to marshal query: " + err.Error(),
		}
	}
	msg := &wafer.Message{
		Kind: "svc.database.query_raw",
		Data: data,
	}
	result := d.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return nil, result.Err
	}
	if result.Response == nil || len(result.Response.Data) == 0 {
		return nil, nil
	}
	var records []Record
	if err := json.Unmarshal(result.Response.Data, &records); err != nil {
		return nil, &wafer.WaferError{
			Code:    "internal",
			Message: "failed to unmarshal records: " + err.Error(),
		}
	}
	return records, nil
}

// ExecRaw executes a raw non-SELECT statement and returns the number of affected rows.
func (d *DatabaseClient) ExecRaw(query string, args ...interface{}) (int64, error) {
	payload := map[string]interface{}{
		"query": query,
		"args":  args,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, &wafer.WaferError{
			Code:    "internal",
			Message: "failed to marshal query: " + err.Error(),
		}
	}
	msg := &wafer.Message{
		Kind: "svc.database.exec_raw",
		Data: data,
	}
	result := d.ctx.Send(msg)
	if result.Action == wafer.ActionError && result.Err != nil {
		return 0, result.Err
	}
	if result.Response == nil || len(result.Response.Data) == 0 {
		return 0, nil
	}
	var affected int64
	if err := json.Unmarshal(result.Response.Data, &affected); err != nil {
		return 0, &wafer.WaferError{
			Code:    "internal",
			Message: "failed to unmarshal affected rows: " + err.Error(),
		}
	}
	return affected, nil
}
