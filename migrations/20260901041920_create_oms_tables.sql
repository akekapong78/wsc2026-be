-- +goose Up
CREATE TABLE oms_customers (
    ca_number CHAR(12) PRIMARY KEY,
    meter_id TEXT NOT NULL,
    transformer_id TEXT NOT NULL,
    feeder_id TEXT NOT NULL
);

CREATE TABLE oms_outage_events (
    event_id TEXT PRIMARY KEY,
    ca_number CHAR(12) NOT NULL REFERENCES oms_customers (ca_number),
    level TEXT NOT NULL CHECK (level IN ('METER', 'TRANSFORMER', 'FEEDER')),
    status TEXT NOT NULL CHECK (status IN ('RECEIVED', 'ACKNOWLEDGED', 'IN_PROGRESS', 'RESTORED')),
    message TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    estimated_restore_at TIMESTAMPTZ
);

-- only one open (non-RESTORED) event per CA at a time — mirrors the
-- 409 ACTIVE_EVENT_EXISTS rule in spec/oms.openapi.yaml
CREATE UNIQUE INDEX oms_outage_events_active_ca_idx
    ON oms_outage_events (ca_number)
    WHERE status <> 'RESTORED';

CREATE TABLE oms_anonymous_reports (
    report_id TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    location TEXT NOT NULL,
    contact_phone TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'RECEIVED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE SEQUENCE oms_meter_event_seq START 2; -- 1 reserved by seeded fixtures below
CREATE SEQUENCE oms_anon_report_seq START 1;

-- seed fixtures, same CA numbers documented as examples in spec/oms.openapi.yaml
INSERT INTO oms_customers (ca_number, meter_id, transformer_id, feeder_id) VALUES
    ('100000000001', 'MTR-001', 'TR-001', 'FDR-01'),
    ('100000000002', 'MTR-002', 'TR-002', 'FDR-02'),
    ('100000000003', 'MTR-003', 'TR-003', 'FDR-03');

INSERT INTO oms_outage_events (event_id, ca_number, level, status, message, started_at) VALUES
    ('OMS-TR-0001', '100000000001', 'TRANSFORMER', 'IN_PROGRESS',
     'พบเหตุไฟฟ้าขัดข้องที่หม้อแปลงซึ่งจ่ายไฟให้ผู้ใช้ไฟรายนี้', now() - interval '2 hours'),
    ('OMS-FDR-0001', '100000000002', 'FEEDER', 'IN_PROGRESS',
     'พบเหตุไฟฟ้าขัดข้องระดับฟีดเดอร์ที่ครอบคลุมพื้นที่ของผู้ใช้ไฟรายนี้', now() - interval '3 hours');
-- 100000000003 intentionally has no active event

-- +goose Down
DROP TABLE oms_anonymous_reports;
DROP TABLE oms_outage_events;
DROP TABLE oms_customers;
DROP SEQUENCE oms_meter_event_seq;
DROP SEQUENCE oms_anon_report_seq;
