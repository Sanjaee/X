package search

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/x-reconstruction/api/internal/domain"
)

var (
	fromRegex       = regexp.MustCompile(`(?i)\bfrom:([a-zA-Z0-9_]+)`)
	toRegex         = regexp.MustCompile(`(?i)\bto:([a-zA-Z0-9_]+)`)
	mentionRegex    = regexp.MustCompile(`(?i)(?:^|\s)@([a-zA-Z0-9_]+)`)
	hasMediaRegex   = regexp.MustCompile(`(?i)\b(?:has:media|filter:media)\b`)
	imagesRegex     = regexp.MustCompile(`(?i)\bfilter:images\b`)
	videosRegex     = regexp.MustCompile(`(?i)\bfilter:videos\b`)
	minLikesRegex   = regexp.MustCompile(`(?i)\b(?:min_faves|min_likes):(\d+)`)
	sinceRegex      = regexp.MustCompile(`(?i)\bsince:(\d{4}-\d{2}-\d{2})`)
	untilRegex      = regexp.MustCompile(`(?i)\buntil:(\d{4}-\d{2}-\d{2})`)
	langRegex       = regexp.MustCompile(`(?i)\blang:([a-zA-Z]{2,5})\b`)
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

// ParseQuery extracts structured search operators and sanitizes the remaining search query text
func ParseQuery(raw string) domain.SearchFilterParams {
	params := domain.SearchFilterParams{
		RawQuery: raw,
		Tab:      "top",
		Limit:    20,
	}

	cleaned := raw

	// 1. from:username
	if m := fromRegex.FindStringSubmatch(cleaned); len(m) > 1 {
		params.FromUsername = m[1]
		cleaned = fromRegex.ReplaceAllString(cleaned, "")
	}

	// 2. to:username
	if m := toRegex.FindStringSubmatch(cleaned); len(m) > 1 {
		params.ToUsername = m[1]
		cleaned = toRegex.ReplaceAllString(cleaned, "")
	}

	// 3. @username mentions
	if m := mentionRegex.FindStringSubmatch(cleaned); len(m) > 1 {
		params.Mentions = m[1]
		cleaned = mentionRegex.ReplaceAllString(cleaned, "")
	}

	// 4. has:media / filter:media
	if hasMediaRegex.MatchString(cleaned) {
		params.HasMedia = true
		cleaned = hasMediaRegex.ReplaceAllString(cleaned, "")
	}

	// 5. filter:images
	if imagesRegex.MatchString(cleaned) {
		params.HasMedia = true
		params.MediaType = "image"
		cleaned = imagesRegex.ReplaceAllString(cleaned, "")
	}

	// 6. filter:videos
	if videosRegex.MatchString(cleaned) {
		params.HasMedia = true
		params.MediaType = "video"
		cleaned = videosRegex.ReplaceAllString(cleaned, "")
	}

	// 7. min_likes / min_faves
	if m := minLikesRegex.FindStringSubmatch(cleaned); len(m) > 1 {
		if val, err := strconv.Atoi(m[1]); err == nil && val >= 0 {
			params.MinLikes = val
		}
		cleaned = minLikesRegex.ReplaceAllString(cleaned, "")
	}

	// 8. since:YYYY-MM-DD
	if m := sinceRegex.FindStringSubmatch(cleaned); len(m) > 1 {
		if t, err := time.Parse("2006-01-02", m[1]); err == nil {
			params.SinceDate = &t
		}
		cleaned = sinceRegex.ReplaceAllString(cleaned, "")
	}

	// 9. until:YYYY-MM-DD
	if m := untilRegex.FindStringSubmatch(cleaned); len(m) > 1 {
		if t, err := time.Parse("2006-01-02", m[1]); err == nil {
			// Include the whole until day by setting to end of day
			eod := t.Add(24*time.Hour - time.Nanosecond)
			params.UntilDate = &eod
		}
		cleaned = untilRegex.ReplaceAllString(cleaned, "")
	}

	// 10. lang:code
	if m := langRegex.FindStringSubmatch(cleaned); len(m) > 1 {
		params.Language = strings.ToLower(m[1])
		cleaned = langRegex.ReplaceAllString(cleaned, "")
	}

	// Clean up remaining text tokens
	cleaned = whitespaceRegex.ReplaceAllString(strings.TrimSpace(cleaned), " ")
	params.Query = cleaned

	return params
}
