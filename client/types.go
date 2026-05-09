package client

import (
	"net/url"
	"strconv"
	"time"
)

// PaginationMetadata describes paginated list response metadata.
type PaginationMetadata struct {
	Page       *int `json:"page,omitempty"`
	PageSize   *int `json:"page_size,omitempty"`
	TotalCount *int `json:"total_count,omitempty"`
	TotalPages *int `json:"total_pages,omitempty"`
}

// PageInput contains common pagination fields.
type PageInput struct {
	Page     *int
	PageSize *int
}

func addPagination(values url.Values, input PageInput) {
	if input.Page != nil {
		values.Set("page", strconv.Itoa(*input.Page))
	}
	if input.PageSize != nil {
		values.Set("page_size", strconv.Itoa(*input.PageSize))
	}
}

func appendQuery(path string, values url.Values) string {
	if len(values) == 0 {
		return path
	}
	return path + "?" + values.Encode()
}

// AgentIdentity describes the authenticated agent profile.
type AgentIdentity struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       *string   `json:"description"`
	Handle            string    `json:"handle"`
	InsertedAt        time.Time `json:"inserted_at"`
	ListedInDirectory *bool     `json:"listed_in_directory"`
	OwnerUUID         string    `json:"owner_uuid"`
	Tags              []string  `json:"tags,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Mention describes a user or agent mention in a chat message.
type Mention struct {
	ID     string `json:"id"`
	Handle string `json:"handle,omitempty"`
	Name   string `json:"name,omitempty"`
}

// ChatRoom describes a chat room returned by the agent chat API.
type ChatRoom struct {
	ID         string    `json:"id"`
	InsertedAt time.Time `json:"inserted_at"`
	TaskID     *string   `json:"task_id"`
	Title      *string   `json:"title"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ChatMessage describes a chat message returned by list/context APIs.
type ChatMessage struct {
	ID          string         `json:"id"`
	ChatRoomID  *string        `json:"chat_room_id"`
	Content     string         `json:"content"`
	MessageType string         `json:"message_type"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	SenderID    string         `json:"sender_id"`
	SenderName  *string        `json:"sender_name"`
	SenderType  string         `json:"sender_type"`
	InsertedAt  *time.Time     `json:"inserted_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
}

// MessageSentResponse is the minimal response after sending a chat message.
type MessageSentResponse struct {
	ID         string             `json:"id"`
	Recipients []MessageRecipient `json:"recipients"`
	Success    bool               `json:"success"`
}

// MessageRecipient describes a participant who will receive a message.
type MessageRecipient struct {
	Handle string  `json:"handle"`
	ID     string  `json:"id"`
	Name   *string `json:"name"`
}

// EventCreatedResponse is the minimal response after creating a chat event.
type EventCreatedResponse struct {
	ID          string `json:"id"`
	MessageType string `json:"message_type"`
	Success     bool   `json:"success"`
}

// MessageStatusResponse is the minimal response after changing message processing status.
type MessageStatusResponse struct {
	AttemptNumber int    `json:"attempt_number"`
	ID            string `json:"id"`
	Status        string `json:"status"`
	Success       bool   `json:"success"`
}

// ChatParticipant describes a chat room participant.
type ChatParticipant struct {
	Handle *string `json:"handle"`
	ID     string  `json:"id"`
	Name   *string `json:"name"`
	Role   string  `json:"role"`
	Status string  `json:"status"`
	Type   string  `json:"type"`
}

// AgentContact describes an agent contact relationship.
type AgentContact struct {
	Description       *string    `json:"description"`
	Handle            string     `json:"handle"`
	ID                string     `json:"id"`
	InsertedAt        *time.Time `json:"inserted_at"`
	IsExternal        *bool      `json:"is_external"`
	ListedInDirectory *bool      `json:"listed_in_directory"`
	Name              *string    `json:"name"`
	Tags              []string   `json:"tags,omitempty"`
	Type              string     `json:"type"`
}

// ContactRequest describes a sent or received contact request.
type ContactRequest struct {
	ID         string     `json:"id"`
	InsertedAt *time.Time `json:"inserted_at"`
	Message    *string    `json:"message"`
	Status     string     `json:"status"`
	FromHandle *string    `json:"from_handle,omitempty"`
	FromName   *string    `json:"from_name,omitempty"`
	ToHandle   *string    `json:"to_handle,omitempty"`
	ToName     *string    `json:"to_name,omitempty"`
}

// Peer describes a user or agent available for collaboration.
type Peer struct {
	Description       *string  `json:"description"`
	Handle            string   `json:"handle"`
	ID                string   `json:"id"`
	IsContact         bool     `json:"is_contact"`
	IsExternal        *bool    `json:"is_external"`
	ListedInDirectory *bool    `json:"listed_in_directory"`
	Name              string   `json:"name"`
	Source            string   `json:"source"`
	Tags              []string `json:"tags,omitempty"`
	Type              string   `json:"type"`
}

// Memory describes a memory entry stored by an agent.
type Memory struct {
	Content        string         `json:"content"`
	ID             string         `json:"id"`
	InsertedAt     time.Time      `json:"inserted_at"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	OrganizationID *string        `json:"organization_id"`
	Scope          string         `json:"scope"`
	Segment        string         `json:"segment"`
	SourceAgentID  *string        `json:"source_agent_id"`
	Status         *string        `json:"status"`
	SubjectID      *string        `json:"subject_id"`
	System         string         `json:"system"`
	Thought        *string        `json:"thought"`
	Type           string         `json:"type"`
}
