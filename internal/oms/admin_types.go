package oms

import "time"

// AdminOutageEvent is the admin-facing view of an outage event, joined
// with its human-readable status label from oms_status.
type AdminOutageEvent struct {
	EventID            string       `json:"eventId"`
	CaNumber           string       `json:"caNumber"`
	Level              EventLevel   `json:"level"`
	Status             OutageStatus `json:"status"`
	StatusLabel        string       `json:"statusLabel"`
	Message            string       `json:"message"`
	StartedAt          time.Time    `json:"startedAt"`
	EstimatedRestoreAt *time.Time   `json:"estimatedRestoreAt"`
}

// UpdateOutageEventRequest is a partial update — nil fields are left
// unchanged. Setting Status to a closed code (e.g. RESTORED) closes the
// complaint/work order.
type UpdateOutageEventRequest struct {
	Status             *OutageStatus `json:"status"`
	Message            *string       `json:"message"`
	EstimatedRestoreAt *time.Time    `json:"estimatedRestoreAt"`
}

// StatusOption is a row from oms_status — the valid set an admin can move
// an event into, managed as data instead of a hardcoded CHECK constraint.
type StatusOption struct {
	Code     OutageStatus `json:"code"`
	Label    string       `json:"label"`
	IsClosed bool         `json:"isClosed"`
}
