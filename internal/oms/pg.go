package oms

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgClient is the real OMS backend for this project — queries the tables
// created by migrations/*_create_oms_tables.sql directly. No mock in
// server code; MockClient stays only as a test double.
type PgClient struct {
	pool *pgxpool.Pool
	gis  *GisClient // nil = coordinate lookup disabled (GIS_DBSTRING not set)
}

func NewPgClient(pool *pgxpool.Pool, gis *GisClient) *PgClient {
	return &PgClient{pool: pool, gis: gis}
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
	var lat, lon *float64
	var gisType *string
	err := p.pool.QueryRow(ctx,
		`SELECT event_id, level, status, message, started_at, estimated_restore_at, lat, lon, gis_type
		 FROM oms_outage_events WHERE ca_number = $1 AND status <> 'RESTORED'
		 ORDER BY started_at DESC LIMIT 1`,
		caNumber,
	).Scan(&e.EventID, &level, &status, &e.Message, &e.StartedAt, &e.EstimatedRestoreAt, &lat, &lon, &gisType)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, internalErr()
	}
	e.Level, e.Status = EventLevel(level), OutageStatus(status)
	if lat != nil && lon != nil {
		e.Location = &GeoPoint{Lat: lat, Lon: lon, GisType: gisType}
	}
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
		`INSERT INTO oms_outage_events (event_id, ca_number, level, status, message, contact_phone)
		 VALUES ($1, $2, 'METER', 'RECEIVED', $3, $4)`,
		eventID, req.CaNumber, req.Description, req.ContactPhone,
	); err != nil {
		return nil, internalErr()
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, internalErr()
	}

	addressText := req.Description
	if req.LocationNote != nil {
		addressText = *req.LocationNote + " " + addressText
	}
	location := p.enrichOutageLocation(ctx, eventID, req.CaNumber, addressText)

	return &CreateOutageResponse{
		EventID: eventID, CaNumber: req.CaNumber, Level: EventLevelMeter, Status: StatusReceived,
		Message:  "OMS รับแจ้งเหตุไฟฟ้าขัดข้องของผู้ใช้ไฟแล้ว",
		Location: location,
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

	location := p.enrichAnonymousLocation(ctx, reportID, req.Location+" "+req.Description)

	return &CreateAnonymousOutageResponse{
		ReportID: reportID, Status: StatusReceived,
		Message:  "OMS รับแจ้งเหตุโดยไม่มีหมายเลขผู้ใช้ไฟแล้ว",
		Location: location,
	}, nil
}

// enrichOutageLocation is best-effort: a GIS lookup failure or miss must
// never fail the outage report itself, so errors are logged and swallowed.
func (p *PgClient) enrichOutageLocation(ctx context.Context, eventID, caNumber, addressText string) *GeoPoint {
	if p.gis == nil {
		return nil
	}
	loc, err := p.gis.Lookup(ctx, caNumber, addressText)
	if err != nil {
		log.Printf("gis lookup failed for outage %s: %v", eventID, err)
		return nil
	}
	if loc == nil {
		return nil
	}
	if _, err := p.pool.Exec(ctx,
		`UPDATE oms_outage_events SET lat = $1, lon = $2, gis_type = $3 WHERE event_id = $4`,
		loc.Lat, loc.Lon, loc.GisType, eventID,
	); err != nil {
		log.Printf("failed to store gis location for outage %s: %v", eventID, err)
		return nil
	}
	return &GeoPoint{Lat: &loc.Lat, Lon: &loc.Lon, GisType: &loc.GisType}
}

func (p *PgClient) enrichAnonymousLocation(ctx context.Context, reportID, addressText string) *GeoPoint {
	if p.gis == nil {
		return nil
	}
	loc, err := p.gis.Lookup(ctx, "", addressText)
	if err != nil {
		log.Printf("gis lookup failed for anonymous report %s: %v", reportID, err)
		return nil
	}
	if loc == nil {
		return nil
	}
	if _, err := p.pool.Exec(ctx,
		`UPDATE oms_anonymous_reports SET lat = $1, lon = $2, gis_type = $3 WHERE report_id = $4`,
		loc.Lat, loc.Lon, loc.GisType, reportID,
	); err != nil {
		log.Printf("failed to store gis location for anonymous report %s: %v", reportID, err)
		return nil
	}
	return &GeoPoint{Lat: &loc.Lat, Lon: &loc.Lon, GisType: &loc.GisType}
}

var _ Client = (*PgClient)(nil)
