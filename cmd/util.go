package cmd

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

func parseDate(s string) (time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch strings.ToLower(s) {
	case "today":
		return today, nil
	case "tomorrow":
		return today.AddDate(0, 0, 1), nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, now.Location())
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: use YYYY-MM-DD, today or tomorrow", s)
	}
	return t, nil
}

func formatDate(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Local().Format("2006-01-02")
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// matchID returns the single index whose id equals or starts with the given prefix.
func matchID(ids []string, prefix string) (int, error) {
	prefix = strings.ToLower(prefix)
	found := -1
	for i, id := range ids {
		lower := strings.ToLower(id)
		if lower == prefix {
			return i, nil
		}
		if strings.HasPrefix(lower, prefix) {
			if found >= 0 {
				return -1, fmt.Errorf("ambiguous id %q", prefix)
			}
			found = i
		}
	}
	if found < 0 {
		return -1, fmt.Errorf("no match for id %q", prefix)
	}
	return found, nil
}

// Matches the uppercase form produced by Swift's UUID().uuidString.
func newID() string {
	return strings.ToUpper(uuid.NewString())
}
