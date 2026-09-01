package voc

import (
	"context"
	"crypto/rand"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type idempotencyEntry struct {
	request  CreateVocCaseRequest
	response *CaseSubmissionResponse
}

// MockClient is a deterministic in-memory VOC backend for dev/demo.
type MockClient struct {
	mu          sync.Mutex
	cases       map[string]*caseRecord // keyed by caseID
	byVocNumber map[string]*caseRecord
	idempotency map[string]idempotencyEntry
	caseSeq     int
}

func NewMockClient() *MockClient {
	return &MockClient{
		cases:       map[string]*caseRecord{},
		byVocNumber: map[string]*caseRecord{},
		idempotency: map[string]idempotencyEntry{},
	}
}

func (m *MockClient) GetCatalog(_ context.Context) (*VocCatalogResponse, error) {
	c := GetCatalog()
	return &c, nil
}

func (m *MockClient) CreateCase(_ context.Context, idempotencyKey string, req CreateVocCaseRequest) (*CaseSubmissionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry, ok := m.idempotency[idempotencyKey]; ok {
		if reflect.DeepEqual(entry.request, req) {
			return entry.response, nil
		}
		return nil, &ApiError{Status: 409, Code: ErrIdempotencyConflict, Message: "คีย์รายการนี้ถูกใช้กับข้อมูลอื่นแล้ว"}
	}

	m.caseSeq++
	now := time.Now()
	record := &caseRecord{
		CaseID:         fmt.Sprintf("00000000-0000-0000-0000-%012d", m.caseSeq),
		VocNumber:      fmt.Sprintf("I-%08d", m.caseSeq),
		KeyCode:        randomKeyCode(),
		Status:         StatusSubmitted,
		JourneyCode:    req.JourneyCode,
		Classification: req.Classification,
		Incident:       req.Incident,
		CreatedAt:      now,
		UpdatedAt:      now,
		Timeline: []TimelineEntry{
			{Status: StatusSubmitted, Label: statusLabel(StatusSubmitted), OccurredAt: now, Message: "ระบบได้รับเรื่อง VOC แล้ว"},
		},
	}
	m.cases[record.CaseID] = record
	m.byVocNumber[record.VocNumber] = record

	resp := record.toSubmissionResponse("รับเรื่อง VOC แบบสาธิตเรียบร้อยแล้ว")
	m.idempotency[idempotencyKey] = idempotencyEntry{request: req, response: resp}
	return resp, nil
}

func (m *MockClient) LookupCase(_ context.Context, req CaseLookupRequest) (*VocCaseDetailResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, ok := m.byVocNumber[req.VocNumber]
	if !ok || record.KeyCode != req.KeyCode {
		return nil, &ApiError{Status: 404, Code: ErrTrackingNotFound, Message: "ไม่พบเคสสำหรับข้อมูลติดตามที่ระบุ"}
	}
	return record.toDetailResponse(), nil
}

func randomKeyCode() string {
	var b [3]byte
	_, _ = rand.Read(b[:])
	n := (int(b[0])<<16 | int(b[1])<<8 | int(b[2])) % 1000000
	return fmt.Sprintf("%06d", n)
}

var _ Client = (*MockClient)(nil)
