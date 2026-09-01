-- +goose Up
CREATE TABLE oms_status (
    code TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    is_closed BOOLEAN NOT NULL DEFAULT false
);

INSERT INTO oms_status (code, label, is_closed) VALUES
    ('RECEIVED', 'รับแจ้งแล้ว', false),
    ('ACKNOWLEDGED', 'รับทราบแล้ว', false),
    ('IN_PROGRESS', 'กำลังดำเนินการ', false),
    ('RESTORED', 'ปิดงาน / จ่ายไฟคืนแล้ว', true);

-- swap the hardcoded CHECK for a FK to oms_status, so admin can see/manage
-- the valid status set as data instead of a migration every time
ALTER TABLE oms_outage_events DROP CONSTRAINT oms_outage_events_status_check;
ALTER TABLE oms_outage_events
    ADD CONSTRAINT oms_outage_events_status_fkey
    FOREIGN KEY (status) REFERENCES oms_status (code);

-- +goose Down
ALTER TABLE oms_outage_events DROP CONSTRAINT oms_outage_events_status_fkey;
ALTER TABLE oms_outage_events
    ADD CONSTRAINT oms_outage_events_status_check
    CHECK (status IN ('RECEIVED', 'ACKNOWLEDGED', 'IN_PROGRESS', 'RESTORED'));
DROP TABLE oms_status;
