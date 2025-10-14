# Protocol Extension Specification

Formal specification for the `_meta.fieldSelection` extension to the Model Context Protocol.

## Status

**Status:** Draft Proposal  
**Version:** 1.0.0  
**Last Updated:** October 2025

This is a proposed extension to the MCP specification. It is currently implemented in SlimMCP as a reference implementation.

## Abstract

This specification defines an optional extension to the Model Context Protocol (MCP) that enables clients to request specific fields in tool responses, reducing token consumption and improving context window efficiency.

## Motivation

Modern MCP servers return complete objects with all available fields, regardless of client needs. This creates several problems:

1. **Token Waste** - Clients receive 60-85% more data than needed
2. **Context Pollution** - Valuable context window space is consumed by unused data
3. **Cost Impact** - Higher API costs due to unnecessary tokens
4. **Performance** - Larger responses increase network transfer time

### Example Problem

A client needs only the `title` and `status` of a Notion page:

**Without field selection:**
```json
{
  "id": "...",
  "title": "Planning",
  "status": "Done",
  "created_time": "...",
  "last_edited_time": "...",
  "properties": { /* 40+ fields */ },
  // ... 35+ more unused fields
}
// Total: 2,847 tokens
```

**With field selection:**
```json
{
  "id": "...",
  "title": "Planning",
  "status": "Done",
  "url": "...",
  "type": "page"
}
// Total: 421 tokens (85% savings)
```

## Design Principles

1. **Optional** - Servers can ignore if not supported
2. **Backward Compatible** - Existing clients work unchanged
3. **Client-Driven** - Clients specify their needs
4. **Graceful Degradation** - Failures return full responses
5. **Progressive Enhancement** - Easy to adopt incrementally

## Specification

### Request Extension

Clients MAY include a `_meta` object in MCP requests with field selection parameters:

```typescript
interface MCPRequest {
  jsonrpc: "2.0";
  method: string;
  params: object;
  id: number | string;
  
  _meta?: {
    fieldSelection?: FieldSelectionRequest;
  };
}

interface FieldSelectionRequest {
  // Fields to include
  include?: string[];
  
  // Fields to exclude
  exclude?: string[];
  
  // Filtering strategy
  strategy?: "explicit" | "smart" | "minimal" | "passthrough";
  
  // Keep fields related to included fields
  keepRelated?: boolean;
  
  // Maximum nesting depth
  maxDepth?: number;
}
```

### Field Path Syntax

Field names support dot notation for nested fields:

```
"title"                          // Top-level field
"properties.status"              // Nested field
"properties.assignee.email"      // Deep nested field
"items[].title"                  // Array items
"items[*].metadata.tags"         // All array items
```

### Response Extension

Servers that apply field selection SHOULD include metadata in responses:

```typescript
interface MCPResponse {
  jsonrpc: "2.0";
  result: any;
  id: number | string;
  
  _meta?: {
    fieldSelection?: FieldSelectionResponse;
  };
}

interface FieldSelectionResponse {
  // Was filtering applied?
  applied: boolean;
  
  // Strategy used
  strategy: "explicit" | "smart" | "minimal" | "passthrough";
  
  // Who applied filtering
  filteredBy: "server" | "proxy" | "none";
  
  // Size metrics (optional)
  originalSize?: number;    // bytes or tokens
  filteredSize?: number;    // bytes or tokens
  savings?: number;         // percentage
  
  // Field details (optional)
  fields?: {
    kept: string[];
    removed: number;
  };
  
  // Performance (optional)
  filteringTime?: number;   // milliseconds
}
```

### Capability Announcement

Servers SHOULD announce field selection support in capabilities:

```typescript
interface ServerCapabilities {
  fieldSelection?: {
    // Does this server support field selection?
    supported: boolean;
    
    // Version of field selection spec
    version: string;
    
    // Supported strategies
    strategies: ("explicit" | "smart" | "minimal")[];
    
    // Is this native or via proxy?
    native: boolean;
  };
}
```

### Tool Schema Extension

Tool schemas MAY include field selection metadata:

```typescript
interface ToolSchema {
  name: string;
  description: string;
  inputSchema: object;
  
  // Field selection metadata
  fieldSelection?: {
    // Does this tool support field selection?
    supported: boolean;
    
    // Default fields if not specified
    defaultFields?: string[];
    
    // Field metadata
    fields?: {
      [fieldName: string]: {
        type: string;
        cost: "low" | "medium" | "high";
        required: boolean;
        relatedTo?: string[];
      };
    };
  };
}
```

## Strategies

### Explicit Strategy

Only include fields explicitly listed in `include`:

```json
{
  "_meta": {
    "fieldSelection": {
      "include": ["id", "title", "status"],
      "strategy": "explicit"
    }
  }
}
```

**Behavior:**
- Include only listed fields
- Plus critical fields: `id`, `type`, `url`
- Plus required fields per schema

### Smart Strategy

Analyze request to infer needed fields:

```json
{
  "_meta": {
    "fieldSelection": {
      "strategy": "smart"
    }
  }
}
```

**Behavior:**
- Parse request parameters for field names
- Include mentioned fields
- Include related fields
- Apply heuristics

### Minimal Strategy

Return only critical fields:

```json
{
  "_meta": {
    "fieldSelection": {
      "strategy": "minimal"
    }
  }
}
```

**Behavior:**
- Include: `id`, `type`, `url`, `title`
- Maximum token savings
- May require follow-up requests

### Passthrough Strategy

No filtering applied:

```json
{
  "_meta": {
    "fieldSelection": {
      "strategy": "passthrough"
    }
  }
}
```

**Behavior:**
- Return full response
- Useful for debugging
- Compatibility mode

## Field Categories

### Critical Fields

Always included regardless of selection:

- `id` - Object identifier
- `type` - Type discriminator
- `url` - Resource URL

### Required Fields

Fields marked as required in the schema MUST be included.

### Related Fields

Fields that reference or are referenced by included fields:

```json
{
  "include": ["assignee"],
  "keepRelated": true
}
```

Would include:
- `assignee.id`
- `assignee.name`
- `assignee.email`
- etc.

## Examples

### Example 1: Explicit Selection

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "notion-fetch",
    "arguments": {
      "id": "page-123"
    }
  },
  "id": 1,
  "_meta": {
    "fieldSelection": {
      "include": ["title", "status", "assignee"],
      "strategy": "explicit",
      "keepRelated": true
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "id": "page-123",
    "type": "page",
    "url": "https://notion.so/page-123",
    "title": "Q4 Planning",
    "status": "In Progress",
    "assignee": {
      "id": "user-456",
      "name": "Alice",
      "email": "alice@example.com"
    }
  },
  "id": 1,
  "_meta": {
    "fieldSelection": {
      "applied": true,
      "strategy": "explicit",
      "filteredBy": "server",
      "savings": 84.2
    }
  }
}
```

### Example 2: Exclusion List

**Request:**
```json
{
  "_meta": {
    "fieldSelection": {
      "exclude": ["icon", "cover", "created_by"],
      "strategy": "smart"
    }
  }
}
```

### Example 3: Nested Fields

**Request:**
```json
{
  "_meta": {
    "fieldSelection": {
      "include": [
        "id",
        "title",
        "properties.status",
        "properties.priority",
        "properties.assignee.name"
      ],
      "strategy": "explicit"
    }
  }
}
```

### Example 4: Array Handling

**Request:**
```json
{
  "_meta": {
    "fieldSelection": {
      "include": [
        "id",
        "title",
        "tasks[].id",
        "tasks[].title",
        "tasks[].status"
      ],
      "strategy": "explicit"
    }
  }
}
```

## Implementation Requirements

### Server Requirements

Servers implementing this extension MUST:

1. **Announce support** via capabilities
2. **Respect `include` list** when provided
3. **Include critical fields** always
4. **Include required fields** per schema
5. **Preserve structure** (no broken references)
6. **Add response metadata** with filtering info

Servers implementing this extension SHOULD:

1. **Respect `exclude` list** when provided
2. **Handle nested fields** with dot notation
3. **Support `keepRelated`** option
4. **Provide field metadata** in tool schemas
5. **Report token savings** in response

Servers implementing this extension MAY:

1. **Implement smart strategy**
2. **Provide field cost hints**
3. **Support custom strategies**

### Client Requirements

Clients using this extension SHOULD:

1. **Check server capabilities** before using
2. **Handle missing support** gracefully
3. **Specify strategy** explicitly
4. **Use `include` for precision**
5. **Check response metadata** for confirmation

Clients using this extension MAY:

1. **Request specific strategies**
2. **Use exclusion lists**
3. **Specify nested fields**
4. **Set keepRelated preference**

### Proxy Requirements

Proxies implementing this extension MUST:

1. **Pass through `_meta`** to servers
2. **Detect native support** per server
3. **Apply filtering** if server doesn't support
4. **Never break responses** (fallback on error)
5. **Add correct `filteredBy`** in response

## Error Handling

### Unsupported Extension

If server doesn't support field selection, it SHOULD:
- Ignore `_meta.fieldSelection`
- Return full response
- Not include `_meta` in response

### Invalid Field Names

If `include` contains unknown fields:
- Server MAY ignore unknown fields
- Server SHOULD include valid fields
- Server SHOULD NOT error

### Filtering Failure

If filtering fails:
- Server MUST return full response
- Server SHOULD log error
- Server MAY include error in `_meta`

## Security Considerations

### Information Disclosure

Field selection does not provide additional access control:
- Clients can only select from accessible fields
- Authorization is enforced before filtering
- Field selection = optimization, not security

### Denial of Service

Servers SHOULD:
- Limit `include` list size (e.g., 100 fields)
- Limit nesting depth (e.g., 10 levels)
- Timeout complex filtering operations

### Privacy

Servers SHOULD:
- Not log field selection patterns
- Not use for user tracking
- Treat as optimization hint only

## Performance Considerations

### Latency

Field selection adds minimal latency:
- Explicit: O(n) where n = number of fields
- Smart: O(n + m) where m = request complexity
- Target: <10ms for typical responses

### Bandwidth

Expected savings:
- Typical: 60-85% reduction
- Minimal strategy: 85-95% reduction
- Smart strategy: 70-85% reduction

### Server Load

Filtering is lightweight:
- CPU: <5% overhead
- Memory: Negligible (streaming)
- Network: 60-85% reduction

## Future Extensions

### Possible Future Additions

1. **Wildcard patterns** - `"properties.*"`
2. **Computed fields** - `"fullName": ["firstName", "lastName"]`
3. **Field transformations** - `"summary": {"field": "description", "maxLength": 100}`
4. **Conditional inclusion** - `"status": {"if": {"field": "isActive", "equals": true}}`
5. **Pagination hints** - Optimize for paginated responses

## Versioning

This specification uses semantic versioning:

- **Major**: Breaking changes
- **Minor**: New optional features
- **Patch**: Clarifications, fixes

Current version: **1.0.0**

## Compliance Levels

### Level 1: Basic (Minimum)

- Respect `include` list
- Include critical fields
- Add response metadata

### Level 2: Standard

- Level 1 features
- Support `exclude` list
- Handle nested fields
- Implement `minimal` strategy

### Level 3: Advanced

- Level 2 features
- Implement `smart` strategy
- Provide field cost hints
- Support `keepRelated`
- Report token savings

## Adoption Path

### For MCP Servers

1. **Phase 1**: Announce support in capabilities
2. **Phase 2**: Implement explicit strategy
3. **Phase 3**: Add smart strategy
4. **Phase 4**: Optimize with field metadata

### For MCP Clients

1. **Phase 1**: Check server capabilities
2. **Phase 2**: Use explicit selection when needed
3. **Phase 3**: Leverage smart strategy
4. **Phase 4**: Monitor savings, optimize usage

### For Proxies

1. **Phase 1**: Detect native support
2. **Phase 2**: Apply filtering for unsupported servers
3. **Phase 3**: Optimize filtering algorithm
4. **Phase 4**: Transparent pass-through when native

## References

- [Model Context Protocol Specification](https://modelcontextprotocol.io)
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [SlimMCP Reference Implementation](https://github.com/slimmcp/slimmcp)

## Acknowledgments

This specification was developed based on real-world usage patterns observed in MCP deployments and feedback from the community.

## Copyright

This specification is released under CC BY 4.0.

---

**For discussion and feedback:** https://github.com/slimmcp/slimmcp/discussions
