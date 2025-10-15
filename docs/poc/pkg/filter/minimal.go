package filter

import (
	"encoding/json"
	"fmt"
)

// MinimalFilter implements Filter with a strategy that removes metadata,
// URLs, timestamps, and IDs from MCP responses.
type MinimalFilter struct {
	// RemoveFields specifies field names to remove at any level of nesting
	RemoveFields []string
}

// NewMinimalFilter creates a new MinimalFilter with default removal rules.
func NewMinimalFilter() *MinimalFilter {
	return &MinimalFilter{
		RemoveFields: []string{
			// MCP protocol overhead
			"_meta", "jsonrpc", "id",
			// URLs (GitHub and Notion)
			"url", "html_url", "api_url", "diff_url", "patch_url",
			"issue_url", "comments_url", "events_url", "labels_url",
			"repository_url", "review_comments_url", "review_comment_url",
			"commits_url", "statuses_url", "avatar_url", "gravatar_id",
			"forks_url", "keys_url", "collaborators_url", "teams_url",
			"hooks_url", "issue_events_url", "assignees_url", "branches_url",
			"tags_url", "blobs_url", "git_tags_url", "git_refs_url",
			"trees_url", "languages_url", "stargazers_url", "contributors_url",
			"subscribers_url", "subscription_url", "git_commits_url",
			"comments_url", "issue_comment_url", "contents_url", "compare_url",
			"merges_url", "archive_url", "downloads_url", "issues_url",
			"pulls_url", "milestones_url", "notifications_url", "releases_url",
			"deployments_url", "reactions", "timeline_url", "public_url",
			// IDs
			"node_id",
			// Timestamps
			"created_at", "updated_at", "closed_at", "merged_at", "pushed_at",
			"created_time", "last_edited_time",
			// Notion-specific metadata
			"object", "parent",
			// Git URLs
			"git_url", "ssh_url", "clone_url", "svn_url", "homepage",
			// Links
			"_links",
		},
	}
}

// Apply filters the response by recursively removing specified fields.
// Original response is not modified (creates a deep copy).
// Handles MCP-specific structure where content[0].text contains embedded JSON.
func (m *MinimalFilter) Apply(response map[string]interface{}) (map[string]interface{}, error) {
	// First, handle MCP-specific structure: content[0].text may contain embedded JSON
	filtered, err := m.filterMCPResponse(response)
	if err != nil {
		return nil, fmt.Errorf("filtering failed: %w", err)
	}

	typedFiltered, ok := filtered.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("filtering produced non-map result")
	}

	return typedFiltered, nil
}

// filterMCPResponse handles the MCP protocol structure where actual data
// is embedded as a JSON string in content[0].text
func (m *MinimalFilter) filterMCPResponse(response map[string]interface{}) (interface{}, error) {
	// Check if this is an MCP response with content array
	content, hasContent := response["content"]
	if !hasContent {
		// No content field, just filter normally
		return m.filterMap(response)
	}

	contentArray, isArray := content.([]interface{})
	if !isArray || len(contentArray) == 0 {
		// Content is not an array or is empty, filter normally
		return m.filterMap(response)
	}

	// Get first content item
	firstItem, isMap := contentArray[0].(map[string]interface{})
	if !isMap {
		// First item is not a map, filter normally
		return m.filterMap(response)
	}

	// Check if it has a text field with embedded JSON
	textField, hasText := firstItem["text"]
	if !hasText {
		// No text field, filter normally
		return m.filterMap(response)
	}

	textStr, isString := textField.(string)
	if !isString {
		// Text is not a string, filter normally
		return m.filterMap(response)
	}

	// Try to parse the embedded JSON
	var embedded interface{}
	if err := json.Unmarshal([]byte(textStr), &embedded); err != nil {
		// Not valid JSON, filter normally
		return m.filterMap(response)
	}

	// Filter the embedded data
	var filteredEmbedded interface{}
	var filterErr error

	switch embeddedData := embedded.(type) {
	case map[string]interface{}:
		filteredEmbedded, filterErr = m.filterMap(embeddedData)
	case []interface{}:
		filteredEmbedded, filterErr = m.filterArray(embeddedData)
	default:
		filteredEmbedded = embedded // Primitive, no filtering needed
	}

	if filterErr != nil {
		return nil, filterErr
	}

	// Re-serialize the filtered embedded data
	filteredJSON, err := json.Marshal(filteredEmbedded)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filtered embedded data: %w", err)
	}

	// Create filtered response with updated text field
	filteredResponse := make(map[string]interface{})
	for k, v := range response {
		if k == "content" {
			// Update content array with filtered text
			filteredContentArray := make([]interface{}, len(contentArray))
			for i, item := range contentArray {
				if i == 0 {
					// Update first item with filtered text
					filteredItem := make(map[string]interface{})
					for itemK, itemV := range firstItem {
						if itemK == "text" {
							filteredItem[itemK] = string(filteredJSON)
						} else {
							filteredItem[itemK] = itemV
						}
					}
					filteredContentArray[i] = filteredItem
				} else {
					filteredContentArray[i] = item
				}
			}
			filteredResponse[k] = filteredContentArray
		} else {
			filteredResponse[k] = v
		}
	}

	return filteredResponse, nil
}

// filterMap recursively filters a map, removing specified fields.
func (m *MinimalFilter) filterMap(data map[string]interface{}) (interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range data {
		// Skip fields in removal list
		if m.shouldRemove(key) {
			continue
		}

		// Recursively filter nested structures
		filtered, err := m.filterValue(value)
		if err != nil {
			return nil, err
		}

		result[key] = filtered
	}

	return result, nil
}

// filterValue recursively filters a value based on its type.
func (m *MinimalFilter) filterValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case map[string]interface{}:
		return m.filterMap(v)
	case []interface{}:
		return m.filterArray(v)
	default:
		// Primitives (string, int, bool, etc.) pass through unchanged
		return value, nil
	}
}

// filterArray recursively filters an array.
func (m *MinimalFilter) filterArray(data []interface{}) (interface{}, error) {
	result := make([]interface{}, 0, len(data))

	for _, item := range data {
		filtered, err := m.filterValue(item)
		if err != nil {
			return nil, err
		}
		result = append(result, filtered)
	}

	return result, nil
}

// shouldRemove checks if a field name should be removed.
func (m *MinimalFilter) shouldRemove(field string) bool {
	for _, remove := range m.RemoveFields {
		if field == remove {
			return true
		}
	}
	return false
}

// Name returns "minimal".
func (m *MinimalFilter) Name() string {
	return "minimal"
}
