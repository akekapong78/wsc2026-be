// Package voc implements the VOC demo integration group
// (spec/voc.openapi.yaml), mounted under /api/v1/voc/...
package voc

import "time"

type JourneyCode string

const (
	JourneyPowerQuality        JourneyCode = "POWER_QUALITY"
	JourneyServiceIssue        JourneyCode = "SERVICE_ISSUE"
	JourneyPraise              JourneyCode = "PRAISE"
	JourneyTipOff              JourneyCode = "TIP_OFF"
	JourneyStakeholderIssue    JourneyCode = "STAKEHOLDER_ISSUE"
	JourneyStakeholderFeedback JourneyCode = "STAKEHOLDER_FEEDBACK"
)

type Audience string

const (
	AudiencePublic      Audience = "PUBLIC"
	AudienceStakeholder Audience = "STAKEHOLDER"
)

type ReporterMode string

const (
	ReporterRequired ReporterMode = "REQUIRED"
	ReporterOptional ReporterMode = "OPTIONAL"
)

type RequestTypeCode string

const (
	Request1 RequestTypeCode = "REQUEST_1"
	Request2 RequestTypeCode = "REQUEST_2"
	Request3 RequestTypeCode = "REQUEST_3"
	Request4 RequestTypeCode = "REQUEST_4"
	Request6 RequestTypeCode = "REQUEST_6"
	Request7 RequestTypeCode = "REQUEST_7"
	Request8 RequestTypeCode = "REQUEST_8"
)

type CaseStatus string

const (
	StatusSubmitted       CaseStatus = "SUBMITTED"
	StatusAcknowledged    CaseStatus = "ACKNOWLEDGED"
	StatusInProgress      CaseStatus = "IN_PROGRESS"
	StatusWaitingCustomer CaseStatus = "WAITING_CUSTOMER"
	StatusResolved        CaseStatus = "RESOLVED"
	StatusRejected        CaseStatus = "REJECTED"
	StatusCancelled       CaseStatus = "CANCELLED"
)

// --- catalog ---

type JourneyDefinition struct {
	Code                     JourneyCode       `json:"code"`
	Label                    string            `json:"label"`
	Audience                 Audience          `json:"audience"`
	Description              string            `json:"description"`
	ReporterMode             ReporterMode      `json:"reporterMode"`
	ClassificationRootCodes  []RequestTypeCode `json:"classificationRootCodes"`
	RequiresFrequency        bool              `json:"requiresFrequency"`
	RequiresSeverity         bool              `json:"requiresSeverity"`
	RequiresSubIssue         bool              `json:"requiresSubIssue"`
	SupportsCaNumber         bool              `json:"supportsCaNumber"`
	RequiresIncidentLocation bool              `json:"requiresIncidentLocation"`
}

type SubIssueDefinition struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type IssueDefinition struct {
	Code      string               `json:"code"`
	Name      string               `json:"name"`
	SubIssues []SubIssueDefinition `json:"subIssues"`
}

type TopicDefinition struct {
	Code   string            `json:"code"`
	Name   string            `json:"name"`
	Issues []IssueDefinition `json:"issues"`
}

type RequestTypeDefinition struct {
	Code   RequestTypeCode   `json:"code"`
	Name   string            `json:"name"`
	Topics []TopicDefinition `json:"topics"`
}

type IncidentFrequency struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Rank int    `json:"rank"`
}

type SeverityLevel struct {
	Level int    `json:"level"`
	Name  string `json:"name"`
}

type TitlePrefix struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type ServiceArea struct {
	ProvinceCode    string `json:"provinceCode"`
	ProvinceName    string `json:"provinceName"`
	DistrictCode    string `json:"districtCode"`
	DistrictName    string `json:"districtName"`
	SubdistrictCode string `json:"subdistrictCode"`
	SubdistrictName string `json:"subdistrictName"`
	PeaOfficeCode   string `json:"peaOfficeCode"`
	PeaOfficeName   string `json:"peaOfficeName"`
}

type VocCatalogResponse struct {
	Simulation          bool                    `json:"simulation"`
	CatalogVersion      string                  `json:"catalogVersion"`
	Journeys            []JourneyDefinition     `json:"journeys"`
	RequestTypes        []RequestTypeDefinition `json:"requestTypes"`
	IncidentFrequencies []IncidentFrequency     `json:"incidentFrequencies"`
	SeverityLevels      []SeverityLevel         `json:"severityLevels"`
	TitlePrefixes       []TitlePrefix           `json:"titlePrefixes"`
	ServiceAreas        []ServiceArea           `json:"serviceAreas"`
	GeneratedAt         time.Time               `json:"generatedAt"`
}

// --- case create/submit ---

// ContactAddress mirrors components.schemas.ContactAddress.
type ContactAddress struct {
	HouseNumber       *string `json:"houseNumber,omitempty"`
	Moo               *string `json:"moo,omitempty"`
	VillageOrBuilding *string `json:"villageOrBuilding,omitempty"`
	Road              *string `json:"road,omitempty"`
	Soi               *string `json:"soi,omitempty"`
	ProvinceCode      *string `json:"provinceCode,omitempty"`
	DistrictCode      *string `json:"districtCode,omitempty"`
	SubdistrictCode   *string `json:"subdistrictCode,omitempty"`
	PostalCode        *string `json:"postalCode,omitempty"`
}

// Reporter is the union of IdentifiedReporter/OptionalReporter — which
// fields are actually required depends on the journey's reporterMode
// (see validateReporter in validate.go).
type Reporter struct {
	PrefixCode     *string         `json:"prefixCode,omitempty"`
	FirstName      *string         `json:"firstName,omitempty"`
	LastName       *string         `json:"lastName,omitempty"`
	Phone          *string         `json:"phone,omitempty"`
	Email          *string         `json:"email,omitempty"`
	NationalID     *string         `json:"nationalId,omitempty"`
	CaNumber       *string         `json:"caNumber,omitempty"`
	MeterNumber    *string         `json:"meterNumber,omitempty"`
	ContactAddress *ContactAddress `json:"contactAddress,omitempty"`
}

type IncidentLocation struct {
	ProvinceCode      string   `json:"provinceCode"`
	DistrictCode      string   `json:"districtCode"`
	SubdistrictCode   string   `json:"subdistrictCode"`
	PeaOfficeCode     string   `json:"peaOfficeCode"`
	LocationText      string   `json:"locationText"`
	HouseNumber       *string  `json:"houseNumber,omitempty"`
	Moo               *string  `json:"moo,omitempty"`
	VillageOrBuilding *string  `json:"villageOrBuilding,omitempty"`
	Road              *string  `json:"road,omitempty"`
	Soi               *string  `json:"soi,omitempty"`
	Latitude          *float64 `json:"latitude,omitempty"`
	Longitude         *float64 `json:"longitude,omitempty"`
}

type ProductImportance string

const (
	ImportanceNeed        ProductImportance = "NEED"
	ImportanceExpectation ProductImportance = "EXPECTATION"
)

type ClassificationSelection struct {
	RequestTypeCode   RequestTypeCode    `json:"requestTypeCode"`
	TopicCode         string             `json:"topicCode"`
	IssueCode         string             `json:"issueCode"`
	SubIssueCode      *string            `json:"subIssueCode,omitempty"`
	ProductImportance *ProductImportance `json:"productImportance,omitempty"`
}

type ConsentChannel string

const (
	ConsentChat  ConsentChannel = "CHAT"
	ConsentVoice ConsentChannel = "VOICE"
)

type ConsentRecord struct {
	Accepted      bool           `json:"accepted"`
	NoticeVersion string         `json:"noticeVersion"`
	AcceptedAt    time.Time      `json:"acceptedAt"`
	Channel       ConsentChannel `json:"channel"`
}

// CreateVocCaseRequest is the union of all six journeyCode-discriminated
// request variants in the spec — required-ness of reporter/frequencyCode/
// severityLevel/subIssueCode is journey-dependent, checked in validate.go
// against the catalog instead of six separate Go structs.
type CreateVocCaseRequest struct {
	JourneyCode    JourneyCode             `json:"journeyCode"`
	Reporter       *Reporter               `json:"reporter,omitempty"`
	Incident       IncidentLocation        `json:"incident"`
	Classification ClassificationSelection `json:"classification"`
	FrequencyCode  *string                 `json:"frequencyCode,omitempty"`
	SeverityLevel  *int                    `json:"severityLevel,omitempty"`
	Detail         string                  `json:"detail"`
	Consent        ConsentRecord           `json:"consent"`
}

type CaseSubmissionResponse struct {
	Simulation  bool        `json:"simulation"`
	CaseID      string      `json:"caseId"`
	VocNumber   string      `json:"vocNumber"`
	KeyCode     string      `json:"keyCode"`
	Status      CaseStatus  `json:"status"`
	JourneyCode JourneyCode `json:"journeyCode"`
	CreatedAt   time.Time   `json:"createdAt"`
	Message     string      `json:"message"`
}

type CaseLookupRequest struct {
	VocNumber string `json:"vocNumber"`
	KeyCode   string `json:"keyCode"`
}

type ResolvedClassification struct {
	RequestTypeCode   RequestTypeCode    `json:"requestTypeCode"`
	RequestTypeName   string             `json:"requestTypeName"`
	TopicCode         string             `json:"topicCode"`
	TopicName         string             `json:"topicName"`
	IssueCode         string             `json:"issueCode"`
	IssueName         string             `json:"issueName"`
	SubIssueCode      *string            `json:"subIssueCode,omitempty"`
	SubIssueName      *string            `json:"subIssueName,omitempty"`
	ProductImportance *ProductImportance `json:"productImportance,omitempty"`
}

type IncidentSummary struct {
	ProvinceName    string `json:"provinceName"`
	DistrictName    string `json:"districtName"`
	SubdistrictName string `json:"subdistrictName"`
	PeaOfficeName   string `json:"peaOfficeName"`
	LocationText    string `json:"locationText"`
}

type TimelineEntry struct {
	Status     CaseStatus `json:"status"`
	Label      string     `json:"label"`
	OccurredAt time.Time  `json:"occurredAt"`
	Message    string     `json:"message"`
}

type VocCaseDetail struct {
	CaseID         string                 `json:"caseId"`
	VocNumber      string                 `json:"vocNumber"`
	Status         CaseStatus             `json:"status"`
	StatusLabel    string                 `json:"statusLabel"`
	JourneyCode    JourneyCode            `json:"journeyCode"`
	JourneyLabel   string                 `json:"journeyLabel"`
	Classification ResolvedClassification `json:"classification"`
	Incident       IncidentSummary        `json:"incident"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	Timeline       []TimelineEntry        `json:"timeline"`
}

type VocCaseDetailResponse struct {
	Simulation bool          `json:"simulation"`
	Case       VocCaseDetail `json:"case"`
}

type ErrorCode string

const (
	ErrInvalidInput          ErrorCode = "INVALID_INPUT"
	ErrInvalidJourney        ErrorCode = "INVALID_JOURNEY"
	ErrInvalidClassification ErrorCode = "INVALID_CLASSIFICATION"
	ErrInvalidLocation       ErrorCode = "INVALID_LOCATION"
	ErrConsentRequired       ErrorCode = "CONSENT_REQUIRED"
	ErrTrackingNotFound      ErrorCode = "TRACKING_NOT_FOUND"
	ErrIdempotencyConflict   ErrorCode = "IDEMPOTENCY_CONFLICT"
	ErrInternal              ErrorCode = "INTERNAL_ERROR"
	ErrVocUnavailable        ErrorCode = "VOC_UNAVAILABLE"
	ErrCaseNotFound          ErrorCode = "CASE_NOT_FOUND" // admin-only, not in public spec's enum
)

type ApiError struct {
	Status  int
	Code    ErrorCode
	Message string
	Fields  []string
}

func (e *ApiError) Error() string { return e.Message }
