# Plan: Tool Management Client Surface

## Summary
Add the missing Band Go SDK tool-management client surface so callers can create, read, update, delete, list, assign, list assigned, and remove tools through public `Client` methods. Done means `client/tools.go` exports the eleven required methods with contract-shaped Go request/response types, `client/tools_test.go` covers paths, HTTP methods, JSON bodies, and 4xx propagation, and the validation commands pass. Implementation waits for human approval after this planning stage.

## Codebase Context
Existing client methods live in small files by surface area and all call `Client.Do`; see `client/chats.go:28`, `client/chats.go:42`, `client/chats.go:57`, `client/contacts.go:86`, `client/contacts.go:99`, and `client/memories.go:77`. Path identifiers are escaped with `url.PathEscape` before being appended, as in `client/chats.go:50`, `client/memories.go:110`, and `client/participants.go:62`.

`Client.Do` builds the request, sets `Accept`, `X-API-Key`, optional `User-Agent`, JSON `Content-Type`, decodes 2xx JSON responses, skips decode for `204 No Content`, and maps non-2xx responses through `newAPIError`; see `client/client.go:42` and `client/client.go:47`.

Pagination helpers already exist in `client/types.go:9`, `client/types.go:17`, `client/types.go:22`, and `client/types.go:31`. Existing endpoint tests use `roundTripFunc` fakes from `client/client_test.go:303` and verify exact methods, URLs, response decoding, and error propagation, as in `client/chats_test.go:17`, `client/chats_test.go:119`, `client/agents_test.go:84`, and `client/messages_test.go:135`.

Local grep gate result: no existing local symbols named `ListTools`, `CreateTool`, `GetTool`, `UpdateTool`, `DeleteTool`, `AssignToolsToAgent`, `ListAgentTools`, `RemoveToolFromAgent`, `ListMyTools`, `AssignToolsToMyself`, `RemoveToolFromMyself`, `ToolListItem`, `AssignedTool`, `AssignedToolDetail`, `ConnectionConfig`, `AuthConfig`, `AssignToolsInput`, `ToolPagination`, `ListToolsResponse`, or `Tool`.

## Upstream Contract Citations
Both upstream tool contracts publish these operations under `/api/v2`, while the path items themselves are bare paths such as `/tools` and `/agents/{agent_id}/tools`:

```yaml
# /Users/pp/thenvoi/product-docs-vault/PRDs/API/Alpha Release API Implementation Specs/thenvoi-api-v2-openapi.yaml:82-86
servers:
  - url: https://api.thenvoi.com/api/v2
    description: Production server
  - url: https://staging-api.thenvoi.com/api/v2
    description: Staging server
```

```yaml
# /Users/pp/thenvoi/product-docs-vault/PRDs/API/Alpha Release API Implementation Specs/tool-api-openapi.yaml:82-86
servers:
  - url: https://api.thenvoi.com/api/v2
    description: Production server
  - url: https://staging-api.thenvoi.com/api/v2
    description: Staging server
```

Endpoint operations from `thenvoi-api-v2-openapi.yaml`:

```yaml
# /Users/pp/thenvoi/product-docs-vault/PRDs/API/Alpha Release API Implementation Specs/thenvoi-api-v2-openapi.yaml:339-445
/agents/me/tools:
  get:
    operationId: listMyTools
    parameters:
      - $ref: '#/components/parameters/PageParam'
      - $ref: '#/components/parameters/PerPageParam'
    responses:
      '200':
        schema:
          properties:
            data:
              type: array
              items:
                $ref: '#/components/schemas/AssignedToolDetail'
            pagination:
              $ref: '#/components/schemas/Pagination'
  post:
    operationId: assignToolsToMyself
    requestBody:
      required: true
      schema:
        required:
          - tool_ids
        properties:
          tool_ids:
            type: array
            items:
              type: string
              format: uuid
    responses:
      '200':
        schema:
          properties:
            data:
              type: object
              properties:
                assigned_tools:
                  type: array
                  items:
                    type: object
                    properties:
                      id:
                        type: string
                        format: uuid
                      name:
                        type: string
/agents/me/tools/{tool_id}:
  delete:
    operationId: removeToolFromMyself
    parameters:
      - $ref: '#/components/parameters/ToolIdParam'
    responses:
      '204':
        description: Tool successfully removed
```

```yaml
# /Users/pp/thenvoi/product-docs-vault/PRDs/API/Alpha Release API Implementation Specs/thenvoi-api-v2-openapi.yaml:868-1105
/agents/{agent_id}/tools:
  post:
    operationId: assignToolsToAgent
    parameters:
      - $ref: '#/components/parameters/AgentIdParam'
    requestBody:
      $ref: '#/components/requestBodies/AssignTools'
    responses:
      '200':
        schema:
          properties:
            data:
              properties:
                agent_id:
                  type: string
                  format: uuid
                assigned_tools:
                  type: array
                  items:
                    properties:
                      id:
                        type: string
                        format: uuid
                      name:
                        type: string
  get:
    operationId: listAgentTools
    parameters:
      - $ref: '#/components/parameters/AgentIdParam'
      - $ref: '#/components/parameters/PageParam'
      - $ref: '#/components/parameters/PerPageParam'
    responses:
      '200':
        schema:
          properties:
            data:
              type: array
              items:
                $ref: '#/components/schemas/AssignedTool'
            pagination:
              $ref: '#/components/schemas/Pagination'
/agents/{agent_id}/tools/{tool_id}:
  delete:
    operationId: removeToolFromAgent
    parameters:
      - $ref: '#/components/parameters/AgentIdParam'
      - $ref: '#/components/parameters/ToolIdParam'
    responses:
      '204':
        description: Tool successfully removed from agent
/tools:
  get:
    operationId: listTools
    parameters:
      - $ref: '#/components/parameters/PageParam'
      - $ref: '#/components/parameters/PerPageParam'
    responses:
      '200':
        schema:
          properties:
            data:
              type: array
              items:
                $ref: '#/components/schemas/ToolListItem'
  post:
    operationId: createTool
    requestBody:
      $ref: '#/components/requestBodies/CreateTool'
    responses:
      '201':
        schema:
          properties:
            data:
              $ref: '#/components/schemas/Tool'
/tools/{tool_id}:
  get:
    operationId: getTool
    parameters:
      - $ref: '#/components/parameters/ToolIdParam'
    responses:
      '200':
        schema:
          properties:
            data:
              $ref: '#/components/schemas/Tool'
  put:
    operationId: updateTool
    parameters:
      - $ref: '#/components/parameters/ToolIdParam'
    requestBody:
      $ref: '#/components/requestBodies/UpdateTool'
    responses:
      '200':
        schema:
          properties:
            data:
              $ref: '#/components/schemas/Tool'
  delete:
    operationId: deleteTool
    parameters:
      - $ref: '#/components/parameters/ToolIdParam'
    responses:
      '204':
        description: Tool successfully deleted
```

Path and query parameter names from `thenvoi-api-v2-openapi.yaml`:

```yaml
# /Users/pp/thenvoi/product-docs-vault/PRDs/API/Alpha Release API Implementation Specs/thenvoi-api-v2-openapi.yaml:2400-2448
AgentIdParam:
  name: agent_id
  in: path
  schema:
    type: string
    format: uuid
ToolIdParam:
  name: tool_id
  in: path
  schema:
    type: string
    format: uuid
PageParam:
  name: page
  in: query
PerPageParam:
  name: per_page
  in: query
```

Public type mirrors:

```text
Go Tool mirrors upstream Tool.
Go ToolListItem mirrors upstream ToolListItem.
Go AssignedTool mirrors upstream AssignedTool.
Go AssignedToolDetail mirrors upstream AssignedToolDetail.
Go ConnectionConfig mirrors upstream ConnectionConfig.
Go AuthConfig mirrors upstream AuthConfig and the discriminator-backed auth variants with optional fields.
Go CreateToolInput mirrors upstream CreateTool request body.
Go UpdateToolInput mirrors upstream UpdateTool request body with every field optional.
Go AssignToolsInput mirrors upstream AssignTools request body and inline assignToolsToMyself body.
Go AssignToolsToAgentResponse mirrors the assignToolsToAgent response body, including `agent_id`.
Go AssignToolsToMyselfResponse mirrors the assignToolsToMyself response body, without `agent_id`.
```

Tool, ToolListItem, AssignedTool, AssignedToolDetail, and ConnectionConfig from `thenvoi-api-v2-openapi.yaml`:

```yaml
# /Users/pp/thenvoi/product-docs-vault/PRDs/API/Alpha Release API Implementation Specs/thenvoi-api-v2-openapi.yaml:2617-2845
Tool:
  required:
    - id
    - name
    - description
    - json_schema
    - connection_config
    - owner_uuid
    - inserted_at
    - updated_at
  properties:
    id:
      type: string
      format: uuid
    name:
      type: string
    description:
      type: string
    json_schema:
      type: object
    connection_config:
      $ref: '#/components/schemas/ConnectionConfig'
    owner_uuid:
      type: string
      format: uuid
    organization_id:
      type: ["string", "null"]
    inserted_at:
      type: string
      format: date-time
    updated_at:
      type: string
      format: date-time
ToolListItem:
  description: Tool summary for list views (excludes json_schema and connection_config)
  properties:
    id:
      type: string
      format: uuid
    name:
      type: string
    description:
      type: string
    owner_uuid:
      type: string
      format: uuid
    organization_id:
      type: ["string", "null"]
    inserted_at:
      type: string
      format: date-time
    updated_at:
      type: string
      format: date-time
AssignedTool:
  required:
    - id
    - name
    - description
    - assigned_at
  properties:
    id:
      type: string
      format: uuid
    name:
      type: string
    description:
      type: string
    assigned_at:
      type: string
      format: date-time
AssignedToolDetail:
  required:
    - id
    - name
    - description
    - json_schema
    - assigned_at
  properties:
    id:
      type: string
      format: uuid
    name:
      type: string
    description:
      type: string
    json_schema:
      type: object
    assigned_at:
      type: string
      format: date-time
ConnectionConfig:
  required:
    - base_url
    - method
    - path
    - param_type
    - auth
  properties:
    base_url:
      type: string
      format: uri
    method:
      type: string
      enum: ["GET", "POST", "PUT", "PATCH", "DELETE"]
    path:
      type: string
    param_type:
      type: string
      enum: ["query", "body", "path"]
    auth:
      $ref: '#/components/schemas/AuthConfig'
```

Auth and pagination shapes from `tool-api-openapi.yaml`:

```yaml
# /Users/pp/thenvoi/product-docs-vault/PRDs/API/Alpha Release API Implementation Specs/tool-api-openapi.yaml:549-639
ConnectionConfig:
  required:
    - base_url
    - method
    - path
    - param_type
    - auth
AuthConfig:
  required:
    - type
  properties:
    type:
      enum: ["api_key", "bearer", "basic", "vercel_bypass", "none"]
    location:
      enum: ["header", "query"]
    header_name:
      type: string
    key_name:
      type: string
ApiKeyAuth:
  required:
    - location
    - header_name
    - key_name
BearerAuth:
  required:
    - key_name
BasicAuth:
  required:
    - key_name
VercelBypassAuth:
  required:
    - key_name
NoAuth:
  properties:
    type:
      const: "none"
```

```yaml
# /Users/pp/thenvoi/product-docs-vault/PRDs/API/Alpha Release API Implementation Specs/tool-api-openapi.yaml:703-725
Pagination:
  required:
    - page
    - per_page
    - total_pages
    - total_items
  properties:
    page:
      type: integer
    per_page:
      type: integer
    total_pages:
      type: integer
    total_items:
      type: integer
```

Create, update, and assignment request bodies from `tool-api-openapi.yaml`:

```yaml
# /Users/pp/thenvoi/product-docs-vault/PRDs/API/Alpha Release API Implementation Specs/tool-api-openapi.yaml:817-970
CreateTool:
  required:
    - name
    - description
    - json_schema
    - connection_config
  properties:
    name:
      type: string
      pattern: '^[a-z0-9_]+$'
    description:
      type: string
    json_schema:
      type: object
    connection_config:
      $ref: '#/components/schemas/ConnectionConfig'
UpdateTool:
  properties:
    name:
      type: string
      pattern: '^[a-z0-9_]+$'
    description:
      type: string
    json_schema:
      type: object
    connection_config:
      $ref: '#/components/schemas/ConnectionConfig'
AssignTools:
  required:
    - tool_ids
  properties:
    tool_ids:
      type: array
      items:
        type: string
        format: uuid
      minItems: 1
      uniqueItems: true
```

## Goals / Non-Goals / Constraints / Risks
Goals: add the eleven exact public methods requested by name; add contract-shaped public Go types for tool resources, auth config, connection config, pagination, list responses, mutation responses, and request inputs; test every method through fake transports; preserve existing client transport and error behavior.

Non-goals: generated OpenAPI client code, live integration tests, README updates, changelog updates, PR creation, or changes outside `client/tools.go` and `client/tools_test.go` unless Go formatting touches those new files.

Constraints: implementation must not start until human approval. Do not modify `CHANGELOG.md`. Do not modify pre-existing `client/*.go` files. Use `Client.Do` for all calls. Use `/api/v2` for these new tool operations because both cited upstream specs declare `/api/v2` server URLs for the bare operation paths. That means `/tools` maps to `/api/v2/tools`, `/agents/{agent_id}/tools` maps to `/api/v2/agents/{agent_id}/tools`, and `/agents/me/tools` maps to `/api/v2/agents/me/tools`. Use `per_page`, not `page_size`, for the new tool endpoints because the cited tool contracts define `PerPageParam`.

Risks: the repo currently has `PaginationMetadata` with `page_size`/`total_count`, while the tool contracts use `Pagination` with `per_page`/`total_items`; adding a tool-specific pagination type avoids field loss. The contract uses generic object schemas for `json_schema`, so `map[string]any` is the practical SDK shape. Auth variants share one discriminator schema, so one `AuthConfig` struct with optional fields is less error-prone than several one-off types.

## Files / Surfaces Expected To Change
- `client/tools.go` — new exported types and all eleven `Client` methods.
- `client/tools_test.go` — method/path/body/error tests using `roundTripFunc`.
- `notes/plan.md` — this plan.

## Implementation Approach
Create `client/tools.go` with `context`, `errors`, `net/http`, `net/url`, and `time` imports. Define `Tool`, `ToolListItem`, `AssignedTool`, `AssignedToolDetail`, `ConnectionConfig`, `AuthConfig`, `ToolPagination`, `ListToolsInput`, `ListToolsResponse`, `ListAgentToolsInput`, `ListAgentToolsResponse`, `ListMyToolsInput`, `ListMyToolsResponse`, `CreateToolInput`, `UpdateToolInput`, `AssignToolsInput`, `AssignedToolSummary`, `AssignToolsToAgentResponse`, and `AssignToolsToMyselfResponse` shapes. Keep all field names aligned to the contract JSON names.

Use `map[string]any` for `json_schema`, `*ConnectionConfig` in update input so omitted config stays omitted, pointer strings for optional update fields, and `*string` for nullable `organization_id`. Use `time.Time` for required timestamp fields, matching existing client types. Keep `AuthConfig.Type` as required by using a plain `string`, while `Location`, `HeaderName`, and `KeyName` stay pointer strings because their requiredness depends on the auth variant. Validate variant-specific auth requirements in `CreateTool` when `ConnectionConfig.Auth.Type` is `api_key`, `bearer`, `basic`, or `vercel_bypass`; for `UpdateTool`, validate only when a replacement `ConnectionConfig` is supplied.

Implement the public methods with receiver spelling `func (client *Client) Name...` to satisfy the acceptance grep. Validate required path IDs before making requests with messages such as `thenvoi: tool id is required` and `thenvoi: agent id is required`. Validate required request fields for create and assignment before transport calls, following `CreateMemory` and contact input patterns.

Build query strings with `page` and `per_page` for `ListTools`, `ListAgentTools`, and `ListMyTools`. Use `url.PathEscape` for `agentID` and `toolID`. Return decoded resource pointers for single-resource responses, response structs for list and assignment endpoints, and `error` only for delete endpoints. Tool requests must pass `/api/v2/...` paths to `Client.Do`.

Write `client/tools_test.go` with table-driven tests where that keeps duplication low. Each required method will be called at least once and the fake transport will assert the HTTP method and exact `/api/v2/...` URL path using placeholder IDs. POST and PUT tests will decode request JSON and assert `tool_ids`, `name`, `json_schema`, and `connection_config` shape. One 4xx propagation test will use a representative tool method and assert `errors.Is(err, client.ErrUnauthorized)` or the matching existing API error behavior; additional 4xx cases can be added for delete or update if the implementation stays concise.

## Verification Strategy
Validation proves the implementation at three levels: compilation catches exported type/method mistakes, vet catches malformed tests or suspicious code, and the unit suite catches request construction, JSON bodies, response decoding, and non-2xx propagation. The grep command verifies all eleven operationIds map to exactly one expected public method in `client/tools.go`.

## verificationCommands
- `go build ./...`
- `go vet ./...`
- `go test ./...`
- `bash -c 'cd client && grep -lE "^func \\(client \\*Client\\) (ListTools|CreateTool|GetTool|UpdateTool|DeleteTool|AssignToolsToAgent|ListAgentTools|RemoveToolFromAgent|ListMyTools|AssignToolsToMyself|RemoveToolFromMyself)\\b" tools.go'`

## Rollback / Safety Notes
Rollback is deleting `client/tools.go` and `client/tools_test.go`, leaving the rest of the SDK unchanged. Since the new surface is additive and no existing client files change, the primary compatibility risk is limited to exported type names chosen for the new API.

## Open Questions
None blocking. The plan no longer assumes `/api/v1`; the cited tool specs require `/api/v2` for these operation paths.
