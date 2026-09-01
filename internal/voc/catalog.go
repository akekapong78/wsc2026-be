package voc

import "time"

// catalog is static MVP data — the 6 journeys and 7 request types are a
// closed set per spec/voc.openapi.yaml (minItems==maxItems==6/7), so this
// is authored once here instead of round-tripping through a DB table.
var catalog = VocCatalogResponse{
	Simulation:     true,
	CatalogVersion: "2026-09-01",
	Journeys: []JourneyDefinition{
		{
			Code: JourneyPowerQuality, Label: "แจ้งปัญหาคุณภาพไฟฟ้า", Audience: AudiencePublic,
			Description:              "ไฟตก ไฟดับ แรงดันไฟฟ้าไม่คงที่ หรือเหตุขัดข้องด้านคุณภาพไฟฟ้า",
			ReporterMode:             ReporterRequired,
			ClassificationRootCodes:  []RequestTypeCode{Request6},
			RequiresFrequency:        true,
			RequiresSeverity:         true,
			RequiresSubIssue:         true,
			SupportsCaNumber:         true,
			RequiresIncidentLocation: true,
		},
		{
			Code: JourneyServiceIssue, Label: "แจ้งปัญหาด้านบริการ", Audience: AudiencePublic,
			Description:              "ปัญหาที่สาขา การติดต่อเจ้าหน้าที่ หรือบริการดิจิทัล",
			ReporterMode:             ReporterRequired,
			ClassificationRootCodes:  []RequestTypeCode{Request1, Request2},
			RequiresFrequency:        true,
			RequiresSeverity:         true,
			RequiresSubIssue:         true,
			SupportsCaNumber:         true,
			RequiresIncidentLocation: true,
		},
		{
			Code: JourneyPraise, Label: "ชื่นชม", Audience: AudiencePublic,
			Description:              "แบ่งปันประสบการณ์ที่ดีและคำชม",
			ReporterMode:             ReporterRequired,
			ClassificationRootCodes:  []RequestTypeCode{Request3},
			RequiresFrequency:        false,
			RequiresSeverity:         false,
			RequiresSubIssue:         true,
			SupportsCaNumber:         true,
			RequiresIncidentLocation: true,
		},
		{
			Code: JourneyTipOff, Label: "แจ้งเบาะแส", Audience: AudiencePublic,
			Description:              "แจ้งความผิดปกติหรือความเสี่ยงด้านความปลอดภัย",
			ReporterMode:             ReporterOptional,
			ClassificationRootCodes:  []RequestTypeCode{Request4},
			RequiresFrequency:        false,
			RequiresSeverity:         false,
			RequiresSubIssue:         false,
			SupportsCaNumber:         true,
			RequiresIncidentLocation: true,
		},
		{
			Code: JourneyStakeholderIssue, Label: "แจ้งปัญหาการดำเนินงาน", Audience: AudienceStakeholder,
			Description:              "คู่ค้าหรือผู้มีส่วนได้ส่วนเสียแจ้งปัญหาการดำเนินงาน",
			ReporterMode:             ReporterRequired,
			ClassificationRootCodes:  []RequestTypeCode{Request7},
			RequiresFrequency:        false,
			RequiresSeverity:         false,
			RequiresSubIssue:         false,
			SupportsCaNumber:         false,
			RequiresIncidentLocation: true,
		},
		{
			Code: JourneyStakeholderFeedback, Label: "ชื่นชม เสนอแนะ ข้อคิดเห็น", Audience: AudienceStakeholder,
			Description:              "คู่ค้าหรือผู้มีส่วนได้ส่วนเสียส่งคำชม ข้อเสนอแนะ หรือข้อคิดเห็น",
			ReporterMode:             ReporterRequired,
			ClassificationRootCodes:  []RequestTypeCode{Request8},
			RequiresFrequency:        false,
			RequiresSeverity:         false,
			RequiresSubIssue:         false,
			SupportsCaNumber:         false,
			RequiresIncidentLocation: true,
		},
	},
	RequestTypes: []RequestTypeDefinition{
		{
			Code: Request1, Name: "ร้องเรียน",
			Topics: []TopicDefinition{
				{Code: "SERVICE", Name: "การให้บริการ", Issues: []IssueDefinition{
					{Code: "SERVICE_DELAY", Name: "บริการล่าช้า", SubIssues: []SubIssueDefinition{
						{Code: "CONTACT_DELAY", Name: "ติดต่อกลับล่าช้า"},
					}},
				}},
			},
		},
		{
			Code: Request2, Name: "ข้อเสนอแนะ/ข้อคิดเห็น",
			Topics: []TopicDefinition{
				{Code: "SERVICE_FEEDBACK", Name: "ข้อเสนอแนะด้านบริการ", Issues: []IssueDefinition{
					{Code: "DIGITAL_SERVICE_SUGGESTION", Name: "ข้อเสนอแนะบริการดิจิทัล", SubIssues: []SubIssueDefinition{
						{Code: "APPLICATION_IMPROVEMENT", Name: "ปรับปรุงแอปพลิเคชัน"},
					}},
				}},
			},
		},
		{
			Code: Request3, Name: "ชื่นชม",
			Topics: []TopicDefinition{
				{Code: "STAFF_PRAISE", Name: "ชื่นชมบุคลากร", Issues: []IssueDefinition{
					{Code: "GOOD_SERVICE", Name: "ให้บริการดี", SubIssues: []SubIssueDefinition{
						{Code: "SERVICE_MIND", Name: "มีจิตบริการ"},
					}},
				}},
			},
		},
		{
			Code: Request4, Name: "แจ้งเบาะแส",
			Topics: []TopicDefinition{
				{Code: "SAFETY", Name: "ความปลอดภัย", Issues: []IssueDefinition{
					{Code: "SUSPICIOUS_ACTIVITY", Name: "พบความผิดปกติหรือความเสี่ยง", SubIssues: []SubIssueDefinition{}},
				}},
			},
		},
		{
			Code: Request6, Name: "แจ้งเหตุ",
			Topics: []TopicDefinition{
				{Code: "POWER_QUALITY", Name: "คุณภาพไฟฟ้า", Issues: []IssueDefinition{
					{Code: "VOLTAGE_ISSUE", Name: "แรงดันไฟฟ้าไม่คงที่", SubIssues: []SubIssueDefinition{
						{Code: "FREQUENT_VOLTAGE_DROP", Name: "ไฟตกบ่อย"},
					}},
				}},
			},
		},
		{
			Code: Request7, Name: "เสียงของผู้มีส่วนได้ส่วนเสีย - แจ้งปัญหา",
			Topics: []TopicDefinition{
				{Code: "STAKEHOLDER_OPERATIONS", Name: "การดำเนินงานกับผู้มีส่วนได้ส่วนเสีย", Issues: []IssueDefinition{
					{Code: "PROCUREMENT_PROCESS", Name: "กระบวนการจัดซื้อจัดจ้าง", SubIssues: []SubIssueDefinition{}},
				}},
			},
		},
		{
			Code: Request8, Name: "เสียงของผู้มีส่วนได้ส่วนเสีย - ชื่นชม/ข้อเสนอแนะ/ข้อคิดเห็น",
			Topics: []TopicDefinition{
				{Code: "STAKEHOLDER_FEEDBACK", Name: "ข้อเสนอแนะจากผู้มีส่วนได้ส่วนเสีย", Issues: []IssueDefinition{
					{Code: "PROCESS_IMPROVEMENT", Name: "ข้อเสนอแนะปรับปรุงกระบวนการ", SubIssues: []SubIssueDefinition{}},
				}},
			},
		},
	},
	IncidentFrequencies: []IncidentFrequency{
		{Code: "IIT01", Name: "เกิดเหตุการณ์เป็นครั้งแรก", Rank: 1},
		{Code: "IIT02", Name: "เดือนละ 1 ครั้ง", Rank: 2},
		{Code: "IIT03", Name: "มากกว่าเดือนละ 1 ครั้ง", Rank: 3},
		{Code: "IIT04", Name: "มากกว่าสัปดาห์ละ 1 ครั้ง", Rank: 4},
		{Code: "IIT05", Name: "มากกว่าวันละ 1 ครั้ง", Rank: 5},
	},
	SeverityLevels: []SeverityLevel{
		{Level: 1, Name: "ผลกระทบต่ำ"},
		{Level: 2, Name: "ผลกระทบเล็กน้อย"},
		{Level: 3, Name: "ผลกระทบปานกลาง"},
		{Level: 4, Name: "ผลกระทบสูง"},
		{Level: 5, Name: "ผลกระทบรุนแรง"},
	},
	TitlePrefixes: []TitlePrefix{
		{Code: "MR", Label: "นาย"},
		{Code: "MRS", Label: "นาง"},
		{Code: "MS", Label: "นางสาว"},
		{Code: "OTHER", Label: "อื่น ๆ"},
	},
	ServiceAreas: []ServiceArea{
		{
			ProvinceCode: "10", ProvinceName: "กรุงเทพมหานคร",
			DistrictCode: "1001", DistrictName: "เขตพระนคร",
			SubdistrictCode: "100101", SubdistrictName: "พระบรมมหาราชวัง",
			PeaOfficeCode: "PEA-BKK-01", PeaOfficeName: "การไฟฟ้าที่รับผิดชอบพื้นที่ตัวอย่างกรุงเทพมหานคร",
		},
	},
}

// GetCatalog stamps a fresh generatedAt on every call — everything else is
// static.
func GetCatalog() VocCatalogResponse {
	c := catalog
	c.GeneratedAt = time.Now()
	return c
}

func findJourney(code JourneyCode) *JourneyDefinition {
	for i := range catalog.Journeys {
		if catalog.Journeys[i].Code == code {
			return &catalog.Journeys[i]
		}
	}
	return nil
}

// findClassification walks requestType -> topic -> issue -> subIssue,
// returning the resolved names or nil if any leg doesn't match.
func findClassification(sel ClassificationSelection) (reqName, topicName, issueName string, subIssueName *string, ok bool) {
	for _, rt := range catalog.RequestTypes {
		if rt.Code != sel.RequestTypeCode {
			continue
		}
		for _, t := range rt.Topics {
			if t.Code != sel.TopicCode {
				continue
			}
			for _, iss := range t.Issues {
				if iss.Code != sel.IssueCode {
					continue
				}
				if sel.SubIssueCode == nil {
					return rt.Name, t.Name, iss.Name, nil, true
				}
				for _, si := range iss.SubIssues {
					if si.Code == *sel.SubIssueCode {
						name := si.Name
						return rt.Name, t.Name, iss.Name, &name, true
					}
				}
				return "", "", "", nil, false
			}
		}
	}
	return "", "", "", nil, false
}

func findServiceArea(provinceCode, districtCode, subdistrictCode, peaOfficeCode string) *ServiceArea {
	for i := range catalog.ServiceAreas {
		a := &catalog.ServiceAreas[i]
		if a.ProvinceCode == provinceCode && a.DistrictCode == districtCode &&
			a.SubdistrictCode == subdistrictCode && a.PeaOfficeCode == peaOfficeCode {
			return a
		}
	}
	return nil
}
