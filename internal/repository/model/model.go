package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// An Log defines possible fields contained with
// an event log.
type Log struct {
	ID        primitive.ObjectID `bson:"_id"`
	Timestamp time.Time          `json:"created_at"`
	Action    string             `json:"action"`
	Actor     Actor              `json:"actor"`
	Entity    Entity             `json:"entity"`
	Context   Context            `json:"context"`
	Extension map[string]any     `json:"extension,omitempty"`
}

// An Actor defines the user or service responsible for
// the event.
type Actor struct {
	Type      string         `json:"type"`
	ID        string         `json:"id"`
	Extension map[string]any `json:"extension,omitempty"`
}

// An Entity defines the resource that was impacted.
type Entity struct {
	Type      string         `json:"type"`
	Extension map[string]any `json:"extension,omitempty"`
}

// A Context describes the source from where the actor
// or service originated.
type Context struct {
	IPAddr    string         `json:"ip_address"`
	Location  string         `json:"location"`
	Extension map[string]any `json:"extension,omitempty"`
}

// A ServiceID defines the service name type.
type ServiceID string

var (
	AnonymousService = ServiceID("")
)

// IsAnonymous checks if a ServiceID instance is the AnonymousService.
func (u *ServiceID) IsAnonymous() bool {
	return *u == AnonymousService
}

// A Token describes authentication token (access token).
type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	ServiceID ServiceID `json:"-"`
}
