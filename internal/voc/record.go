package voc

import "time"

var statusLabels = map[CaseStatus]string{
	StatusSubmitted:       "รับเรื่องแล้ว",
	StatusAcknowledged:    "รับทราบแล้ว",
	StatusInProgress:      "อยู่ระหว่างดำเนินการ",
	StatusWaitingCustomer: "รอข้อมูลเพิ่มเติมจากผู้แจ้ง",
	StatusResolved:        "ดำเนินการเสร็จสิ้น",
	StatusRejected:        "ปฏิเสธเรื่อง",
	StatusCancelled:       "ยกเลิกเรื่อง",
}

func statusLabel(s CaseStatus) string {
	if l, ok := statusLabels[s]; ok {
		return l
	}
	return string(s)
}

// caseRecord is the shared shape MockClient and PgClient both build
// responses from, so the catalog-lookup logic (journey label, resolved
// classification names) lives in one place.
type caseRecord struct {
	CaseID         string
	VocNumber      string
	KeyCode        string
	Status         CaseStatus
	JourneyCode    JourneyCode
	Classification ClassificationSelection
	Incident       IncidentLocation
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Timeline       []TimelineEntry
}

func (r *caseRecord) toDetailResponse() *VocCaseDetailResponse {
	journey := findJourney(r.JourneyCode)
	journeyLabel := string(r.JourneyCode)
	if journey != nil {
		journeyLabel = journey.Label
	}

	reqName, topicName, issueName, subIssueName, _ := findClassification(r.Classification)

	area := findServiceArea(r.Incident.ProvinceCode, r.Incident.DistrictCode, r.Incident.SubdistrictCode, r.Incident.PeaOfficeCode)
	incidentSummary := IncidentSummary{LocationText: r.Incident.LocationText}
	if area != nil {
		incidentSummary.ProvinceName = area.ProvinceName
		incidentSummary.DistrictName = area.DistrictName
		incidentSummary.SubdistrictName = area.SubdistrictName
		incidentSummary.PeaOfficeName = area.PeaOfficeName
	}

	return &VocCaseDetailResponse{
		Simulation: true,
		Case: VocCaseDetail{
			CaseID:       r.CaseID,
			VocNumber:    r.VocNumber,
			Status:       r.Status,
			StatusLabel:  statusLabel(r.Status),
			JourneyCode:  r.JourneyCode,
			JourneyLabel: journeyLabel,
			Classification: ResolvedClassification{
				RequestTypeCode:   r.Classification.RequestTypeCode,
				RequestTypeName:   reqName,
				TopicCode:         r.Classification.TopicCode,
				TopicName:         topicName,
				IssueCode:         r.Classification.IssueCode,
				IssueName:         issueName,
				SubIssueCode:      r.Classification.SubIssueCode,
				SubIssueName:      subIssueName,
				ProductImportance: r.Classification.ProductImportance,
			},
			Incident:  incidentSummary,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			Timeline:  r.Timeline,
		},
	}
}

func (r *caseRecord) toSubmissionResponse(message string) *CaseSubmissionResponse {
	return &CaseSubmissionResponse{
		Simulation:  true,
		CaseID:      r.CaseID,
		VocNumber:   r.VocNumber,
		KeyCode:     r.KeyCode,
		Status:      r.Status,
		JourneyCode: r.JourneyCode,
		CreatedAt:   r.CreatedAt,
		Message:     message,
	}
}
