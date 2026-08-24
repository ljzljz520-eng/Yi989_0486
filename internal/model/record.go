package model

import "time"

type RecordState string

const (
	RecordReserved  RecordState = "reserved"
	RecordConfirmed RecordState = "confirmed"
	RecordReclaimed RecordState = "reclaimed"
	RecordCancelled RecordState = "cancelled"
)

type Record struct {
	ID          string      `json:"id"`
	BatchID     string      `json:"batch_id"`
	CustomerID  string      `json:"customer_id"`
	State       RecordState `json:"state"`
	AmountCents int64       `json:"amount_cents"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	ConfirmedAt *time.Time  `json:"confirmed_at,omitempty"`
	ReclaimedAt *time.Time  `json:"reclaimed_at,omitempty"`
	Reason      string      `json:"reason,omitempty"`
	Version     int         `json:"version"`
}

type Audit struct {
	ID        string    `json:"id"`
	Entity    string    `json:"entity"`
	EntityID  string    `json:"entity_id"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

type Profile struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	Roles       []string  `json:"roles"`
	Nodes       []string  `json:"nodes"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Claim struct {
	RecordID   string      `json:"record_id"`
	BatchID    string      `json:"batch_id"`
	CustomerID string      `json:"customer_id"`
	State      RecordState `json:"state"`
}
