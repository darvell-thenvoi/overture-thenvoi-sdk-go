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

// V2PaginationMetadata describes API v2 paginated list response metadata.
type V2PaginationMetadata struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// Pagination describes API v2 pagination fields returned as pagination.
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	TotalItems int `json:"total_items"`
}

// PageInput contains common pagination fields.
type PageInput struct {
	Page     *int
	PageSize *int
}

// V2PageInput contains common API v2 pagination fields.
type V2PageInput struct {
	Page    *int
	PerPage *int
}

func addPagination(values url.Values, input PageInput) {
	if input.Page != nil {
		values.Set("page", strconv.Itoa(*input.Page))
	}
	if input.PageSize != nil {
		values.Set("page_size", strconv.Itoa(*input.PageSize))
	}
}

func addV2Pagination(values url.Values, input V2PageInput) {
	if input.Page != nil {
		values.Set("page", strconv.Itoa(*input.Page))
	}
	if input.PerPage != nil {
		values.Set("per_page", strconv.Itoa(*input.PerPage))
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

// Agent describes an API v2 agent.
type Agent struct {
	ID                     string         `json:"id"`
	Name                   string         `json:"name"`
	Description            string         `json:"description"`
	ModelType              string         `json:"model_type"`
	IsExternal             bool           `json:"is_external"`
	OwnerUUID              string         `json:"owner_uuid"`
	OrganizationID         *string        `json:"organization_id"`
	SystemPromptID         *string        `json:"system_prompt_id"`
	StructuredOutputSchema map[string]any `json:"structured_output_schema"`
	InsertedAt             time.Time      `json:"inserted_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
}

// AgentListItem describes an API v2 agent list row.
type AgentListItem struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ModelType      string    `json:"model_type"`
	IsExternal     bool      `json:"is_external"`
	OwnerUUID      string    `json:"owner_uuid"`
	SystemPromptID *string   `json:"system_prompt_id"`
	OrganizationID *string   `json:"organization_id"`
	InsertedAt     time.Time `json:"inserted_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Mention describes a user or agent mention in a chat message.
type Mention struct {
	ID     string `json:"id"`
	Handle string `json:"handle,omitempty"`
	Name   string `json:"name,omitempty"`
}

// ChatRoom describes a chat room returned by the agent chat API.
type ChatRoom struct {
	ID         string         `json:"id"`
	DeletedAt  *time.Time     `json:"deleted_at"`
	InsertedAt time.Time      `json:"inserted_at"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Status     *string        `json:"status"`
	TaskID     *string        `json:"task_id"`
	Title      *string        `json:"title"`
	Type       *string        `json:"type"`
	UpdatedAt  time.Time      `json:"updated_at"`
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

// ExecutionResponse describes the result of starting an API v2 agent execution.
type ExecutionResponse struct {
	ExecutionID *string        `json:"execution_id"`
	TaskID      string         `json:"task_id"`
	ChatRoomID  string         `json:"chat_room_id"`
	Status      string         `json:"status"`
	Agent       ExecutionAgent `json:"agent"`
	Request     string         `json:"request"`
	CreatedAt   time.Time      `json:"created_at"`
	Links       ExecutionLinks `json:"links"`
}

// ExecutionAgent describes the executed agent.
type ExecutionAgent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ExecutionLinks contains API links for an execution.
type ExecutionLinks struct {
	ChatRoom string `json:"chat_room"`
	Messages string `json:"messages"`
	Task     string `json:"task"`
}

// Contact describes an API v2 contact.
type Contact struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Type               string              `json:"type"`
	Description        *string             `json:"description"`
	Email              *string             `json:"email"`
	AvatarURL          *string             `json:"avatar_url"`
	CanAddToChats      bool                `json:"can_add_to_chats"`
	CanExecuteRequests *bool               `json:"can_execute_requests"`
	Status             ContactStatus       `json:"status"`
	Workload           map[string]any      `json:"workload"`
	Capabilities       []string            `json:"capabilities"`
	Metadata           map[string]any      `json:"metadata"`
	Relationship       ContactRelationship `json:"relationship"`
}

// ContactStatus describes contact connection and availability state.
type ContactStatus struct {
	Connection   string  `json:"connection"`
	LastSeen     *string `json:"last_seen,omitempty"`
	Availability *string `json:"availability,omitempty"`
}

// ContactRelationship describes a contact relationship.
type ContactRelationship struct {
	AddedAt          *time.Time          `json:"added_at,omitempty"`
	InteractionCount int                 `json:"interaction_count,omitempty"`
	Permissions      *ContactPermissions `json:"permissions,omitempty"`
}

// ContactPermissions describes allowed actions for a detailed contact.
type ContactPermissions struct {
	CanAddToChats      bool `json:"can_add_to_chats"`
	CanRemoveFromChats bool `json:"can_remove_from_chats"`
	CanExecuteRequests bool `json:"can_execute_requests"`
	CanViewPerformance bool `json:"can_view_performance"`
}

// ContactDetail describes an API v2 contact detail response.
type ContactDetail struct {
	Contact
	CreatedAt        time.Time      `json:"created_at"`
	PerformanceStats map[string]any `json:"performance_stats"`
}

// ContactAvailability describes a contact availability result.
type ContactAvailability struct {
	ID                string        `json:"id"`
	Available         bool          `json:"available"`
	Status            ContactStatus `json:"status"`
	EstimatedWaitTime *string       `json:"estimated_wait_time"`
	Confidence        float64       `json:"confidence"`
	UnavailableReason *string       `json:"unavailable_reason"`
}

// Model describes an API v2 model.
type Model struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
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
