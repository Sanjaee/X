package search

import (
	"testing"
	"time"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		wantQuery     string
		wantFrom      string
		wantTo        string
		wantHasMedia  bool
		wantMediaType string
		wantMinLikes  int
		wantSince     string
	}{
		{
			name:      "simple text query",
			raw:       "golang programming",
			wantQuery: "golang programming",
		},
		{
			name:          "operators extracted properly",
			raw:           "crypto from:vitalik min_faves:50 filter:images since:2026-01-01",
			wantQuery:     "crypto",
			wantFrom:      "vitalik",
			wantHasMedia:  true,
			wantMediaType: "image",
			wantMinLikes:  50,
			wantSince:     "2026-01-01",
		},
		{
			name:      "replies to user",
			raw:       "announcement to:satyanadella",
			wantQuery: "announcement",
			wantTo:    "satyanadella",
		},
		{
			name:          "filter videos",
			raw:           "gameplay trailer filter:videos",
			wantQuery:     "gameplay trailer",
			wantHasMedia:  true,
			wantMediaType: "video",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQuery(tt.raw)
			if got.Query != tt.wantQuery {
				t.Errorf("got Query %q, want %q", got.Query, tt.wantQuery)
			}
			if got.FromUsername != tt.wantFrom {
				t.Errorf("got FromUsername %q, want %q", got.FromUsername, tt.wantFrom)
			}
			if got.ToUsername != tt.wantTo {
				t.Errorf("got ToUsername %q, want %q", got.ToUsername, tt.wantTo)
			}
			if got.HasMedia != tt.wantHasMedia {
				t.Errorf("got HasMedia %v, want %v", got.HasMedia, tt.wantHasMedia)
			}
			if got.MediaType != tt.wantMediaType {
				t.Errorf("got MediaType %q, want %q", got.MediaType, tt.wantMediaType)
			}
			if got.MinLikes != tt.wantMinLikes {
				t.Errorf("got MinLikes %d, want %d", got.MinLikes, tt.wantMinLikes)
			}
			if tt.wantSince != "" {
				expectedSince, _ := time.Parse("2006-01-02", tt.wantSince)
				if got.SinceDate == nil || !got.SinceDate.Equal(expectedSince) {
					t.Errorf("got SinceDate %v, want %v", got.SinceDate, expectedSince)
				}
			}
		})
	}
}
