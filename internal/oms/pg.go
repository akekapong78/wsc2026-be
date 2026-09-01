package oms

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgClient is the real OMS backend for this project — queries the tables
// created by migrations/*_create_oms_tables.sql directly. No mock in
// server code; MockClient stays only as a test double.
type PgClient struct {
	pool *pgxpool.Pool
}

func NewPgClient(pool *pgxpool.Pool) *PgClient {
	return &PgClient{pool: pool}
}

func internalErr() error {
	return &ApiError{Status: 500, Code: ErrInternal, Message: "เกิดข้อผิดพลาดภายในระบบ OMS"}
}

func (p *PgClient) GetOutageByCA(ctx context.Context, caNumber string) (*OutageCheckResponse, error) {
	var network NetworkReference
	err := p.pool.QueryRow(ctx,
		`SELECT meter_id, transformer_id, feeder_id FROM oms_customers WHERE ca_number = $1`,
		caNumber,
	).Scan(&network.MeterID, &network.TransformerID, &network.FeederID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &ApiError{Status: 404, Code: ErrCANotFound, Message: "ไม่พบหมายเลขผู้ใช้ไฟในระบบ OMS"}
	}
	if err != nil {
		return nil, internalErr()
	}

	event, err := p.activeEvent(ctx, caNumber)
	if err != nil {
		return nil, err
	}

	action := ActionCreateMeter
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

// activeEvent returns the open (non-RESTORED) event for a CA, or nil if none.
func (p *PgClient) activeEvent(ctx context.Context, caNumber string) (*ActiveOutageEvent, error) {
	var e ActiveOutageEvent
	var level, status string
	err := p.pool.QueryRow(ctx,
		`SELECT event_id, level, status, message, started_at, estimated_restore_at
		 FROM oms_outage_events WHERE ca_number = $1 AND status <> 'RESTORED'
		 ORDER BY started_at DESC LIMIT 1`,
		caNumber,
	).Scan(&e.EventID, &level, &status, &e.Message, &e.StartedAt, &e.EstimatedRestoreAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, internalErr()
	}
	e.Level, e.Status = EventLevel(level), OutageStatus(status)
	return &e, nil
}

func (p *PgClient) CreateOutage(ctx context.Context, req CreateOutageRequest) (*CreateOutageResponse, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, internalErr()
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op once committed

	var exists bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM oms_customers WHERE ca_number = $1)`, req.CaNumber,
	).Scan(&exists); err != nil {
		return nil, internalErr()
	}
	if !exists {
		return nil, &ApiError{Status: 404, Code: ErrCANotFound, Message: "ไม่พบหมายเลขผู้ใช้ไฟในระบบ OMS"}
	}

	var existingEventID string
	err = tx.QueryRow(ctx,
		`SELECT event_id FROM oms_outage_events WHERE ca_number = $1 AND status <> 'RESTORED' LIMIT 1`,
		req.CaNumber,
	).Scan(&existingEventID)
	if err == nil {
		return nil, &ApiError{
			Status: 409, Code: ErrActiveEventExist,
			Message: "พบเหตุการณ์ที่เกี่ยวข้องใน OMS แล้ว", ExistingEventID: existingEventID,
		}
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, internalErr()
	}

	var seq int64
	if err := tx.QueryRow(ctx, `SELECT nextval('oms_meter_event_seq')`).Scan(&seq); err != nil {
		return nil, internalErr()
	}
	eventID := fmt.Sprintf("OMS-METER-%04d", seq)

	if _, err := tx.Exec(ctx,
		`INSERT INTO oms_outage_events (event_id, ca_number, level, status, message)
		 VALUES ($1, $2, 'METER', 'RECEIVED', $3)`,
		eventID, req.CaNumber, req.Description,
	); err != nil {
		return nil, internalErr()
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, internalErr()
	}

	return &CreateOutageResponse{
		EventID: eventID, CaNumber: req.CaNumber, Level: EventLevelMeter, Status: StatusReceived,
		Message: "OMS รับแจ้งเหตุไฟฟ้าขัดข้องของผู้ใช้ไฟแล้ว",
	}, nil
}

func (p *PgClient) CreateAnonymousOutage(ctx context.Context, req CreateAnonymousOutageRequest) (*CreateAnonymousOutageResponse, error) {
	var seq int64
	if err := p.pool.QueryRow(ctx, `SELECT nextval('oms_anon_report_seq')`).Scan(&seq); err != nil {
		return nil, internalErr()
	}
	reportID := fmt.Sprintf("OMS-ANON-%04d", seq)

	if _, err := p.pool.Exec(ctx,
		`INSERT INTO oms_anonymous_reports (report_id, description, location, contact_phone, status)
		 VALUES ($1, $2, $3, $4, 'RECEIVED')`,
		reportID, req.Description, req.Location, req.ContactPhone,
	); err != nil {
		return nil, internalErr()
	}

	return &CreateAnonymousOutageResponse{
		ReportID: reportID, Status: StatusReceived,
		Message: "OMS รับแจ้งเหตุโดยไม่มีหมายเลขผู้ใช้ไฟแล้ว",
	}, nil
}

var _ Client = (*PgClient)(nil)
