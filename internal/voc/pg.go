package voc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgClient is the real VOC backend — queries the tables created by
// migrations/*_create_voc_tables.sql directly. No mock in server code;
// MockClient stays only as a test double.
type PgClient struct {
	pool *pgxpool.Pool
}

func NewPgClient(pool *pgxpool.Pool) *PgClient {
	return &PgClient{pool: pool}
}

func internalErr() error {
	return &ApiError{Status: 500, Code: ErrInternal, Message: "เกิดข้อผิดพลาดภายในระบบ VOC"}
}

func (p *PgClient) GetCatalog(_ context.Context) (*VocCatalogResponse, error) {
	c := GetCatalog()
	return &c, nil
}

func (p *PgClient) CreateCase(ctx context.Context, idempotencyKey string, req CreateVocCaseRequest) (*CaseSubmissionResponse, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, internalErr()
	}
	hash := sha256.Sum256(reqJSON)
	requestHash := hex.EncodeToString(hash[:])

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, internalErr()
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op once committed

	var existingHash string
	var existingResp []byte
	err = tx.QueryRow(ctx,
		`SELECT request_hash, response FROM voc_idempotency WHERE idempotency_key = $1`, idempotencyKey,
	).Scan(&existingHash, &existingResp)
	if err == nil {
		if existingHash != requestHash {
			return nil, &ApiError{Status: 409, Code: ErrIdempotencyConflict, Message: "คีย์รายการนี้ถูกใช้กับข้อมูลอื่นแล้ว"}
		}
		var resp CaseSubmissionResponse
		if err := json.Unmarshal(existingResp, &resp); err != nil {
			return nil, internalErr()
		}
		return &resp, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, internalErr()
	}

	var seq int64
	if err := tx.QueryRow(ctx, `SELECT nextval('voc_case_seq')`).Scan(&seq); err != nil {
		return nil, internalErr()
	}

	record := &caseRecord{
		CaseID:         uuid.NewString(),
		VocNumber:      fmt.Sprintf("I-%08d", seq),
		KeyCode:        randomKeyCode(),
		Status:         StatusSubmitted,
		JourneyCode:    req.JourneyCode,
		Classification: req.Classification,
		Incident:       req.Incident,
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO voc_cases
			(case_id, voc_number, key_code, status, journey_code, classification, incident, reporter,
			 frequency_code, severity_level, detail, consent)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING created_at, updated_at`,
		record.CaseID, record.VocNumber, record.KeyCode, record.Status, record.JourneyCode,
		req.Classification, req.Incident, req.Reporter, req.FrequencyCode, req.SeverityLevel, req.Detail, req.Consent,
	).Scan(&record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return nil, internalErr()
	}

	timelineMsg := "ระบบได้รับเรื่อง VOC แล้ว"
	if _, err := tx.Exec(ctx,
		`INSERT INTO voc_case_timeline (case_id, status, label, message) VALUES ($1, $2, $3, $4)`,
		record.CaseID, StatusSubmitted, statusLabel(StatusSubmitted), timelineMsg,
	); err != nil {
		return nil, internalErr()
	}

	resp := record.toSubmissionResponse("รับเรื่อง VOC แบบสาธิตเรียบร้อยแล้ว")
	respJSON, err := json.Marshal(resp)
	if err != nil {
		return nil, internalErr()
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO voc_idempotency (idempotency_key, request_hash, response) VALUES ($1, $2, $3)`,
		idempotencyKey, requestHash, respJSON,
	); err != nil {
		return nil, internalErr()
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, internalErr()
	}
	return resp, nil
}

func (p *PgClient) LookupCase(ctx context.Context, req CaseLookupRequest) (*VocCaseDetailResponse, error) {
	notFound := &ApiError{Status: 404, Code: ErrTrackingNotFound, Message: "ไม่พบเคสสำหรับข้อมูลติดตามที่ระบุ"}

	record, err := p.scanCase(ctx,
		`SELECT case_id, voc_number, status, journey_code, classification, incident, created_at, updated_at
		 FROM voc_cases WHERE voc_number = $1 AND key_code = $2`,
		req.VocNumber, req.KeyCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, notFound
	}
	if err != nil {
		return nil, internalErr()
	}

	timeline, err := p.timelineFor(ctx, record.CaseID)
	if err != nil {
		return nil, internalErr()
	}
	record.Timeline = timeline

	return record.toDetailResponse(), nil
}

func (p *PgClient) scanCase(ctx context.Context, query string, args ...any) (*caseRecord, error) {
	var r caseRecord
	err := p.pool.QueryRow(ctx, query, args...).Scan(
		&r.CaseID, &r.VocNumber, &r.Status, &r.JourneyCode, &r.Classification, &r.Incident, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (p *PgClient) timelineFor(ctx context.Context, caseID string) ([]TimelineEntry, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT status, label, occurred_at, message FROM voc_case_timeline WHERE case_id = $1 ORDER BY occurred_at`,
		caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []TimelineEntry{}
	for rows.Next() {
		var t TimelineEntry
		if err := rows.Scan(&t.Status, &t.Label, &t.OccurredAt, &t.Message); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

var _ Client = (*PgClient)(nil)
