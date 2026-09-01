package voc

import "time"

// AdminCase is the admin-facing view of a case — unlike VocCaseDetail
// (public lookup) it includes reporter/consent, which the public spec
// deliberately omits from case-tracking responses (PII).
type AdminCase struct {
	CaseID         string                  `json:"caseId"`
	VocNumber      string                  `json:"vocNumber"`
	Status         CaseStatus              `json:"status"`
	StatusLabel    string                  `json:"statusLabel"`
	JourneyCode    JourneyCode             `json:"journeyCode"`
	Classification ClassificationSelection `json:"classification"`
	Incident       IncidentLocation        `json:"incident"`
	Reporter       *Reporter               `json:"reporter"`
	FrequencyCode  *string                 `json:"frequencyCode"`
	SeverityLevel  *int                    `json:"severityLevel"`
	Detail         string                  `json:"detail"`
	Consent        ConsentRecord           `json:"consent"`
	CreatedAt      time.Time               `json:"createdAt"`
	UpdatedAt      time.Time               `json:"updatedAt"`
	Timeline       []TimelineEntry         `json:"timeline"`
}

// UpdateCaseRequest is a partial update — nil fields are left unchanged.
// Setting Status appends a new timeline entry (label auto-resolved from
// voc_status; Message defaults to the status label if omitted).
type UpdateCaseRequest struct {
	Status  *CaseStatus `json:"status"`
	Message *string     `json:"message"`
}

// VocStatusOption is a row from voc_status.
type VocStatusOption struct {
	Code     CaseStatus `json:"code"`
	Label    string     `json:"label"`
	IsClosed bool       `json:"isClosed"`
}
