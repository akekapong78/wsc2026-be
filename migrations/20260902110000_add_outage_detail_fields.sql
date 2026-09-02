-- +goose Up
-- FE dashboard (wsc2026-fe/src/app/oms/api/outages/route.ts) derived these
-- from event.level instead of storing them — severity/cause/peaBranch were
-- hardcoded per-level, address/affected/vip came from a 3-entry map keyed
-- by CA number. Move them into real columns so admin can set them via the
-- existing PATCH /oms/admin/outages/:eventId instead of the FE guessing.
ALTER TABLE oms_outage_events
    ADD COLUMN severity TEXT NOT NULL DEFAULT 'ต่ำ' CHECK (severity IN ('ต่ำ', 'ปานกลาง', 'สูง', 'วิกฤต')),
    ADD COLUMN cause TEXT NOT NULL DEFAULT '',
    ADD COLUMN pea_branch TEXT NOT NULL DEFAULT '',
    ADD COLUMN address TEXT,
    ADD COLUMN sub_district TEXT,
    ADD COLUMN district TEXT,
    ADD COLUMN province TEXT,
    ADD COLUMN affected_count INT NOT NULL DEFAULT 0,
    ADD COLUMN priority_customers INT NOT NULL DEFAULT 0,
    ADD COLUMN vip_customers INT NOT NULL DEFAULT 0,
    ADD COLUMN vip_details TEXT,
    ADD COLUMN repeated_outage BOOLEAN NOT NULL DEFAULT false,
    -- variable-length activity checklist, e.g.
    -- [{"activity":"...", "estimatedDate":"2026-09-01T10:00:00Z", "actualDate":null, "status":"PENDING"}]
    ADD COLUMN tasks JSONB NOT NULL DEFAULT '[]';

-- backfill the 2 seeded demo events with the values that used to be
-- hardcoded in the FE, keyed by their level (same mapping the FE used)
UPDATE oms_outage_events SET
    severity = 'สูง', cause = 'ฟิวส์แรงสูงหม้อแปลงชำรุด', pea_branch = 'กฟจ.นราธิวาส',
    priority_customers = 2, vip_customers = 1, vip_details = 'สถานีสูบน้ำคลองเปรม',
    affected_count = 80
WHERE event_id = 'OMS-TR-0001';

UPDATE oms_outage_events SET
    severity = 'วิกฤต', cause = 'สายส่งแรงสูงขัดข้อง (Recloser Trip)', pea_branch = 'กฟจ.นราธิวาส',
    priority_customers = 6, vip_customers = 2, vip_details = 'ศูนย์บริการสาธารณสุข',
    affected_count = 500
WHERE event_id = 'OMS-FDR-0001';

-- +goose Down
ALTER TABLE oms_outage_events
    DROP COLUMN severity, DROP COLUMN cause, DROP COLUMN pea_branch,
    DROP COLUMN address, DROP COLUMN sub_district, DROP COLUMN district, DROP COLUMN province,
    DROP COLUMN affected_count, DROP COLUMN priority_customers, DROP COLUMN vip_customers, DROP COLUMN vip_details,
    DROP COLUMN repeated_outage, DROP COLUMN tasks;
