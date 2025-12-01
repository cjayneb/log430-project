package models

import "encoding/json"

type OutboxRecord struct {
	ID            int64
	Topic         string
	EventType     string
	TraceID       string
	UserID        string
	JWT           string
	PayloadString string
	RetryCount    int
}

func (r *OutboxRecord) UnmarshalPayload(v any) error {
	return json.Unmarshal([]byte(r.PayloadString), v)
}
