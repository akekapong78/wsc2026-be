package oms

import "context"

// Client is the OMS backend seam — swap MockClient (dev/demo, deterministic)
// for a RestClient (real OMS service) without touching handler.go.
type Client interface {
	GetOutageByCA(ctx context.Context, caNumber string) (*OutageCheckResponse, error)
	CreateOutage(ctx context.Context, req CreateOutageRequest) (*CreateOutageResponse, error)
	CreateAnonymousOutage(ctx context.Context, req CreateAnonymousOutageRequest) (*CreateAnonymousOutageResponse, error)
}
