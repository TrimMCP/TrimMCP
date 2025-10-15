package filter

// Filter transforms MCP responses by removing unnecessary data.
type Filter interface {
	// Apply filters the given response, returning a new filtered response.
	// Original response is not modified (immutable).
	// Returns error if filtering fails or produces invalid output.
	Apply(response map[string]interface{}) (map[string]interface{}, error)

	// Name returns the human-readable name of this filtering strategy.
	// Example: "minimal"
	Name() string
}
