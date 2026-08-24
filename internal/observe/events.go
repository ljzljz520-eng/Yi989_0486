package observe

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

type Event struct {
	Time   time.Time      `json:"time"`
	Name   string         `json:"name"`
	Actor  string         `json:"actor,omitempty"`
	Entity string         `json:"entity,omitempty"`
	Fields map[string]any `json:"fields,omitempty"`
}

type Logger struct {
	Writer io.Writer
	Mu     sync.Mutex
}

func (l *Logger) Emit(event Event) error {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	if event.Time.IsZero() {
		event.Time = time.Now()
	}
	return json.NewEncoder(l.Writer).Encode(event)
}

func NewEvent(name, actor, entity string) Event {
	return Event{Time: time.Now(), Name: name, Actor: actor, Entity: entity, Fields: map[string]any{}}
}
