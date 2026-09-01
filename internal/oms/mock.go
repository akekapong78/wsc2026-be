package oms

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockClient is a deterministic in-memory OMS backend for dev/demo,
// seeded from the CA fixtures documented as examples in spec/oms.openapi.yaml.
type MockClient struct {
	mu           sync.Mutex
	networks     map[string]NetworkReference
	activeEvents map[string]*ActiveOutageEvent // keyed by caNumber
	eventSeq     int
	reportSeq    int
}

func NewMockClient() *MockClient {
	now := time.Now()
	return &MockClient{
		networks: map[string]NetworkReference{
			"100000000001": {MeterID: "MTR-001", TransformerID: "TR-001", FeederID: "FDR-01"},
			"100000000002": {MeterID: "MTR-002", TransformerID: "TR-002", FeederID: "FDR-02"},
			"100000000003": {MeterID: "MTR-003", TransformerID: "TR-003", FeederID: "FDR-03"},
		},
		activeEvents: map[string]*ActiveOutageEvent{
			"100000000001": {
				EventID:   "OMS-TR-0001",
				Level:     EventLevelTransformer,
				Status:    StatusInProgress,
				Message:   "พบเหตุไฟฟ้าขัดข้องที่หม้อแปลงซึ่งจ่ายไฟให้ผู้ใช้ไฟรายนี้",
				StartedAt: now.Add(-2 * time.Hour),
			},
			"100000000002": {
				EventID:   "OMS-FDR-0001",
				Level:     EventLevelFeeder,
				Status:    StatusInProgress,
				Message:   "พบเหตุไฟฟ้าขัดข้องระดับฟีดเดอร์ที่ครอบคลุมพื้นที่ของผู้ใช้ไฟรายนี้",
				StartedAt: now.Add(-3 * time.Hour),
			},
			// "100000000003" intentionally has no active event.
		},
		eventSeq: 1,
	}
}

func (m *MockClient) GetOutageByCA(_ context.Context, caNumber string) (*OutageCheckResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	network, ok := m.networks[caNumber]
	if !ok {
		return nil, &ApiError{Status: 404, Code: ErrCANotFound, Message: "ไม่พบหมายเลขผู้ใช้ไฟในระบบ OMS"}
	}

	action := ActionCreateMeter
	event := m.activeEvents[caNumber]
	if event != nil {
		action = ActionInformExisting
	}

	return &OutageCheckResponse{
		CaNumber:          caNumber,
		CustomerFound:     true,
		Network:           network,
		ActiveEvent:       event,
		RecommendedAction: action,
	}, nil
}

func (m *MockClient) CreateOutage(_ context.Context, req CreateOutageRequest) (*CreateOutageResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.networks[req.CaNumber]; !ok {
		return nil, &ApiError{Status: 404, Code: ErrCANotFound, Message: "ไม่พบหมายเลขผู้ใช้ไฟในระบบ OMS"}
	}

	if existing := m.activeEvents[req.CaNumber]; existing != nil {
		return nil, &ApiError{
			Status:          409,
			Code:            ErrActiveEventExist,
			Message:         "พบเหตุการณ์ที่เกี่ยวข้องใน OMS แล้ว",
			ExistingEventID: existing.EventID,
		}
	}

	m.eventSeq++
	eventID := fmt.Sprintf("OMS-METER-%04d", m.eventSeq)
	m.activeEvents[req.CaNumber] = &ActiveOutageEvent{
		EventID:   eventID,
		Level:     EventLevelMeter,
		Status:    StatusReceived,
		Message:   req.Description,
		StartedAt: time.Now(),
	}

	return &CreateOutageResponse{
		EventID:  eventID,
		CaNumber: req.CaNumber,
		Level:    EventLevelMeter,
		Status:   StatusReceived,
		Message:  "OMS รับแจ้งเหตุไฟฟ้าขัดข้องของผู้ใช้ไฟแล้ว",
	}, nil
}

func (m *MockClient) CreateAnonymousOutage(_ context.Context, _ CreateAnonymousOutageRequest) (*CreateAnonymousOutageResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.reportSeq++
	return &CreateAnonymousOutageResponse{
		ReportID: fmt.Sprintf("OMS-ANON-%04d", m.reportSeq),
		Status:   StatusReceived,
		Message:  "OMS รับแจ้งเหตุโดยไม่มีหมายเลขผู้ใช้ไฟแล้ว",
	}, nil
}

var _ Client = (*MockClient)(nil)
