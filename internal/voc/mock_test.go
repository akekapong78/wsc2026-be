package voc

import (
	"context"
	"testing"
	"time"
)

func validPraiseRequest() CreateVocCaseRequest {
	prefix, first, last, phone := "MR", "สมชาย", "ใจดี", "0812345678"
	return CreateVocCaseRequest{
		JourneyCode: JourneyPraise,
		Reporter:    &Reporter{PrefixCode: &prefix, FirstName: &first, LastName: &last, Phone: &phone},
		Incident: IncidentLocation{
			ProvinceCode: "10", DistrictCode: "1001", SubdistrictCode: "100101",
			PeaOfficeCode: "PEA-BKK-01", LocationText: "บริเวณสำนักงานบริการตัวอย่าง",
		},
		Classification: ClassificationSelection{RequestTypeCode: Request3, TopicCode: "STAFF_PRAISE", IssueCode: "GOOD_SERVICE", SubIssueCode: strPtr("SERVICE_MIND")},
		Detail:         "ให้บริการดีมาก",
		Consent:        ConsentRecord{Accepted: true, NoticeVersion: "v1", AcceptedAt: time.Now(), Channel: ConsentChat},
	}
}

func strPtr(s string) *string { return &s }

func TestValidateCreateCase(t *testing.T) {
	if _, err := validateCreateCase(validPraiseRequest()); err != nil {
		t.Fatalf("expected valid request to pass, got %v", err)
	}

	badJourney := validPraiseRequest()
	badJourney.JourneyCode = "NOT_A_JOURNEY"
	if _, err := validateCreateCase(badJourney); err == nil || err.Code != ErrInvalidJourney {
		t.Fatalf("expected ErrInvalidJourney, got %v", err)
	}

	noReporter := validPraiseRequest()
	noReporter.Reporter = nil
	if _, err := validateCreateCase(noReporter); err == nil || err.Code != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput for missing required reporter, got %v", err)
	}

	badClassification := validPraiseRequest()
	badClassification.Classification.IssueCode = "NOT_AN_ISSUE"
	if _, err := validateCreateCase(badClassification); err == nil || err.Code != ErrInvalidClassification {
		t.Fatalf("expected ErrInvalidClassification, got %v", err)
	}

	noConsent := validPraiseRequest()
	noConsent.Consent.Accepted = false
	if _, err := validateCreateCase(noConsent); err == nil || err.Code != ErrConsentRequired {
		t.Fatalf("expected ErrConsentRequired, got %v", err)
	}

	// TIP_OFF: reporter optional, no frequency/severity/subIssue required.
	tipOff := CreateVocCaseRequest{
		JourneyCode: JourneyTipOff,
		Incident: IncidentLocation{
			ProvinceCode: "10", DistrictCode: "1001", SubdistrictCode: "100101",
			PeaOfficeCode: "PEA-BKK-01", LocationText: "บริเวณเสาไฟฟ้า",
		},
		Classification: ClassificationSelection{RequestTypeCode: Request4, TopicCode: "SAFETY", IssueCode: "SUSPICIOUS_ACTIVITY"},
		Detail:         "พบความผิดปกติ",
		Consent:        ConsentRecord{Accepted: true, NoticeVersion: "v1", AcceptedAt: time.Now(), Channel: ConsentChat},
	}
	if _, err := validateCreateCase(tipOff); err != nil {
		t.Fatalf("expected anonymous tip-off to pass, got %v", err)
	}
}

func TestMockClientCreateAndLookupCase(t *testing.T) {
	m := NewMockClient()
	ctx := context.Background()
	req := validPraiseRequest()

	resp, err := m.CreateCase(ctx, "key-1", req)
	if err != nil {
		t.Fatalf("CreateCase: %v", err)
	}
	if !vocNumberRe.MatchString(resp.VocNumber) || !keyCodeRe.MatchString(resp.KeyCode) {
		t.Fatalf("malformed vocNumber/keyCode: %+v", resp)
	}

	// idempotent replay: same key + same payload -> same response.
	resp2, err := m.CreateCase(ctx, "key-1", req)
	if err != nil || resp2.CaseID != resp.CaseID {
		t.Fatalf("expected idempotent replay, got resp=%+v err=%v", resp2, err)
	}

	// same key + different payload -> conflict.
	other := req
	other.Detail = "รายละเอียดอื่น"
	if _, err := m.CreateCase(ctx, "key-1", other); err == nil || err.(*ApiError).Code != ErrIdempotencyConflict {
		t.Fatalf("expected ErrIdempotencyConflict, got %v", err)
	}

	detail, err := m.LookupCase(ctx, CaseLookupRequest{VocNumber: resp.VocNumber, KeyCode: resp.KeyCode})
	if err != nil {
		t.Fatalf("LookupCase: %v", err)
	}
	if detail.Case.CaseID != resp.CaseID || detail.Case.JourneyLabel != "ชื่นชม" {
		t.Fatalf("unexpected detail: %+v", detail.Case)
	}

	if _, err := m.LookupCase(ctx, CaseLookupRequest{VocNumber: resp.VocNumber, KeyCode: "000000"}); err == nil || err.(*ApiError).Code != ErrTrackingNotFound {
		t.Fatalf("expected ErrTrackingNotFound for wrong key code, got %v", err)
	}
}
