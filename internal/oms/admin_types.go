package oms

import "time"

// OutageTask is one row of the "activity checklist" shown on the FE detail
// panel — stored as a jsonb array on oms_outage_events (variable-length,
// admin-editable, not worth a child table for a demo-scale checklist).
type OutageTask struct {
	Activity      string  `json:"activity"`
	EstimatedDate string  `json:"estimatedDate"`
	ActualDate    *string `json:"actualDate"`
	Status        string  `json:"status"` // COMPLETED | IN_PROGRESS | PENDING
}

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
	Location           *GeoPoint    `json:"location"`
	ContactPhone       *string      `json:"contactPhone"` // optional callback number, see CreateOutageRequest

	// Detail-panel fields — used to be guessed client-side from Level;
	// now real, admin-editable columns (see UpdateOutageEventRequest).
	Severity          string       `json:"severity"`
	Cause             string       `json:"cause"`
	PeaBranch         string       `json:"peaBranch"`
	Address           *string      `json:"address"`
	SubDistrict       *string      `json:"subDistrict"`
	District          *string      `json:"district"`
	Province          *string      `json:"province"`
	AffectedCount     int          `json:"affectedCount"`
	PriorityCustomers int          `json:"priorityCustomers"`
	VipCustomers      int          `json:"vipCustomers"`
	VipDetails        *string      `json:"vipDetails"`
	RepeatedOutage    bool         `json:"repeatedOutage"`
	Tasks             []OutageTask `json:"tasks"`
}

// UpdateOutageEventRequest is a partial update — nil fields are left
// unchanged. Setting Status to a closed code (e.g. RESTORED) closes the
// complaint/work order.
type UpdateOutageEventRequest struct {
	Status             *OutageStatus `json:"status"`
	Message            *string       `json:"message"`
	EstimatedRestoreAt *time.Time    `json:"estimatedRestoreAt"`
	ContactPhone       *string       `json:"contactPhone"`

	Severity          *string       `json:"severity"`
	Cause             *string       `json:"cause"`
	PeaBranch         *string       `json:"peaBranch"`
	Address           *string       `json:"address"`
	SubDistrict       *string       `json:"subDistrict"`
	District          *string       `json:"district"`
	Province          *string       `json:"province"`
	AffectedCount     *int          `json:"affectedCount"`
	PriorityCustomers *int          `json:"priorityCustomers"`
	VipCustomers      *int          `json:"vipCustomers"`
	VipDetails        *string       `json:"vipDetails"`
	RepeatedOutage    *bool         `json:"repeatedOutage"`
	Tasks             *[]OutageTask `json:"tasks"`
}

// StatusOption is a row from oms_status — the valid set an admin can move
// an event into, managed as data instead of a hardcoded CHECK constraint.
type StatusOption struct {
	Code     OutageStatus `json:"code"`
	Label    string       `json:"label"`
	IsClosed bool         `json:"isClosed"`
}

// AdminAnonymousReport is the admin-facing view of a report filed without a
// CA number (POST /oms/outages/anonymous), joined with its status label
// from oms_status the same way AdminOutageEvent is.
type AdminAnonymousReport struct {
	ReportID     string       `json:"reportId"`
	Description  string       `json:"description"`
	Location     string       `json:"location"`
	ContactPhone string       `json:"contactPhone"`
	Status       OutageStatus `json:"status"`
	StatusLabel  string       `json:"statusLabel"`
	CreatedAt    time.Time    `json:"createdAt"`
	GeoLocation  *GeoPoint    `json:"geoLocation"`
}

// UpdateAnonymousReportRequest is a partial update, same shape as
// UpdateOutageEventRequest minus the fields that don't apply here.
type UpdateAnonymousReportRequest struct {
	Status *OutageStatus `json:"status"`
}

type OutageSource string

const (
	SourceOutageEvent     OutageSource = "OUTAGE_EVENT"
	SourceAnonymousReport OutageSource = "ANONYMOUS_REPORT"
)

// AdminOutageEntry is the merged view GET /oms/admin/outages returns —
// oms_outage_events (has caNumber/level, known customer) and
// oms_anonymous_reports (no CA, has contactPhone) shown as one timeline.
// To edit/delete a specific entry, use the dedicated per-source endpoint
// (/outages/:eventId or /anonymous-reports/:reportId) keyed by id.
type AdminOutageEntry struct {
	Source             OutageSource `json:"source"`
	ID                 string       `json:"id"`
	CaNumber           *string      `json:"caNumber"`
	Level              *EventLevel  `json:"level"`
	ContactPhone       *string      `json:"contactPhone"`
	Status             OutageStatus `json:"status"`
	StatusLabel        string       `json:"statusLabel"`
	Message            string       `json:"message"`
	StartedAt          time.Time    `json:"startedAt"`
	EstimatedRestoreAt *time.Time   `json:"estimatedRestoreAt"`
	Location           *GeoPoint    `json:"location"`

	// Detail-panel fields — only populated for SourceOutageEvent entries;
	// anonymous reports don't carry them (zero values / nil).
	Severity          string       `json:"severity"`
	Cause             string       `json:"cause"`
	PeaBranch         string       `json:"peaBranch"`
	Address           *string      `json:"address"`
	SubDistrict       *string      `json:"subDistrict"`
	District          *string      `json:"district"`
	Province          *string      `json:"province"`
	AffectedCount     int          `json:"affectedCount"`
	PriorityCustomers int          `json:"priorityCustomers"`
	VipCustomers      int          `json:"vipCustomers"`
	VipDetails        *string      `json:"vipDetails"`
	RepeatedOutage    bool         `json:"repeatedOutage"`
	Tasks             []OutageTask `json:"tasks"`
}

func entryFromOutageEvent(e AdminOutageEvent) AdminOutageEntry {
	level := e.Level
	return AdminOutageEntry{
		Source: SourceOutageEvent, ID: e.EventID, CaNumber: &e.CaNumber, Level: &level,
		Status: e.Status, StatusLabel: e.StatusLabel, Message: e.Message,
		StartedAt: e.StartedAt, EstimatedRestoreAt: e.EstimatedRestoreAt, Location: e.Location,
		ContactPhone: e.ContactPhone,
		Severity:     e.Severity, Cause: e.Cause, PeaBranch: e.PeaBranch,
		Address: e.Address, SubDistrict: e.SubDistrict, District: e.District, Province: e.Province,
		AffectedCount: e.AffectedCount, PriorityCustomers: e.PriorityCustomers, VipCustomers: e.VipCustomers,
		VipDetails: e.VipDetails, RepeatedOutage: e.RepeatedOutage, Tasks: e.Tasks,
	}
}

func entryFromAnonymousReport(r AdminAnonymousReport) AdminOutageEntry {
	return AdminOutageEntry{
		Source: SourceAnonymousReport, ID: r.ReportID, ContactPhone: &r.ContactPhone,
		Tasks:  []OutageTask{}, // avoid JSON null — FE always .map()s this
		Status: r.Status, StatusLabel: r.StatusLabel, Message: r.Description,
		StartedAt: r.CreatedAt, Location: r.GeoLocation,
		// r.Location is the free-text place the user typed in chat (e.g.
		// "aspine condo") — was dropped here entirely, so the FE fell back
		// to a synthetic "CA: null" string for every anonymous report.
		Address: &r.Location,
	}
}
