-- +goose Up
CREATE TABLE voc_cases (
    case_id UUID PRIMARY KEY,
    voc_number TEXT NOT NULL UNIQUE,
    key_code TEXT NOT NULL,
    status TEXT NOT NULL REFERENCES voc_status (code) DEFAULT 'SUBMITTED',
    journey_code TEXT NOT NULL,
    -- classification/incident/reporter/consent are write-once per case and
    -- only re-read whole (admin view, lookup) — JSONB avoids a dozen
    -- columns for fields the catalog, not this table, is authoritative on.
    classification JSONB NOT NULL,
    incident JSONB NOT NULL,
    reporter JSONB,
    frequency_code TEXT,
    severity_level INT,
    detail TEXT NOT NULL,
    consent JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE voc_case_timeline (
    id BIGSERIAL PRIMARY KEY,
    case_id UUID NOT NULL REFERENCES voc_cases (case_id) ON DELETE CASCADE,
    status TEXT NOT NULL REFERENCES voc_status (code),
    label TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    message TEXT NOT NULL
);

-- Idempotency-Key ledger for POST /api/v1/voc/cases — same key + same
-- request replays the stored response; same key + different request is a
-- 409 IDEMPOTENCY_CONFLICT (checked in application code against request_hash).
CREATE TABLE voc_idempotency (
    idempotency_key TEXT PRIMARY KEY,
    request_hash TEXT NOT NULL,
    response JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE SEQUENCE voc_case_seq START 1;

-- +goose Down
DROP TABLE voc_idempotency;
DROP TABLE voc_case_timeline;
DROP TABLE voc_cases;
DROP SEQUENCE voc_case_seq;
