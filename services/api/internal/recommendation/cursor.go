package recommendation

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EncodeCursor produces an opaque base64 string from created timestamp and post ID
func EncodeCursor(createdAt time.Time, postID uuid.UUID) string {
	raw := fmt.Sprintf("%d:%s", createdAt.UnixNano(), postID.String())
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor parses an opaque base64 cursor token into time and UUID
func DecodeCursor(cursorStr string) (*FeedCursor, error) {
	if cursorStr == "" {
		return nil, nil
	}

	rawBytes, err := base64.URLEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding: %w", err)
	}

	parts := strings.Split(string(rawBytes), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("malformed cursor payload")
	}

	nano, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor timestamp: %w", err)
	}

	pID, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid cursor uuid: %w", err)
	}

	return &FeedCursor{
		CreatedAt: time.Unix(0, nano).UTC(),
		PostID:    pID,
	}, nil
}
