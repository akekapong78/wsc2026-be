package oms

import (
	"context"
	"errors"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GisLocation is a best-effort coordinate for an outage report, meant for
// the separate FE-osm map client.
type GisLocation struct {
	Lat     float64
	Lon     float64
	GisType string // "POINT" (exact meter match) or "AREA" (approximated)
}

// GisClient looks up coordinates from the PEA GIS meter database
// (gis_l table, a different Supabase project from the app DB — see
// GIS_DBSTRING). Read-only, best-effort: callers must not fail a request
// just because a lookup errors or finds nothing.
type GisClient struct {
	pool *pgxpool.Pool
}

func NewGisClient(pool *pgxpool.Pool) *GisClient {
	return &GisClient{pool: pool}
}

var (
	tambonRe = regexp.MustCompile(`ต\.[^\s]+`)
	amphoeRe = regexp.MustCompile(`อ\.[^\s]+`)
)

// Lookup tries an exact ca_no match first (POINT — real meter coordinate),
// then falls back to averaging meter coordinates whose address text
// mentions the same ตำบล/อำเภอ extracted from addressText (AREA — only an
// approximation). Returns nil, nil when nothing is found.
func (g *GisClient) Lookup(ctx context.Context, caNumber, addressText string) (*GisLocation, error) {
	if caNumber != "" {
		var lat, lon float64
		err := g.pool.QueryRow(ctx,
			`SELECT lat, lon FROM gis_l WHERE ca_no = $1 AND lat IS NOT NULL ORDER BY loaded_at DESC LIMIT 1`,
			caNumber,
		).Scan(&lat, &lon)
		if err == nil {
			return &GisLocation{Lat: lat, Lon: lon, GisType: "POINT"}, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	for _, area := range extractAreas(addressText) {
		var lat, lon *float64
		var n int
		err := g.pool.QueryRow(ctx,
			`SELECT AVG(lat), AVG(lon), COUNT(*) FROM gis_l WHERE address LIKE '%' || $1 || '%' AND lat IS NOT NULL`,
			area,
		).Scan(&lat, &lon, &n)
		if err != nil {
			return nil, err
		}
		if n > 0 && lat != nil && lon != nil {
			return &GisLocation{Lat: *lat, Lon: *lon, GisType: "AREA"}, nil
		}
	}

	return nil, nil
}

// extractAreas pulls ตำบล (narrower) then อำเภอ (wider) tokens out of free
// text, narrowest first so an exact ตำบล match is tried before อำเภอ.
func extractAreas(text string) []string {
	var out []string
	if m := tambonRe.FindString(text); m != "" {
		out = append(out, m)
	}
	if m := amphoeRe.FindString(text); m != "" {
		out = append(out, m)
	}
	return out
}
