package voc

import "regexp"

var (
	locationCodeRe = regexp.MustCompile(`^[0-9A-Z_\-]{1,32}$`)
	frequencyRe    = regexp.MustCompile(`^IIT[0-9]{2}$`)
	phoneRe        = regexp.MustCompile(`^[0-9]{9,10}$`)
	caNumberRe     = regexp.MustCompile(`^[0-9]{12}$`)
	vocNumberRe    = regexp.MustCompile(`^I-[0-9]{8}$`)
	keyCodeRe      = regexp.MustCompile(`^[0-9]{6}$`)
)

// validateCreateCase checks a request against the catalog (journey rules,
// taxonomy tree) and the field-level formats the spec would otherwise
// enforce via JSON Schema. Returns the matching JourneyDefinition on
// success so callers don't have to look it up twice.
func validateCreateCase(req CreateVocCaseRequest) (*JourneyDefinition, *ApiError) {
	journey := findJourney(req.JourneyCode)
	if journey == nil {
		return nil, &ApiError{Status: 400, Code: ErrInvalidJourney, Message: "ไม่รู้จัก journey ที่ระบุ", Fields: []string{"journeyCode"}}
	}

	if err := validateReporter(journey, req.Reporter); err != nil {
		return nil, err
	}
	if err := validateClassification(journey, req.Classification); err != nil {
		return nil, err
	}
	if err := validateIncident(journey, req.Incident); err != nil {
		return nil, err
	}
	if journey.RequiresFrequency && (req.FrequencyCode == nil || !frequencyRe.MatchString(*req.FrequencyCode)) {
		return nil, &ApiError{Status: 400, Code: ErrInvalidInput, Message: "journey นี้ต้องระบุความถี่ของเหตุการณ์", Fields: []string{"frequencyCode"}}
	}
	if journey.RequiresSeverity && (req.SeverityLevel == nil || *req.SeverityLevel < 1 || *req.SeverityLevel > 5) {
		return nil, &ApiError{Status: 400, Code: ErrInvalidInput, Message: "journey นี้ต้องระบุระดับความรุนแรง", Fields: []string{"severityLevel"}}
	}
	if len(req.Detail) < 1 || len(req.Detail) > 2000 {
		return nil, &ApiError{Status: 400, Code: ErrInvalidInput, Message: "กรุณาระบุรายละเอียดเรื่อง", Fields: []string{"detail"}}
	}
	if !req.Consent.Accepted {
		return nil, &ApiError{Status: 400, Code: ErrConsentRequired, Message: "ต้องได้รับความยินยอมก่อนส่งเรื่อง", Fields: []string{"consent"}}
	}

	return journey, nil
}

func validateReporter(journey *JourneyDefinition, r *Reporter) *ApiError {
	if journey.ReporterMode == ReporterOptional && r == nil {
		return nil
	}
	if journey.ReporterMode == ReporterRequired && r == nil {
		return &ApiError{Status: 400, Code: ErrInvalidInput, Message: "journey นี้ต้องระบุผู้แจ้ง", Fields: []string{"reporter"}}
	}
	// r != nil from here — required fields of IdentifiedReporter apply
	// whenever a reporter block is present, REQUIRED or OPTIONAL journey.
	if r.PrefixCode == nil || r.FirstName == nil || r.LastName == nil || r.Phone == nil || !phoneRe.MatchString(*r.Phone) {
		return &ApiError{Status: 400, Code: ErrInvalidInput, Message: "ข้อมูลผู้แจ้งไม่ครบถ้วนหรือไม่ถูกต้อง", Fields: []string{"reporter"}}
	}
	if r.CaNumber != nil && !caNumberRe.MatchString(*r.CaNumber) {
		return &ApiError{Status: 400, Code: ErrInvalidInput, Message: "หมายเลขผู้ใช้ไฟต้องเป็นตัวเลข 12 หลัก", Fields: []string{"reporter.caNumber"}}
	}
	return nil
}

func validateClassification(journey *JourneyDefinition, sel ClassificationSelection) *ApiError {
	inRoot := false
	for _, code := range journey.ClassificationRootCodes {
		if code == sel.RequestTypeCode {
			inRoot = true
			break
		}
	}
	if !inRoot {
		return &ApiError{Status: 400, Code: ErrInvalidClassification, Message: "ประเภทเรื่องไม่สัมพันธ์กับ journey ที่เลือก", Fields: []string{"classification"}}
	}

	_, _, _, subIssueName, ok := findClassification(sel)
	if !ok {
		return &ApiError{Status: 400, Code: ErrInvalidClassification, Message: "ประเภทเรื่องไม่สัมพันธ์กับ taxonomy ที่เลือก", Fields: []string{"classification"}}
	}
	if journey.RequiresSubIssue && subIssueName == nil {
		return &ApiError{Status: 400, Code: ErrInvalidClassification, Message: "journey นี้ต้องระบุ sub-issue", Fields: []string{"classification.subIssueCode"}}
	}
	return nil
}

func validateIncident(journey *JourneyDefinition, inc IncidentLocation) *ApiError {
	if !journey.RequiresIncidentLocation {
		return nil
	}
	if !locationCodeRe.MatchString(inc.ProvinceCode) || !locationCodeRe.MatchString(inc.DistrictCode) ||
		!locationCodeRe.MatchString(inc.SubdistrictCode) || inc.PeaOfficeCode == "" || len(inc.LocationText) < 1 {
		return &ApiError{Status: 400, Code: ErrInvalidLocation, Message: "ข้อมูลสถานที่เกิดเหตุไม่ครบถ้วนหรือไม่ถูกต้อง", Fields: []string{"incident"}}
	}
	return nil
}
