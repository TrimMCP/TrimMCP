package benchmark

// Validator checks filtered responses for data integrity.
type Validator interface {
	// Validate checks if the filtered response meets integrity requirements.
	// Returns error describing any validation failures.
	// Returns nil if validation passes.
	Validate(operation string, filtered map[string]interface{}) error

	// RegisterSchema adds a validation schema for a specific operation.
	// Must be called before Validate for that operation.
	RegisterSchema(operation string, schema Schema)
}

// Schema defines required fields and types for a response.
type Schema interface {
	// CheckRequiredFields verifies all required fields are present.
	// Returns error listing missing fields, or nil if all present.
	CheckRequiredFields(data map[string]interface{}) error

	// CheckTypes verifies field values have correct types.
	// Returns error describing type mismatches, or nil if all correct.
	CheckTypes(data map[string]interface{}) error
}
