package filter

import (
	"context"
)

// Filter defines the interface for filtering response data
type Filter interface {
	// Filter applies filtering to the given data based on the specified fields
	// Returns the filtered data or an error
	Filter(ctx context.Context, data interface{}, fields []string) (interface{}, error)
}
