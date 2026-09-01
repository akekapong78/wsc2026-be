-- +goose Up
-- gis_type: 'POINT' = exact ca_no match against the GIS meter database
-- (real coordinate), 'AREA' = no exact match, approximated by averaging
-- meter coordinates near the reported ตำบล/อำเภอ text. NULL = no match at all.
ALTER TABLE oms_outage_events
    ADD COLUMN lat DOUBLE PRECISION,
    ADD COLUMN lon DOUBLE PRECISION,
    ADD COLUMN gis_type TEXT CHECK (gis_type IN ('POINT', 'AREA'));

ALTER TABLE oms_anonymous_reports
    ADD COLUMN lat DOUBLE PRECISION,
    ADD COLUMN lon DOUBLE PRECISION,
    ADD COLUMN gis_type TEXT CHECK (gis_type IN ('POINT', 'AREA'));

-- +goose Down
ALTER TABLE oms_outage_events DROP COLUMN lat, DROP COLUMN lon, DROP COLUMN gis_type;
ALTER TABLE oms_anonymous_reports DROP COLUMN lat, DROP COLUMN lon, DROP COLUMN gis_type;
