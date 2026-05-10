package core

import "time"

// Metadata mirrors the open-ended SDK metadata map.
type Metadata map[string]any

type ToolOperationResult struct {
	OK     *bool    `json:"ok,omitempty"`
	Status *string  `json:"status,omitempty"`
	Extra  Metadata `json:"-"`
}

type MentionReference struct {
	ID       string `json:"id"`
	Handle   string `json:"handle,omitempty"`
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
}

// MentionInput represents either plain handles or structured references.
// Runtime serialization should choose one populated arm.
type MentionInput struct {
	Handles    []string
	References []MentionReference
}

type PaginationMetadataLike struct {
	Page       *int     `json:"page,omitempty"`
	PageSize   *int     `json:"pageSize,omitempty"`
	TotalPages *int     `json:"totalPages,omitempty"`
	TotalCount *int     `json:"totalCount,omitempty"`
	Extra      Metadata `json:"-"`
}

type PaginatedList[T any] struct {
	Data     []T                     `json:"data"`
	Metadata *PaginationMetadataLike `json:"metadata,omitempty"`
}

type ParticipantRecord struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Handle *string `json:"handle,omitempty"`
	Role   *string `json:"role,omitempty"`
}

type PeerRecord struct {
	ID          *string `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Type        *string `json:"type,omitempty"`
	Handle      *string `json:"handle,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ContactRecord struct {
	ID          *string `json:"id,omitempty"`
	Handle      *string `json:"handle,omitempty"`
	Name        *string `json:"name,omitempty"`
	Type        *string `json:"type,omitempty"`
	Description *string `json:"description,omitempty"`
	IsExternal  *bool   `json:"is_external,omitempty"`
	InsertedAt  *string `json:"inserted_at,omitempty"`
}

type ContactRequestAction string

const (
	ContactRequestActionApprove ContactRequestAction = "approve"
	ContactRequestActionReject  ContactRequestAction = "reject"
	ContactRequestActionCancel  ContactRequestAction = "cancel"
)

type ContactRequestsResult struct {
	Received []ContactRecord `json:"received"`
	Sent     []ContactRecord `json:"sent"`
	Metadata Metadata        `json:"metadata,omitempty"`
}

type ListContactsArgs struct {
	Page     *int
	PageSize *int
}

type AddContactArgs struct {
	Handle  string
	Message *string
}

type RemoveContactArgs struct {
	Target    string
	Handle    string
	ContactID string
}

type ListContactRequestsArgs struct {
	Page       *int
	PageSize   *int
	SentStatus *string
}

type RespondContactRequestArgs struct {
	Action    ContactRequestAction
	Target    string
	Handle    string
	RequestID string
}

type MemorySystem string

const (
	MemorySystemSensory  MemorySystem = "sensory"
	MemorySystemWorking  MemorySystem = "working"
	MemorySystemLongTerm MemorySystem = "long_term"
)

type MemoryType string

const (
	MemoryTypeIconic     MemoryType = "iconic"
	MemoryTypeEchoic     MemoryType = "echoic"
	MemoryTypeHaptic     MemoryType = "haptic"
	MemoryTypeEpisodic   MemoryType = "episodic"
	MemoryTypeSemantic   MemoryType = "semantic"
	MemoryTypeProcedural MemoryType = "procedural"
)

type MemorySegment string

const (
	MemorySegmentUser      MemorySegment = "user"
	MemorySegmentAgent     MemorySegment = "agent"
	MemorySegmentTool      MemorySegment = "tool"
	MemorySegmentGuideline MemorySegment = "guideline"
)

type MemoryStatus string

const (
	MemoryStatusActive     MemoryStatus = "active"
	MemoryStatusSuperseded MemoryStatus = "superseded"
	MemoryStatusArchived   MemoryStatus = "archived"
	MemoryStatusAll        MemoryStatus = "all"
)

type MemoryVisibility string

const (
	MemoryVisibilitySubject      MemoryVisibility = "subject"
	MemoryVisibilityOrganization MemoryVisibility = "organization"
)

type MemoryScope string

const (
	MemoryScopeSubject      MemoryScope = "subject"
	MemoryScopeOrganization MemoryScope = "organization"
	MemoryScopeAll          MemoryScope = "all"
)

type ListMemoriesArgs struct {
	SubjectID    *string        `json:"subject_id,omitempty"`
	Scope        *MemoryScope   `json:"scope,omitempty"`
	System       *MemorySystem  `json:"system,omitempty"`
	Type         *MemoryType    `json:"type,omitempty"`
	Segment      *MemorySegment `json:"segment,omitempty"`
	ContentQuery *string        `json:"content_query,omitempty"`
	PageSize     *int           `json:"page_size,omitempty"`
	Status       *MemoryStatus  `json:"status,omitempty"`
}

type StoreMemoryArgs struct {
	Content   string            `json:"content"`
	System    MemorySystem      `json:"system"`
	Type      MemoryType        `json:"type"`
	Segment   MemorySegment     `json:"segment"`
	Thought   string            `json:"thought"`
	Scope     *MemoryVisibility `json:"scope,omitempty"`
	SubjectID *string           `json:"subject_id,omitempty"`
	Metadata  Metadata          `json:"metadata,omitempty"`
}

type MemoryRecord struct {
	ID             *string  `json:"id,omitempty"`
	Content        *string  `json:"content,omitempty"`
	System         *string  `json:"system,omitempty"`
	Type           *string  `json:"type,omitempty"`
	Segment        *string  `json:"segment,omitempty"`
	Thought        *string  `json:"thought,omitempty"`
	SubjectID      *string  `json:"subject_id,omitempty"`
	SourceAgentID  *string  `json:"source_agent_id,omitempty"`
	OrganizationID *string  `json:"organization_id,omitempty"`
	Scope          *string  `json:"scope,omitempty"`
	Status         *string  `json:"status,omitempty"`
	Metadata       Metadata `json:"metadata,omitempty"`
	InsertedAt     *string  `json:"inserted_at,omitempty"`
}

type ToolSchemaRecord map[string]any

type PlatformMessageLike struct {
	ID          string    `json:"id"`
	RoomID      string    `json:"roomId"`
	Content     string    `json:"content"`
	SenderID    string    `json:"senderId"`
	SenderType  string    `json:"senderType"`
	SenderName  *string   `json:"senderName"`
	MessageType string    `json:"messageType"`
	Metadata    Metadata  `json:"metadata"`
	CreatedAt   time.Time `json:"createdAt"`
}
