package voc

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const adminCaseCols = `
	c.case_id, c.voc_number, c.status, s.label, c.journey_code, c.classification, c.incident, c.reporter,
	c.frequency_code, c.severity_level, c.detail, c.consent, c.created_at, c.updated_at`

func (p *PgClient) ListStatuses(ctx context.Context) ([]VocStatusOption, error) {
	rows, err := p.pool.Query(ctx, `SELECT code, label, is_closed FROM voc_status ORDER BY code`)
	if err != nil {
		return nil, internalErr()
	}
	defer rows.Close()

	out := []VocStatusOption{}
	for rows.Next() {
		var s VocStatusOption
		if err := rows.Scan(&s.Code, &s.Label, &s.IsClosed); err != nil {
			return nil, internalErr()
		}
		out = append(out, s)
	}
	return out, nil
}

func (p *PgClient) ListCases(ctx context.Context) ([]AdminCase, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT `+adminCaseCols+`
		FROM voc_cases c JOIN voc_status s ON s.code = c.status
		ORDER BY c.created_at DESC`)
	if err != nil {
		return nil, internalErr()
	}
	defer rows.Close()

	out := []AdminCase{}
	for rows.Next() {
		a, err := scanAdminCase(rows)
		if err != nil {
			return nil, internalErr()
		}
		out = append(out, a)
	}
	return out, nil
}

func (p *PgClient) GetCase(ctx context.Context, caseID string) (*AdminCase, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT `+adminCaseCols+`
		FROM voc_cases c JOIN voc_status s ON s.code = c.status
		WHERE c.case_id = $1`, caseID)

	a, err := scanAdminCase(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &ApiError{Status: 404, Code: ErrCaseNotFound, Message: "ไม่พบเคสที่ระบุ"}
	}
	if err != nil {
		return nil, internalErr()
	}

	timeline, err := p.timelineFor(ctx, a.CaseID)
	if err != nil {
		return nil, internalErr()
	}
	a.Timeline = timeline
	return &a, nil
}

func (p *PgClient) UpdateCase(ctx context.Context, caseID string, req UpdateCaseRequest) (*AdminCase, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, internalErr()
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op once committed

	var label string
	tag, err := tx.Exec(ctx,
		`UPDATE voc_cases SET status = COALESCE($2, status), updated_at = now() WHERE case_id = $1`,
		caseID, req.Status)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
		return nil, &ApiError{Status: 400, Code: ErrInvalidInput, Message: "status ที่ระบุไม่มีอยู่ในระบบ"}
	}
	if err != nil {
		return nil, internalErr()
	}
	if tag.RowsAffected() == 0 {
		return nil, &ApiError{Status: 404, Code: ErrCaseNotFound, Message: "ไม่พบเคสที่ระบุ"}
	}

	if req.Status != nil {
		label = statusLabel(*req.Status)
		message := label
		if req.Message != nil {
			message = *req.Message
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO voc_case_timeline (case_id, status, label, message) VALUES ($1, $2, $3, $4)`,
			caseID, *req.Status, label, message,
		); err != nil {
			return nil, internalErr()
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, internalErr()
	}
	return p.GetCase(ctx, caseID)
}

func (p *PgClient) DeleteCase(ctx context.Context, caseID string) error {
	tag, err := p.pool.Exec(ctx, `DELETE FROM voc_cases WHERE case_id = $1`, caseID)
	if err != nil {
		return internalErr()
	}
	if tag.RowsAffected() == 0 {
		return &ApiError{Status: 404, Code: ErrCaseNotFound, Message: "ไม่พบเคสที่ระบุ"}
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAdminCase(row scannable) (AdminCase, error) {
	var a AdminCase
	err := row.Scan(&a.CaseID, &a.VocNumber, &a.Status, &a.StatusLabel, &a.JourneyCode,
		&a.Classification, &a.Incident, &a.Reporter, &a.FrequencyCode, &a.SeverityLevel,
		&a.Detail, &a.Consent, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}
