package voc

import "context"

// Client is the VOC backend seam — swap MockClient (dev/demo, deterministic)
// for a PgClient without touching handler.go.
type Client interface {
	GetCatalog(ctx context.Context) (*VocCatalogResponse, error)
	// CreateCase must be idempotent on idempotencyKey: same key + same
	// request returns the original response; same key + different request
	// returns an ApiError with Code == ErrIdempotencyConflict.
	CreateCase(ctx context.Context, idempotencyKey string, req CreateVocCaseRequest) (*CaseSubmissionResponse, error)
	LookupCase(ctx context.Context, req CaseLookupRequest) (*VocCaseDetailResponse, error)
}
