package oms

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const outageEventCols = `
	e.event_id, e.ca_number, e.level, e.status, s.label, e.message, e.started_at, e.estimated_restore_at`

func (p *PgClient) ListStatuses(ctx context.Context) ([]StatusOption, error) {
	rows, err := p.pool.Query(ctx, `SELECT code, label, is_closed FROM oms_status ORDER BY code`)
	if err != nil {
		return nil, internalErr()
	}
	defer rows.Close()

	out := []StatusOption{}
	for rows.Next() {
		var s StatusOption
		var code string
		if err := rows.Scan(&code, &s.Label, &s.IsClosed); err != nil {
			return nil, internalErr()
		}
		s.Code = OutageStatus(code)
		out = append(out, s)
	}
	return out, nil
}

func (p *PgClient) ListOutageEvents(ctx context.Context) ([]AdminOutageEvent, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT `+outageEventCols+`
		FROM oms_outage_events e JOIN oms_status s ON s.code = e.status
		ORDER BY e.started_at DESC`)
	if err != nil {
		return nil, internalErr()
	}
	defer rows.Close()

	out := []AdminOutageEvent{}
	for rows.Next() {
		e, err := scanAdminOutageEvent(rows)
		if err != nil {
			return nil, internalErr()
		}
		out = append(out, e)
	}
	return out, nil
}

func (p *PgClient) GetOutageEvent(ctx context.Context, eventID string) (*AdminOutageEvent, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT `+outageEventCols+`
		FROM oms_outage_events e JOIN oms_status s ON s.code = e.status
		WHERE e.event_id = $1`, eventID)

	e, err := scanAdminOutageEvent(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &ApiError{Status: 404, Code: ErrEventNotFound, Message: "ไม่พบเหตุการณ์ที่ระบุ"}
	}
	if err != nil {
		return nil, internalErr()
	}
	return &e, nil
}

func (p *PgClient) UpdateOutageEvent(ctx context.Context, eventID string, req UpdateOutageEventRequest) (*AdminOutageEvent, error) {
	row := p.pool.QueryRow(ctx, `
		UPDATE oms_outage_events
		SET status = COALESCE($2, status),
		    message = COALESCE($3, message),
		    estimated_restore_at = COALESCE($4, estimated_restore_at)
		WHERE event_id = $1
		RETURNING event_id, ca_number, level, status, message, started_at, estimated_restore_at`,
		eventID, req.Status, req.Message, req.EstimatedRestoreAt)

	var ignored AdminOutageEvent
	err := row.Scan(&ignored.EventID, &ignored.CaNumber, &ignored.Level, &ignored.Status,
		&ignored.Message, &ignored.StartedAt, &ignored.EstimatedRestoreAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &ApiError{Status: 404, Code: ErrEventNotFound, Message: "ไม่พบเหตุการณ์ที่ระบุ"}
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503": // foreign_key_violation — status code not in oms_status
			return nil, &ApiError{Status: 400, Code: ErrInvalidInput, Message: "status ที่ระบุไม่มีอยู่ในระบบ"}
		case "23505": // unique_violation — active-event-per-CA index
			return nil, &ApiError{Status: 409, Code: ErrActiveEventExist, Message: "มีเหตุการณ์ที่เปิดอยู่ของผู้ใช้ไฟรายนี้แล้ว"}
		}
	}
	if err != nil {
		return nil, internalErr()
	}

	// re-fetch so the response carries the joined status label
	return p.GetOutageEvent(ctx, eventID)
}

func (p *PgClient) DeleteOutageEvent(ctx context.Context, eventID string) error {
	tag, err := p.pool.Exec(ctx, `DELETE FROM oms_outage_events WHERE event_id = $1`, eventID)
	if err != nil {
		return internalErr()
	}
	if tag.RowsAffected() == 0 {
		return &ApiError{Status: 404, Code: ErrEventNotFound, Message: "ไม่พบเหตุการณ์ที่ระบุ"}
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAdminOutageEvent(row scannable) (AdminOutageEvent, error) {
	var e AdminOutageEvent
	var level, status string
	err := row.Scan(&e.EventID, &e.CaNumber, &level, &status, &e.StatusLabel,
		&e.Message, &e.StartedAt, &e.EstimatedRestoreAt)
	e.Level, e.Status = EventLevel(level), OutageStatus(status)
	return e, err
}
