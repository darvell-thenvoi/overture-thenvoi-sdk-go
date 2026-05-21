package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Task describes an API v2 task.
type Task struct {
	ID           string         `json:"id"`
	AgentID      *string        `json:"agent_id"`
	ChatRoomID   *string        `json:"chat_room_id"`
	Description  string         `json:"description"`
	ExecutionID  *string        `json:"execution_id"`
	InsertedAt   time.Time      `json:"inserted_at"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	ParentTaskID *string        `json:"parent_task_id"`
	Participants []string       `json:"participants,omitempty"`
	Status       string         `json:"status"`
	Summary      *string        `json:"summary"`
	Title        string         `json:"title"`
	UpdatedAt    time.Time      `json:"updated_at"`
	UserID       string         `json:"user_id"`
}

// TaskPagination describes the API v2 chat task list pagination object.
type TaskPagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	TotalItems int `json:"total_items"`
}

// ListMyTasksInChatInput contains filters for ListMyTasksInChat.
type ListMyTasksInChatInput struct {
	Status          *string
	Page            *int
	PerPage         *int
	IncludeSubtasks *bool
}

// ListMyTasksInChatResponse contains chat-scoped tasks and pagination metadata.
type ListMyTasksInChatResponse struct {
	Data       []Task         `json:"data"`
	Pagination TaskPagination `json:"pagination"`
}

// CreateTaskInChatInput contains fields for creating a chat-scoped task.
type CreateTaskInChatInput struct {
	Description  string         `json:"description"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	ParentTaskID *string        `json:"parent_task_id,omitempty"`
	Status       *string        `json:"status,omitempty"`
	Summary      *string        `json:"summary,omitempty"`
	Title        *string        `json:"title,omitempty"`
}

// CreateBulkTasksInChatInput contains fields for creating chat-scoped tasks in bulk.
type CreateBulkTasksInChatInput struct {
	Tasks []BulkTaskInput `json:"tasks"`
}

// BulkTaskInput contains one task in a bulk creation request.
type BulkTaskInput struct {
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Status      *string        `json:"status,omitempty"`
	Summary     *string        `json:"summary,omitempty"`
	Title       *string        `json:"title,omitempty"`
}

// BulkTask describes one task returned by the bulk creation endpoint.
type BulkTask struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// CreateBulkTasksInChatResponse contains bulk task creation results.
type CreateBulkTasksInChatResponse struct {
	Created int        `json:"created"`
	Tasks   []BulkTask `json:"tasks"`
}

// UpdateTaskInChatInput contains fields for updating a chat-scoped task.
type UpdateTaskInChatInput struct {
	Artifacts []TaskArtifactInput `json:"artifacts,omitempty"`
	Metadata  map[string]any      `json:"metadata,omitempty"`
	Status    *string             `json:"status,omitempty"`
	Summary   *string             `json:"summary,omitempty"`
	Title     *string             `json:"title,omitempty"`
}

// TaskArtifactInput contains one artifact reference for a task update.
type TaskArtifactInput struct {
	ArtifactID  string         `json:"artifact_id"`
	Description *string        `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	MimeType    *string        `json:"mime_type,omitempty"`
	Name        string         `json:"name"`
	URL         string         `json:"url"`
}

// ListMyUserTasksInput contains filters for ListMyUserTasks.
type ListMyUserTasksInput struct {
	Status  *string
	Page    *int
	PerPage *int
}

// ListMyUserTasksResponse contains user-scoped tasks and pagination metadata.
type ListMyUserTasksResponse struct {
	Data     []Task               `json:"data"`
	Metadata V2PaginationMetadata `json:"metadata"`
}

// CreateUserTaskInput contains fields for creating a user-scoped task.
type CreateUserTaskInput struct {
	AgentID     *string        `json:"agent_id,omitempty"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Summary     *string        `json:"summary,omitempty"`
}

// ListMyTasksInChat lists the current agent's tasks in a chat.
func (client *Client) ListMyTasksInChat(ctx context.Context, chatID string, input *ListMyTasksInChatInput) (*ListMyTasksInChatResponse, error) {
	if chatID == "" {
		return nil, errors.New("band: chat id is required")
	}
	values := url.Values{}
	if input != nil {
		addV2Pagination(values, V2PageInput{Page: input.Page, PerPage: input.PerPage})
		if input.Status != nil {
			values.Set("status", *input.Status)
		}
		if input.IncludeSubtasks != nil {
			values.Set("include_subtasks", strconv.FormatBool(*input.IncludeSubtasks))
		}
	}

	var out ListMyTasksInChatResponse
	path := "/api/v2/agents/me/chats/" + url.PathEscape(chatID) + "/tasks"
	if err := client.Do(ctx, http.MethodGet, appendQuery(path, values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateTaskInChat creates a task in a chat for the current agent.
func (client *Client) CreateTaskInChat(ctx context.Context, chatID string, input CreateTaskInChatInput) (*Task, error) {
	if chatID == "" {
		return nil, errors.New("band: chat id is required")
	}
	if input.Description == "" {
		return nil, errors.New("band: description is required")
	}

	var out struct {
		Data Task `json:"data"`
	}
	path := "/api/v2/agents/me/chats/" + url.PathEscape(chatID) + "/tasks"
	if err := client.Do(ctx, http.MethodPost, path, input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// CreateBulkTasksInChat creates multiple tasks in a chat for the current agent.
func (client *Client) CreateBulkTasksInChat(ctx context.Context, chatID string, input CreateBulkTasksInChatInput) (*CreateBulkTasksInChatResponse, error) {
	if chatID == "" {
		return nil, errors.New("band: chat id is required")
	}
	if len(input.Tasks) == 0 {
		return nil, errors.New("band: at least one task is required")
	}
	for _, task := range input.Tasks {
		if task.Description == "" {
			return nil, errors.New("band: task description is required")
		}
	}

	var out struct {
		Data CreateBulkTasksInChatResponse `json:"data"`
	}
	path := "/api/v2/agents/me/chats/" + url.PathEscape(chatID) + "/tasks/bulk"
	if err := client.Do(ctx, http.MethodPost, path, input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// GetTask fetches a chat-scoped task for the current agent.
func (client *Client) GetTask(ctx context.Context, chatID string, taskID string) (*Task, error) {
	if chatID == "" {
		return nil, errors.New("band: chat id is required")
	}
	if taskID == "" {
		return nil, errors.New("band: task id is required")
	}

	var out struct {
		Data Task `json:"data"`
	}
	path := taskInChatPath(chatID, taskID)
	if err := client.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// UpdateTaskInChat updates a chat-scoped task for the current agent.
func (client *Client) UpdateTaskInChat(ctx context.Context, chatID string, taskID string, input UpdateTaskInChatInput) (*Task, error) {
	if chatID == "" {
		return nil, errors.New("band: chat id is required")
	}
	if taskID == "" {
		return nil, errors.New("band: task id is required")
	}

	var out struct {
		Data Task `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPut, taskInChatPath(chatID, taskID), input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// DeleteTaskInChat deletes a chat-scoped task for the current agent.
func (client *Client) DeleteTaskInChat(ctx context.Context, chatID string, taskID string) error {
	if chatID == "" {
		return errors.New("band: chat id is required")
	}
	if taskID == "" {
		return errors.New("band: task id is required")
	}
	return client.Do(ctx, http.MethodDelete, taskInChatPath(chatID, taskID), nil, nil)
}

// ListMyUserTasks lists the current user's tasks.
func (client *Client) ListMyUserTasks(ctx context.Context, input *ListMyUserTasksInput) (*ListMyUserTasksResponse, error) {
	values := url.Values{}
	if input != nil {
		addV2Pagination(values, V2PageInput{Page: input.Page, PerPage: input.PerPage})
		if input.Status != nil {
			values.Set("status", *input.Status)
		}
	}

	var out ListMyUserTasksResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery("/api/v2/me/tasks", values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateUserTask creates a task for the current user.
func (client *Client) CreateUserTask(ctx context.Context, input CreateUserTaskInput) (*Task, error) {
	if input.Description == "" {
		return nil, errors.New("band: description is required")
	}

	var out struct {
		Data Task `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPost, "/api/v2/me/tasks", input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// GetUserTask fetches a task for the current user.
func (client *Client) GetUserTask(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, errors.New("band: task id is required")
	}

	var out struct {
		Data Task `json:"data"`
	}
	if err := client.Do(ctx, http.MethodGet, userTaskPath(taskID), nil, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// DeleteUserTask deletes a task for the current user.
func (client *Client) DeleteUserTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return errors.New("band: task id is required")
	}
	return client.Do(ctx, http.MethodDelete, userTaskPath(taskID), nil, nil)
}

func taskInChatPath(chatID string, taskID string) string {
	return "/api/v2/agents/me/chats/" + url.PathEscape(chatID) + "/tasks/" + url.PathEscape(taskID)
}

func userTaskPath(taskID string) string {
	return "/api/v2/me/tasks/" + url.PathEscape(taskID)
}
