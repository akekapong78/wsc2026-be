-- +goose Up
-- oms_anonymous_reports.status had no FK (just a free-text default) — link
-- it to oms_status so admin can list/filter/validate it the same way as
-- oms_outage_events.status.
ALTER TABLE oms_anonymous_reports
    ADD CONSTRAINT oms_anonymous_reports_status_fkey
    FOREIGN KEY (status) REFERENCES oms_status (code);

-- +goose Down
ALTER TABLE oms_anonymous_reports DROP CONSTRAINT oms_anonymous_reports_status_fkey;
