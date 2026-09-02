// Package oms implements the OMS outage integration group
// (spec/oms.openapi.yaml), mounted under /api/v1/oms/...
package oms

import "time"

type EventLevel string

const (
	EventLevelMeter       EventLevel = "METER"
	EventLevelTransformer EventLevel = "TRANSFORMER"
	EventLevelFeeder      EventLevel = "FEEDER"
)

type OutageStatus string

const (
	StatusReceived     OutageStatus = "RECEIVED"
	StatusAcknowledged OutageStatus = "ACKNOWLEDGED"
	StatusInProgress   OutageStatus = "IN_PROGRESS"
	StatusRestored     OutageStatus = "RESTORED"
)

type RecommendedAction string

const (
	ActionInformExisting RecommendedAction = "INFORM_EXISTING_EVENT"
	ActionCreateMeter    RecommendedAction = "CREATE_METER_EVENT"
)

type NetworkReference struct {
	MeterID       string `json:"meterId"`
	TransformerID string `json:"transformerId"`
	FeederID      string `json:"feederId"`
}

// GeoPoint is a best-effort coordinate for the FE-osm map client — see
// GisClient.Lookup. Nil fields mean no match was found (not sent as 0,0).
type GeoPoint struct {
	Lat     *float64 `json:"lat"`
	Lon     *float64 `json:"lon"`
	GisType *string  `json:"gisType"` // "POINT" (exact meter match) or "AREA" (approximated)
}

type ActiveOutageEvent struct {
	EventID            string       `json:"eventId"`
	Level              EventLevel   `json:"level"`
	Status             OutageStatus `json:"status"`
	Message            string       `json:"message"`
	StartedAt          time.Time    `json:"startedAt"`
	EstimatedRestoreAt *time.Time   `json:"estimatedRestoreAt"`
	Location           *GeoPoint    `json:"location"`
}

type OutageCheckResponse struct {
	CaNumber          string             `json:"caNumber"`
	CustomerFound     bool               `json:"customerFound"`
	Network           NetworkReference   `json:"network"`
	ActiveEvent       *ActiveOutageEvent `json:"activeEvent"`
	RecommendedAction RecommendedAction  `json:"recommendedAction"`
}

type CreateOutageRequest struct {
	CaNumber     string  `json:"caNumber"`
	Description  string  `json:"description"`
	ContactPhone *string `json:"contactPhone"`
	LocationNote *string `json:"locationNote"`
}

type CreateOutageResponse struct {
	EventID  string       `json:"eventId"`
	CaNumber string       `json:"caNumber"`
	Level    EventLevel   `json:"level"`
	Status   OutageStatus `json:"status"`
	Message  string       `json:"message"`
	Location *GeoPoint    `json:"location"`
}

type CreateAnonymousOutageRequest struct {
	Description  string `json:"description"`
	Location     string `json:"location"`
	ContactPhone string `json:"contactPhone"`
	// Optional browser geolocation fallback — no CA means no MST GIS lookup
	// is possible, so the client can supply coordinates directly instead.
	// When present these are used as-is and skip enrichAnonymousLocation.
	Lat *float64 `json:"lat"`
	Lon *float64 `json:"lon"`
}

type CreateAnonymousOutageResponse struct {
	ReportID string       `json:"reportId"`
	Status   OutageStatus `json:"status"`
	Message  string       `json:"message"`
	Location *GeoPoint    `json:"location"`
}

type ErrorCode string

const (
	ErrInvalidCA        ErrorCode = "INVALID_CA"
	ErrCANotFound       ErrorCode = "CA_NOT_FOUND"
	ErrEventNotFound    ErrorCode = "EVENT_NOT_FOUND"
	ErrReportNotFound   ErrorCode = "REPORT_NOT_FOUND"
	ErrInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrActiveEventExist ErrorCode = "ACTIVE_EVENT_EXISTS"
	ErrOmsUnavailable   ErrorCode = "OMS_UNAVAILABLE"
	ErrOmsTimeout       ErrorCode = "OMS_TIMEOUT"
	ErrInvalidResponse  ErrorCode = "INVALID_RESPONSE"
	ErrInternal         ErrorCode = "INTERNAL_ERROR"
)

// ApiError carries the HTTP status alongside the spec's error code/message
// so handlers can translate a Client error straight into a response.
type ApiError struct {
	Status          int
	Code            ErrorCode
	Message         string
	ExistingEventID string // set only for 409 ACTIVE_EVENT_EXISTS
}

func (e *ApiError) Error() string { return e.Message }
