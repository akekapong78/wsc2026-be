-- +goose Up
-- CreateOutageRequest.ContactPhone (POST /oms/outages, known-CA case) was
-- accepted by the handler but silently dropped — never persisted. Only
-- oms_anonymous_reports had a phone column. Optional (not required, unlike
-- the anonymous-report flow) so the CA number itself already identifies the
-- customer; phone is just a callback number for the ops team.
ALTER TABLE oms_outage_events
    ADD COLUMN contact_phone TEXT;

-- +goose Down
ALTER TABLE oms_outage_events DROP COLUMN contact_phone;
