package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

const taskResponseBody = `{"data":{"id":"task_1","title":"Ship SDK","status":"in_progress","description":"Add task methods","summary":"Expose v2 task operations","agent_id":"agent_1","chat_room_id":"chat_1","parent_task_id":"parent_1","execution_id":"execution_1","participants":["agent_1","user_1"],"metadata":{"priority":"high"},"user_id":"user_1","inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z"}}`

func TestListMyTasksInChat(t *testing.T) {
	t.Parallel()
	page := 2
	perPage := 10
	status := "in_progress"
	includeSubtasks := true
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("method = %s", req.Method)
		}
		if req.URL.String() != "https://api.test/api/v2/agents/me/chats/chat%2F1/tasks?include_subtasks=true&page=2&per_page=10&status=in_progress" {
			t.Fatalf("url = %s", req.URL.String())
		}
		return jsonResponse(http.StatusOK, `{"data":[{"id":"task_1","title":"Ship SDK","status":"in_progress","description":"Add task methods","user_id":"user_1","inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z"}],"pagination":{"page":2,"per_page":10,"total_pages":3,"total_items":25}}`), nil
	})
	sdk := newTaskTestClient(transport)

	out, err := sdk.ListMyTasksInChat(context.Background(), "chat/1", &client.ListMyTasksInChatInput{
		Status:          &status,
		Page:            &page,
		PerPage:         &perPage,
		IncludeSubtasks: &includeSubtasks,
	})
	if err != nil {
		t.Fatalf("ListMyTasksInChat returned error: %v", err)
	}
	if len(out.Data) != 1 || out.Data[0].ID != "task_1" {
		t.Fatalf("data = %#v", out.Data)
	}
	if out.Pagination.TotalItems != 25 {
		t.Fatalf("pagination = %#v", out.Pagination)
	}
}

func TestCreateTaskInChat(t *testing.T) {
	t.Parallel()
	title := "SDK task"
	summary := "Add coverage"
	parentTaskID := "parent_1"
	status := "new"
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assertMethodAndURL(t, req, http.MethodPost, "https://api.test/api/v2/agents/me/chats/chat_1/tasks")
		var body map[string]any
		decodeRequestBody(t, req, &body)
		if body["description"] != "Add task methods" || body["title"] != title || body["summary"] != summary || body["parent_task_id"] != parentTaskID || body["status"] != status {
			t.Fatalf("body = %#v", body)
		}
		metadata, ok := body["metadata"].(map[string]any)
		if !ok || metadata["source"] != "test" {
			t.Fatalf("metadata = %#v", body["metadata"])
		}
		return jsonResponse(http.StatusCreated, taskResponseBody), nil
	})
	sdk := newTaskTestClient(transport)

	out, err := sdk.CreateTaskInChat(context.Background(), "chat_1", client.CreateTaskInChatInput{
		Description:  "Add task methods",
		Title:        &title,
		Summary:      &summary,
		ParentTaskID: &parentTaskID,
		Metadata:     map[string]any{"source": "test"},
		Status:       &status,
	})
	assertDecodedTask(t, out, err)
}

func TestCreateBulkTasksInChat(t *testing.T) {
	t.Parallel()
	status := "waiting"
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assertMethodAndURL(t, req, http.MethodPost, "https://api.test/api/v2/agents/me/chats/chat_1/tasks/bulk")
		var body map[string][]map[string]any
		decodeRequestBody(t, req, &body)
		if len(body["tasks"]) != 1 || body["tasks"][0]["description"] != "First task" || body["tasks"][0]["status"] != status {
			t.Fatalf("body = %#v", body)
		}
		return jsonResponse(http.StatusCreated, `{"data":{"created":1,"tasks":[{"id":"task_1","description":"First task","status":"waiting"}]}}`), nil
	})
	sdk := newTaskTestClient(transport)

	out, err := sdk.CreateBulkTasksInChat(context.Background(), "chat_1", client.CreateBulkTasksInChatInput{
		Tasks: []client.BulkTaskInput{{Description: "First task", Status: &status}},
	})
	if err != nil {
		t.Fatalf("CreateBulkTasksInChat returned error: %v", err)
	}
	if out.Created != 1 || len(out.Tasks) != 1 || out.Tasks[0].ID != "task_1" {
		t.Fatalf("out = %#v", out)
	}
}

func TestGetTask(t *testing.T) {
	t.Parallel()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assertMethodAndURL(t, req, http.MethodGet, "https://api.test/api/v2/agents/me/chats/chat%2F1/tasks/task%2F1")
		return jsonResponse(http.StatusOK, taskResponseBody), nil
	})
	sdk := newTaskTestClient(transport)

	out, err := sdk.GetTask(context.Background(), "chat/1", "task/1")
	assertDecodedTask(t, out, err)
}

func TestUpdateTaskInChat(t *testing.T) {
	t.Parallel()
	status := "completed"
	title := "Done"
	artifactDescription := "Build log"
	mimeType := "text/plain"
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assertMethodAndURL(t, req, http.MethodPut, "https://api.test/api/v2/agents/me/chats/chat_1/tasks/task_1")
		var body map[string]any
		decodeRequestBody(t, req, &body)
		if body["status"] != status || body["title"] != title {
			t.Fatalf("body = %#v", body)
		}
		artifacts, ok := body["artifacts"].([]any)
		if !ok || len(artifacts) != 1 {
			t.Fatalf("artifacts = %#v", body["artifacts"])
		}
		artifact, ok := artifacts[0].(map[string]any)
		if !ok || artifact["artifact_id"] != "artifact_1" || artifact["name"] != "log.txt" || artifact["url"] != "https://files.test/log.txt" || artifact["mime_type"] != mimeType || artifact["description"] != artifactDescription {
			t.Fatalf("artifact = %#v", artifacts[0])
		}
		return jsonResponse(http.StatusOK, taskResponseBody), nil
	})
	sdk := newTaskTestClient(transport)

	out, err := sdk.UpdateTaskInChat(context.Background(), "chat_1", "task_1", client.UpdateTaskInChatInput{
		Status: &status,
		Title:  &title,
		Artifacts: []client.TaskArtifactInput{{
			ArtifactID:  "artifact_1",
			Description: &artifactDescription,
			MimeType:    &mimeType,
			Name:        "log.txt",
			URL:         "https://files.test/log.txt",
		}},
	})
	assertDecodedTask(t, out, err)
}

func TestDeleteTaskInChat(t *testing.T) {
	t.Parallel()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assertMethodAndURL(t, req, http.MethodDelete, "https://api.test/api/v2/agents/me/chats/chat_1/tasks/task_1")
		return jsonResponse(http.StatusNoContent, ""), nil
	})
	sdk := newTaskTestClient(transport)

	if err := sdk.DeleteTaskInChat(context.Background(), "chat_1", "task_1"); err != nil {
		t.Fatalf("DeleteTaskInChat returned error: %v", err)
	}
}

func TestListMyUserTasks(t *testing.T) {
	t.Parallel()
	page := 3
	perPage := 20
	status := "reviewing"
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assertMethodAndURL(t, req, http.MethodGet, "https://api.test/api/v2/me/tasks?page=3&per_page=20&status=reviewing")
		return jsonResponse(http.StatusOK, `{"data":[{"id":"task_1","title":"Ship SDK","status":"reviewing","description":"Add task methods","user_id":"user_1","inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z"}],"metadata":{"page":3,"per_page":20,"total_count":45,"total_pages":3}}`), nil
	})
	sdk := newTaskTestClient(transport)

	out, err := sdk.ListMyUserTasks(context.Background(), &client.ListMyUserTasksInput{Status: &status, Page: &page, PerPage: &perPage})
	if err != nil {
		t.Fatalf("ListMyUserTasks returned error: %v", err)
	}
	if len(out.Data) != 1 || out.Metadata.TotalCount != 45 {
		t.Fatalf("out = %#v", out)
	}
}

func TestCreateUserTask(t *testing.T) {
	t.Parallel()
	agentID := "agent_1"
	summary := "User task"
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assertMethodAndURL(t, req, http.MethodPost, "https://api.test/api/v2/me/tasks")
		var body map[string]any
		decodeRequestBody(t, req, &body)
		if body["description"] != "Ask agent" || body["agent_id"] != agentID || body["summary"] != summary {
			t.Fatalf("body = %#v", body)
		}
		return jsonResponse(http.StatusCreated, taskResponseBody), nil
	})
	sdk := newTaskTestClient(transport)

	out, err := sdk.CreateUserTask(context.Background(), client.CreateUserTaskInput{
		AgentID:     &agentID,
		Description: "Ask agent",
		Summary:     &summary,
	})
	assertDecodedTask(t, out, err)
}

func TestGetUserTask(t *testing.T) {
	t.Parallel()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assertMethodAndURL(t, req, http.MethodGet, "https://api.test/api/v2/me/tasks/task%2F1")
		return jsonResponse(http.StatusOK, taskResponseBody), nil
	})
	sdk := newTaskTestClient(transport)

	out, err := sdk.GetUserTask(context.Background(), "task/1")
	assertDecodedTask(t, out, err)
}

func TestDeleteUserTask(t *testing.T) {
	t.Parallel()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assertMethodAndURL(t, req, http.MethodDelete, "https://api.test/api/v2/me/tasks/task_1")
		return jsonResponse(http.StatusNoContent, ""), nil
	})
	sdk := newTaskTestClient(transport)

	if err := sdk.DeleteUserTask(context.Background(), "task_1"); err != nil {
		t.Fatalf("DeleteUserTask returned error: %v", err)
	}
}

func TestTaskValidationErrorsDoNotCallTransport(t *testing.T) {
	t.Parallel()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
		return nil, errors.New("unexpected request")
	})
	sdk := newTaskTestClient(transport)
	ctx := context.Background()

	assertError(t, func() error {
		_, err := sdk.ListMyTasksInChat(ctx, "", nil)
		return err
	}, "band: chat id is required")
	assertError(t, func() error {
		_, err := sdk.CreateTaskInChat(ctx, "chat_1", client.CreateTaskInChatInput{})
		return err
	}, "band: description is required")
	assertError(t, func() error {
		_, err := sdk.CreateBulkTasksInChat(ctx, "chat_1", client.CreateBulkTasksInChatInput{})
		return err
	}, "band: at least one task is required")
	assertError(t, func() error {
		_, err := sdk.CreateBulkTasksInChat(ctx, "chat_1", client.CreateBulkTasksInChatInput{Tasks: []client.BulkTaskInput{{}}})
		return err
	}, "band: task description is required")
	assertError(t, func() error {
		_, err := sdk.GetTask(ctx, "", "task_1")
		return err
	}, "band: chat id is required")
	assertError(t, func() error {
		_, err := sdk.GetTask(ctx, "chat_1", "")
		return err
	}, "band: task id is required")
	assertError(t, func() error {
		_, err := sdk.UpdateTaskInChat(ctx, "", "task_1", client.UpdateTaskInChatInput{})
		return err
	}, "band: chat id is required")
	assertError(t, func() error {
		_, err := sdk.UpdateTaskInChat(ctx, "chat_1", "", client.UpdateTaskInChatInput{})
		return err
	}, "band: task id is required")
	assertError(t, func() error {
		return sdk.DeleteTaskInChat(ctx, "", "task_1")
	}, "band: chat id is required")
	assertError(t, func() error {
		return sdk.DeleteTaskInChat(ctx, "chat_1", "")
	}, "band: task id is required")
	assertError(t, func() error {
		_, err := sdk.CreateUserTask(ctx, client.CreateUserTaskInput{})
		return err
	}, "band: description is required")
	assertError(t, func() error {
		_, err := sdk.GetUserTask(ctx, "")
		return err
	}, "band: task id is required")
	assertError(t, func() error {
		return sdk.DeleteUserTask(ctx, "")
	}, "band: task id is required")
}

func newTaskTestClient(transport roundTripFunc) *client.Client {
	return client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
}

func assertMethodAndURL(t *testing.T, req *http.Request, method string, requestURL string) {
	t.Helper()
	if req.Method != method {
		t.Fatalf("method = %s", req.Method)
	}
	if req.URL.String() != requestURL {
		t.Fatalf("url = %s", req.URL.String())
	}
}

func decodeRequestBody(t *testing.T, req *http.Request, out any) {
	t.Helper()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if err := json.Unmarshal(body, out); err != nil {
		t.Fatalf("decode body %q: %v", string(body), err)
	}
}

func assertDecodedTask(t *testing.T, out *client.Task, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("task method returned error: %v", err)
	}
	if out == nil || out.ID != "task_1" || out.Description != "Add task methods" || out.Title != "Ship SDK" || out.Status != "in_progress" {
		t.Fatalf("task = %#v", out)
	}
	wantInserted, err := time.Parse(time.RFC3339, "2026-01-02T03:04:05Z")
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	if !out.InsertedAt.Equal(wantInserted) {
		t.Fatalf("inserted_at = %s", out.InsertedAt)
	}
	if out.Summary == nil || *out.Summary != "Expose v2 task operations" || out.AgentID == nil || *out.AgentID != "agent_1" || out.ChatRoomID == nil || *out.ChatRoomID != "chat_1" {
		t.Fatalf("task nullable fields = %#v", out)
	}
	if len(out.Participants) != 2 || out.Metadata["priority"] != "high" {
		t.Fatalf("task participants/metadata = %#v %#v", out.Participants, out.Metadata)
	}
}

func assertError(t *testing.T, call func() error, want string) {
	t.Helper()
	err := call()
	if err == nil || err.Error() != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
}
