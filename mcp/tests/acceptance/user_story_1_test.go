package acceptance

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/slimmcp/slimmcp/tests/util"
	"github.com/slimmcp/slimmcp/pkg/types"
)

// TestUserStory1_NotionPageWith45Fields tests the scenario:
// Given LLM client connected to proxy
// When requests Notion page with 45 fields
// Then response has only 5 essential fields with 85%+ savings
func TestUserStory1_NotionPageWith45Fields(t *testing.T) {
	// Setup mock MCP server
	mockServer := util.NewMockMCPServer()

	// Load Notion fixtures
	fixture, err := util.LoadFixture("notion", "query_database")
	if err != nil {
		t.Fatalf("Failed to load fixture: %v", err)
	}

	// Get the response content from fixture
	responseContent, err := fixture.GetResponseContent()
	if err != nil {
		t.Fatalf("Failed to get response content: %v", err)
	}

	mockServer.SetRawResponse("query_database", string(responseContent))

	// Start mock server
	mockServerInstance := httptest.NewServer(mockServer)
	defer mockServerInstance.Close()

	// TODO: Setup proxy server pointing to mock server
	// proxyServer := setupProxyServer(mockServerInstance.URL)
	// defer proxyServer.Close()

	// Create request (no _meta.fieldSelection - automatic filtering)
	request := types.MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: &types.RequestParams{
			Name: "query_database",
			Arguments: map[string]interface{}{
				"database_id": "db123456-7890-abcd-ef01-234567890abc",
			},
		},
		ID: 1,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// TODO: Send request to proxy
	// resp, err := http.Post(proxyServer.URL+"/notion", "application/json", bytes.NewBuffer(requestBody))
	// if err != nil {
	//     t.Fatalf("Failed to send request: %v", err)
	// }
	// defer resp.Body.Close()

	// For now, just verify the test structure is correct
	t.Skip("Proxy server not yet implemented - will implement in T026-T030")

	// TODO: Parse response
	// var response types.MCPResponse
	// if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
	//     t.Fatalf("Failed to decode response: %v", err)
	// }

	// TODO: Verify response has only essential fields
	// expectedFields := []string{"id", "type", "url", "title", "object"}
	// ... verify fields ...

	// TODO: Verify 85%+ token savings
	// if response.Meta == nil || response.Meta.FieldSelection == nil {
	//     t.Fatal("Expected filtering metadata in response")
	// }
	//
	// if response.Meta.FieldSelection.Savings < 85.0 {
	//     t.Errorf("Expected at least 85%% savings, got %.2f%%", response.Meta.FieldSelection.Savings)
	// }

	_ = requestBody // Use the variable to avoid unused error
}

// TestUserStory1_FieldInferenceFromParameters tests the scenario:
// Given tool call mentioning specific fields in parameters
// When proxy analyzes request
// Then infers needed fields and filters accordingly
func TestUserStory1_FieldInferenceFromParameters(t *testing.T) {
	// Setup mock MCP server
	mockServer := util.NewMockMCPServer()

	// Create a response with many fields including "status"
	notionResponse := util.CreateNotionPageResponse()
	mockServer.SetResponse("get_page", notionResponse)

	mockServerInstance := httptest.NewServer(mockServer)
	defer mockServerInstance.Close()

	// TODO: Setup proxy with smart strategy
	// proxyServer := setupProxyServerWithStrategy(mockServerInstance.URL, "smart")
	// defer proxyServer.Close()

	// Create request that mentions "status" in arguments
	request := types.MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: &types.RequestParams{
			Name: "get_page",
			Arguments: map[string]interface{}{
				"page_id": "page-123",
				"filter":  "status=in_progress", // Mentions "status" field
			},
		},
		ID: 2,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// TODO: Send request and verify smart analyzer includes "status" field
	t.Skip("Smart analyzer not yet implemented - will implement in Phase 8")

	// TODO: Verify response includes status field
	// var response types.MCPResponse
	// ... parse response ...
	// ... verify "status" field is included ...

	_ = requestBody // Use the variable
}

// TestUserStory1_ConsistentFiltering tests the scenario:
// Given multiple consecutive requests for similar data
// When proxy processes each
// Then responses filtered consistently
func TestUserStory1_ConsistentFiltering(t *testing.T) {
	// Setup mock MCP server
	mockServer := util.NewMockMCPServer()

	// Load fixtures
	if err := util.SetupMockServerWithFixtures(mockServer, "notion"); err != nil {
		t.Fatalf("Failed to setup fixtures: %v", err)
	}

	mockServerInstance := httptest.NewServer(mockServer)
	defer mockServerInstance.Close()

	// TODO: Setup proxy server
	// proxyServer := setupProxyServer(mockServerInstance.URL)
	// defer proxyServer.Close()

	// Send multiple requests for the same tool
	numRequests := 5
	var responses []types.MCPResponse

	for i := 0; i < numRequests; i++ {
		request := types.MCPRequest{
			JSONRPC: "2.0",
			Method:  "tools/call",
			Params: &types.RequestParams{
				Name: "query_database",
				Arguments: map[string]interface{}{
					"database_id": "db123456",
				},
			},
			ID: i + 1,
		}

		requestBody, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request %d: %v", i, err)
		}

		// TODO: Send request
		// resp, err := http.Post(proxyServer.URL+"/notion", "application/json", bytes.NewBuffer(requestBody))
		// ... parse response ...
		// responses = append(responses, parsedResponse)

		_ = requestBody // Use the variable
	}

	t.Skip("Proxy server not yet implemented - will implement in T026-T030")

	// TODO: Verify all responses have same fields filtered
	// if len(responses) < 2 {
	//     t.Fatal("Not enough responses to compare")
	// }
	//
	// firstFields := responses[0].Meta.FieldSelection.FieldsKept
	// for i, resp := range responses[1:] {
	//     if !slicesEqual(firstFields, resp.Meta.FieldSelection.FieldsKept) {
	//         t.Errorf("Response %d has different fields than first response", i+1)
	//     }
	// }

	_ = responses // Use the variable
}

// Helper functions (will be moved to testutil later)

func setupProxyServer(mockServerURL string) *httptest.Server {
	// TODO: Implement proxy server setup
	return nil
}

func setupProxyServerWithStrategy(mockServerURL, strategy string) *httptest.Server {
	// TODO: Implement proxy server setup with specific strategy
	return nil
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]bool)
	for _, item := range a {
		aMap[item] = true
	}

	for _, item := range b {
		if !aMap[item] {
			return false
		}
	}

	return true
}
