-- +goose Up
-- swap the synthetic 100000000001..003 fixtures for real CA numbers /
-- meter_id / lat / lon pulled from the GIS meter database (gis_l), so
-- oms mockups line up with data that actually exists there.
-- transformer_id = gis_l.point_code (real); feeder_id has no source in
-- gis_l, mocked from the point_code's area prefix.
DELETE FROM oms_outage_events WHERE ca_number IN ('100000000001', '100000000002', '100000000003');
DELETE FROM oms_customers WHERE ca_number IN ('100000000001', '100000000002', '100000000003');

INSERT INTO oms_customers (ca_number, meter_id, transformer_id, feeder_id) VALUES
    ('020007161258', '19163911',   'LRAG0202', 'LRAG-FDR'),
    ('020007992928', '6500481663', 'LNTW9077', 'LNTW-FDR'),
    ('020008040804', '6500482063', 'LBJO0202', 'LBJO-FDR'),
    ('020008076556', '21539803',   'LNTW0799', 'LNTW-FDR'),
    ('020008084432', '6500482402', 'LRAG9999', 'LRAG-FDR'),
    ('020008084480', '6500482354', 'LRAG9999', 'LRAG-FDR'),
    ('020026125630', '6101432525', 'LNTW8005', 'LNTW-FDR'),
    ('020026730121', '6201325412', 'LRAG9999', 'LRAG-FDR'),
    ('020027219860', '6602846772', 'LNTW0329', 'LNTW-FDR'),
    ('020027350582', '6602613526', 'LNTW0327', 'LNTW-FDR'),
    ('020027545639', '6600024803', 'LBJO0202', 'LBJO-FDR');

-- 2 of the 11 have a pre-existing active event (transformer/feeder level,
-- same demo shape as before); the rest have none — same as the old
-- 003-has-no-event fixture, just spread across real CA numbers now.
-- lat/lon/gis_type filled straight from gis_l (POINT — exact meter match).
INSERT INTO oms_outage_events (event_id, ca_number, level, status, message, started_at, lat, lon, gis_type) VALUES
    ('OMS-TR-0001', '020027219860', 'TRANSFORMER', 'IN_PROGRESS',
     'พบเหตุไฟฟ้าขัดข้องที่หม้อแปลงซึ่งจ่ายไฟให้ผู้ใช้ไฟรายนี้', now() - interval '2 hours',
     6.41051256, 101.8209484, 'POINT'),
    ('OMS-FDR-0001', '020007992928', 'FEEDER', 'IN_PROGRESS',
     'พบเหตุไฟฟ้าขัดข้องระดับฟีดเดอร์ที่ครอบคลุมพื้นที่ของผู้ใช้ไฟรายนี้', now() - interval '3 hours',
     6.42665666, 101.80049166, 'POINT');

-- +goose Down
DELETE FROM oms_outage_events WHERE ca_number IN (
    '020007161258', '020007992928', '020008040804', '020008076556', '020008084432',
    '020008084480', '020026125630', '020026730121', '020027219860', '020027350582', '020027545639'
);
DELETE FROM oms_customers WHERE ca_number IN (
    '020007161258', '020007992928', '020008040804', '020008076556', '020008084432',
    '020008084480', '020026125630', '020026730121', '020027219860', '020027350582', '020027545639'
);

INSERT INTO oms_customers (ca_number, meter_id, transformer_id, feeder_id) VALUES
    ('100000000001', 'MTR-001', 'TR-001', 'FDR-01'),
    ('100000000002', 'MTR-002', 'TR-002', 'FDR-02'),
    ('100000000003', 'MTR-003', 'TR-003', 'FDR-03');

INSERT INTO oms_outage_events (event_id, ca_number, level, status, message, started_at) VALUES
    ('OMS-TR-0001', '100000000001', 'TRANSFORMER', 'IN_PROGRESS',
     'พบเหตุไฟฟ้าขัดข้องที่หม้อแปลงซึ่งจ่ายไฟให้ผู้ใช้ไฟรายนี้', now() - interval '2 hours'),
    ('OMS-FDR-0001', '100000000002', 'FEEDER', 'IN_PROGRESS',
     'พบเหตุไฟฟ้าขัดข้องระดับฟีดเดอร์ที่ครอบคลุมพื้นที่ของผู้ใช้ไฟรายนี้', now() - interval '3 hours');
