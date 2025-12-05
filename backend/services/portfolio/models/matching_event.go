package models

type MatchingEvent struct {
	Event      string
	TraceID    string
	UserId     string
	JWT        string
	Order      Order
	Orders     []*ClaimedCandidate
	Executions []*ExecutionRecord
	Error      string
}
