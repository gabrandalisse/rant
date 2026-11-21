package rant

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidSeverity = errors.New("severity must be between 1 and 5")
	ErrEmptyContent    = errors.New("content cannot be empty")
)

type Severity int

const (
	SeverityMild     Severity = 1
	SeverityModerate Severity = 2
	SeverityHigh     Severity = 3
	SeveritySevere   Severity = 4
	SeverityCritical Severity = 5
)

func ParseSeverity(value int) (Severity, error) {
	if value < int(SeverityMild) || value > int(SeverityCritical) {
		return 0, ErrInvalidSeverity
	}

	return Severity(value), nil
}

type Rant struct {
	ID        string
	CreatedAt time.Time
	Severity  Severity
	Content   string
	Tags      []string
}

type EncryptedRecord struct {
	ID               string
	CreatedAt        time.Time
	DayOfWeek        int
	HourOfDay        int
	Severity         Severity
	EncryptedPayload []byte
	Nonce            []byte
	Tags             []string
}

func normalizeTags(rawTags []string) []string {
	seen := make(map[string]struct{}, len(rawTags))
	cleaned := make([]string, 0, len(rawTags))

	for _, raw := range rawTags {
		tag := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(raw, "#")))

		if tag == "" {
			continue
		}

		if _, exists := seen[tag]; exists {
			continue
		}

		seen[tag] = struct{}{}
		cleaned = append(cleaned, tag)
	}

	return cleaned
}
