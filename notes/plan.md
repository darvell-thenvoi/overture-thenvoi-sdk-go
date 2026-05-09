# Plan: Task Operations Surface for the Go SDK

## Summary
Add the missing task-related client methods for the Band Go SDK so it covers every task operation named in `thenvoi-api-v2-openapi.yaml`. Done means `client/tasks.go` exposes the six agent-self chat-scoped task methods and four user-scoped task methods, `client/chats.go` exposes `ListMyChats`, tests cover each method with the existing `roundTripFunc` fake transport style, and the validation commands pass. Engineering must wait for human approval of this plan before code implementation begins.

## Codebase Context
`client/client.go:31` exposes `Client.Do(ctx, method, path, body, out)`, which every new public method should call. `client/client.go:35` has `client.do`, which is only needed for status-sensitive responses; task deletes can use `Do` with `out == nil` because 204 is already handled.

`client/chats.go:10` defines `ListChatRoomsInput` with `Page` and `PageSize`; `client/chats.go:28` shows the current optional input pointer pattern and query construction. `client/chats.go:40` and `client/messages.go:37` show path parameter validation with `errors.New("thenvoi: chat id is required")`.

`client/types.go:8` defines `PaginationMetadata` using `page_size`; the v2 contract uses `per_page`, so the task surface should add a v2 pagination metadata/input shape rather than changing the existing v1 helpers. `client/types.go:60` defines `ChatRoom` but lacks v2 `status`, `type`, `metadata`, and `deleted_at`; `ListMyChats` should extend that struct with pointer fields so existing v1 tests continue to decode.

`client/chats_test.go:13` and `client/messages_test.go:193` show request assertions, query assertions, response decoding, and pointer input setup using `roundTripFunc`. `client/client_test.go:303` defines `roundTripFunc`; `client/client_test.go:309` defines `jsonResponse`.

Local grep found no existing public `Client` methods named `ListMyTasksInChat`, `CreateTaskInChat`, `CreateBulkTasksInChat`, `GetTask`, `UpdateTaskInChat`, `DeleteTaskInChat`, `ListMyUserTasks`, `CreateUserTask`, `GetUserTask`, `DeleteUserTask`, or `ListMyChats`. Local grep found only `TaskID` fields in `client/types.go:63` and `client/chats.go:24`, so new `Task` and task input/output types will not duplicate existing types.

## Upstream Contract Citations
Source file: `/Users/pp/thenvoi/product-docs-vault/PRDs/API/Alpha Release API Implementation Specs/thenvoi-api-v2-openapi.yaml`.

Endpoints and response envelope citations:

```yaml
# Lines 1744-1810
/agents/me/chats/{chat_id}/tasks:
  get:
    operationId: listMyTasksInChat
    parameters:
      - name: chat_id
        in: path
        required: true
      - name: status
        in: query
      - name: page
        in: query
      - name: per_page
        in: query
      - name: include_subtasks
        in: query
    responses:
      '200':
        properties:
          data:
            type: array
            items:
              $ref: '#/components/schemas/Task'
          pagination:
            $ref: '#/components/schemas/Pagination'
```

```yaml
# Lines 1819-1886
post:
  operationId: createTaskInChat
  requestBody:
    required: true
    content:
      application/json:
        schema:
          required:
            - description
          properties:
            description: {type: string, minLength: 1, maxLength: 500}
            title: {type: string, maxLength: 200}
            summary: {type: string, maxLength: 2000}
            parent_task_id: {type: string, format: uuid}
            metadata: {type: object, additionalProperties: true}
            status:
              type: string
              enum: ["new", "waiting", "engaging", "in_progress", "reviewing", "completed", "failed"]
  responses:
    '201':
      properties:
        data:
          $ref: '#/components/schemas/Task'
```

```yaml
# Lines 1898-1954
/agents/me/chats/{chat_id}/tasks/bulk:
  post:
    operationId: createBulkTasksInChat
    requestBody:
      required:
        - tasks
      properties:
        tasks:
          type: array
          minItems: 1
          maxItems: 100
          items:
            required:
              - description
            properties:
              description: {type: string, minLength: 1, maxLength: 500}
              title: {type: string, maxLength: 200}
              summary: {type: string, maxLength: 2000}
              status:
                type: string
                enum: ["new", "waiting", "engaging", "in_progress", "reviewing", "completed", "failed"]
              metadata: {type: object, additionalProperties: true}
    responses:
      '201':
        properties:
          data:
            properties:
              created: {type: integer}
              tasks:
                type: array
```

```yaml
# Lines 1986-2162
/agents/me/chats/{chat_id}/tasks/{task_id}:
  get:
    operationId: getTask
    responses:
      '200':
        properties:
          data:
            $ref: '#/components/schemas/Task'
  put:
    operationId: updateTaskInChat
    requestBody:
      properties:
        status:
          type: string
          enum: ["new", "waiting", "engaging", "in_progress", "reviewing", "completed", "failed"]
        title: {type: string, maxLength: 200}
        summary: {type: string, maxLength: 2000}
        metadata: {type: object, additionalProperties: true}
        artifacts:
          type: array
          items:
            required:
              - artifact_id
              - name
              - url
            properties:
              artifact_id: {type: string, format: uuid}
              name: {type: string}
              url: {type: string, format: uri}
              mime_type: {type: string}
              description: {type: string}
              metadata: {type: object, additionalProperties: true}
    responses:
      '200':
        properties:
          data:
            $ref: '#/components/schemas/Task'
  delete:
    operationId: deleteTaskInChat
    responses:
      '204':
        description: Task successfully deleted
```

```yaml
# Lines 2180-2225
/me/chats:
  get:
    operationId: listMyChats
    parameters:
      - name: status
        in: query
        schema:
          enum: ["active", "archived", "closed"]
      - name: page
        in: query
      - name: per_page
        in: query
    responses:
      '200':
        properties:
          data:
            type: array
            items:
              $ref: '#/components/schemas/ChatRoom'
          metadata:
            $ref: '#/components/schemas/PaginationMetadata'
```

```yaml
# Lines 2237-2382
/me/tasks:
  get:
    operationId: listMyUserTasks
    parameters:
      - name: status
        in: query
      - name: page
        in: query
      - name: per_page
        in: query
    responses:
      '200':
        properties:
          data:
            type: array
            items:
              $ref: '#/components/schemas/Task'
          metadata:
            $ref: '#/components/schemas/PaginationMetadata'
  post:
    operationId: createUserTask
    requestBody:
      required:
        - description
      properties:
        description: {type: string, minLength: 1, maxLength: 500}
        summary: {type: string, maxLength: 2000}
        agent_id: {type: string, format: uuid}
        metadata: {type: object, additionalProperties: true}
    responses:
      '201':
        properties:
          data:
            $ref: '#/components/schemas/Task'
/me/tasks/{task_id}:
  get:
    operationId: getUserTask
    responses:
      '200':
        properties:
          data:
            $ref: '#/components/schemas/Task'
  delete:
    operationId: deleteUserTask
    responses:
      '204':
        description: Task deleted successfully
```

Public mirrored types:

```yaml
# Lines 3040-3134
Task:
  required:
    - id
    - title
    - status
    - description
    - user_id
    - inserted_at
    - updated_at
  properties:
    id: {type: string, format: uuid}
    title: {type: string}
    status:
      type: string
      enum: ["new", "waiting", "engaging", "in_progress", "reviewing", "completed", "failed"]
    description: {type: string}
    summary: {type: ["string", "null"]}
    agent_id: {type: ["string", "null"], format: uuid}
    chat_room_id: {type: ["string", "null"], format: uuid}
    parent_task_id: {type: ["string", "null"], format: uuid}
    execution_id: {type: ["string", "null"], format: uuid}
    participants:
      type: array
      items: {type: string, format: uuid}
    metadata: {type: ["object", "null"], additionalProperties: true}
    user_id: {type: string, format: uuid}
    inserted_at: {type: string, format: date-time}
    updated_at: {type: string, format: date-time}
```

```yaml
# Lines 3136-3195
ChatRoom:
  required:
    - id
    - title
    - status
    - type
    - inserted_at
    - updated_at
  properties:
    id: {type: string, format: uuid}
    title: {type: string}
    status:
      type: string
      enum: ["active", "archived", "closed"]
    type:
      type: string
      enum: ["direct", "group", "task"]
    metadata: {type: ["object", "null"], additionalProperties: true}
    task_id: {type: ["string", "null"], format: uuid}
    deleted_at: {type: ["string", "null"], format: date-time}
    inserted_at: {type: string, format: date-time}
    updated_at: {type: string, format: date-time}
```

```yaml
# Lines 3327-3350
PaginationMetadata:
  required:
    - page
    - per_page
    - total_count
    - total_pages
  properties:
    page: {type: integer}
    per_page: {type: integer}
    total_count: {type: integer}
    total_pages: {type: integer}
```

```yaml
# Lines 3577-3598
Pagination:
  required:
    - page
    - per_page
    - total_pages
    - total_items
  properties:
    page: {type: integer}
    per_page: {type: integer}
    total_pages: {type: integer}
    total_items: {type: integer}
```

Side-by-side public type mapping for validation:

`#/components/schemas/Task` -> `client.Task`
`#/components/schemas/ChatRoom` -> existing `client.ChatRoom`, extended for v2 fields
`#/components/schemas/PaginationMetadata` -> new `client.V2PaginationMetadata`
`#/components/schemas/Pagination` -> new `client.TaskPagination`
Inline create task request body -> `client.CreateTaskInChatInput`
Inline bulk task request body -> `client.CreateBulkTasksInChatInput` and `client.BulkTaskInput`
Inline bulk task response body -> `client.CreateBulkTasksInChatResponse` and `client.BulkTask` with `ID`, `Description`, and `Status` from `data.tasks[]`
Inline update task request body artifact -> `client.TaskArtifactInput`
Inline user create task request body -> `client.CreateUserTaskInput`

## Goals / Non-Goals / Constraints / Risks
Goals: add exactly one public `Client` method for each requested Go name, use `client.Do` for all transport, use v2 contract paths under `/api/v2`, use `url.PathEscape` for `chatID` and `taskID`, use optional pointer inputs for list filters, and add per-call tests that assert method, URL, JSON body, decoding, required ID validation, and 204 delete handling.

Non-goals: do not modify `CHANGELOG.md`, do not modify `client/tools.go`, do not change existing v1 methods, do not push a branch or open a PR.

Constraints: implement field shapes from the cited v2 contract, preserve existing client test behavior, use `per_page` for the new v2 list methods, keep `page_size` behavior for existing v1 methods, and do not start implementation until human approval is recorded.

Risks: the spec has two pagination object names with different total fields, so tests should pin both `pagination.total_items` for `ListMyTasksInChat` and `metadata.total_count` for `ListMyUserTasks`/`ListMyChats`. The existing `ChatRoom` struct is v1-shaped, so v2 fields must be added as optional fields to avoid breaking older fixtures.

## Files / Surfaces Expected To Change
- `client/tasks.go` — new task models, v2 pagination models, chat-scoped task methods, and user-scoped task methods.
- `client/tasks_test.go` — new `roundTripFunc` tests for every task method and task body/response shape.
- `client/chats.go` — add `ListMyChatsInput`, `ListMyChatsResponse`, and `ListMyChats`.
- `client/chats_test.go` — add coverage for `ListMyChats`.
- `client/types.go` — extend existing `ChatRoom` with optional v2 fields: `Status`, `Type`, `Metadata`, and `DeletedAt`.
- `notes/plan.md` — planning artifact only.

## Implementation Approach
Create `client/tasks.go` in package `client` with `TaskStatus` as a string alias only if it helps readability without complicating callers; plain `string` fields are acceptable and match current SDK style. Define `Task` with `time.Time` for required timestamps, `*string` for nullable string fields, `[]string` for participants, and `map[string]any` for metadata. Define `V2PaginationInput` with `Page` and `PerPage`, `V2PaginationMetadata` with `Page`, `PerPage`, `TotalCount`, `TotalPages`, and `TaskPagination` with `Page`, `PerPage`, `TotalPages`, `TotalItems`.

Implement `ListMyTasksInChat(ctx, chatID string, input *ListMyTasksInChatInput) (*ListMyTasksInChatResponse, error)` against `GET /api/v2/agents/me/chats/{chat_id}/tasks`. Query fields are `status`, `page`, `per_page`, and `include_subtasks`. The response uses `Data []Task` and `Pagination TaskPagination`.

Implement `CreateTaskInChat(ctx, chatID string, input CreateTaskInChatInput) (*Task, error)` against `POST /api/v2/agents/me/chats/{chat_id}/tasks` with JSON envelope-free body matching the contract fields. Validate `chatID` and required `Description`. Keep optional fields as `*string` or `map[string]any` with `omitempty`.

Implement `CreateBulkTasksInChat(ctx, chatID string, input CreateBulkTasksInChatInput) (*CreateBulkTasksInChatResponse, error)` against `POST /api/v2/agents/me/chats/{chat_id}/tasks/bulk`. Validate `chatID` and that at least one task has non-empty `Description`. Decode `data.created` and `data.tasks`.

`BulkTask` should mirror the inline response item schema at contract lines 1958-1970: `id`, `description`, and `status`. It should not reuse `Task`, because the bulk response omits most `Task` fields.

Implement `GetTask`, `UpdateTaskInChat`, and `DeleteTaskInChat` against `/api/v2/agents/me/chats/{chat_id}/tasks/{task_id}`. Validate both IDs. `UpdateTaskInChatInput` should include optional `Status`, `Title`, `Summary`, `Metadata`, and `Artifacts []TaskArtifactInput`; `DeleteTaskInChat` returns only `error`.

Implement user-scoped methods in the same file: `ListMyUserTasks` on `GET /api/v2/me/tasks` with `status`, `page`, and `per_page`; `CreateUserTask` on `POST /api/v2/me/tasks` with required `Description`, optional `Summary`, `AgentID`, and `Metadata`; `GetUserTask` and `DeleteUserTask` on `/api/v2/me/tasks/{task_id}` with task ID validation.

Extend `client/chats.go` with `ListMyChats(ctx context.Context, input *ListMyChatsInput) (*ListMyChatsResponse, error)` using `GET /api/v2/me/chats`, query fields `status`, `page`, and `per_page`, and response `Data []ChatRoom` plus `Metadata V2PaginationMetadata`. Extend `client.ChatRoom` with v2 fields in `client/types.go`.

Tests should follow the current package `client_test` style. They should assert exact URLs including escaped IDs and v2 query names, decode sample timestamps and metadata, inspect JSON request bodies for create/update/bulk methods, assert delete methods do not require a response body, and assert missing `chatID`, `taskID`, and required descriptions return local errors without calling transport.

## Verification Strategy
Validation will run Go build, vet, and test across the module. The grep command proves every requested operation name has exactly one public `func (client *Client) <Name>` in `client/*.go`; engineering should also review the grep output count during implementation because the command supplied by the work order lists files, not per-method counts. Unit tests should cover all new methods so regressions in path shape, body shape, and response decoding are caught without network access.

## verificationCommands
- `go build ./...`
- `go vet ./...`
- `go test ./...`
- `bash -c 'cd client && grep -lE "^func \\(client \\*Client\\) (ListMyTasksInChat|CreateTaskInChat|CreateBulkTasksInChat|GetTask|UpdateTaskInChat|DeleteTaskInChat|ListMyUserTasks|CreateUserTask|GetUserTask|DeleteUserTask|ListMyChats)\\b" *.go'`

## Rollback / Safety Notes
The change is additive except for optional fields added to `ChatRoom`; rollback can remove `client/tasks.go`, `client/tasks_test.go`, the new `ListMyChats` additions, and the optional `ChatRoom` fields. No data migration or external side effects are involved. Keep existing v1 paths untouched.

## Open Questions
No implementation-detail question is blocking. Human approval of this plan is required before engineering starts.
