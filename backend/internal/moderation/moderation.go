package moderation

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

//go:embed blocklist.txt
var blocklistData string

var (
	blocklist []string
	urlRegex  = regexp.MustCompile(`https?://[^\s]+`)
)

const (
	maxReviewLength       = 800
	maxVideoTitleLength   = 120
	maxVideoDescription   = 2000
	maxEmojiCount         = 10
	maxURLCount           = 1
)

var ErrModeration = errors.New("content failed moderation")

func init() {
	for _, line := range strings.Split(blocklistData, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			blocklist = append(blocklist, strings.ToLower(trimmed))
		}
	}
}

// ValidateReview ensures a review comment passes moderation.
func ValidateReview(comment string) error {
	if err := validateLength(comment, maxReviewLength, "review comment"); err != nil {
		return err
	}
	if err := validateBlocklist(comment, "review comment"); err != nil {
		return err
	}
	if err := validateEmoji(comment, "review comment"); err != nil {
		return err
	}
	if err := validateURLs(comment, "review comment"); err != nil {
		return err
	}
	return nil
}

// ValidateVideoMeta ensures video metadata is safe.
func ValidateVideoMeta(title, description string) error {
	if err := validateLength(title, maxVideoTitleLength, "video title"); err != nil {
		return err
	}
	if err := validateLength(description, maxVideoDescription, "video description"); err != nil {
		return err
	}
	if err := validateBlocklist(title, "video title"); err != nil {
		return err
	}
	if err := validateBlocklist(description, "video description"); err != nil {
		return err
	}
	if err := validateEmoji(description, "video description"); err != nil {
		return err
	}
	if err := validateURLs(description, "video description"); err != nil {
		return err
	}
	return nil
}

func validateLength(value string, limit int, field string) error {
	if len(strings.TrimSpace(value)) == 0 {
		return fmt.Errorf("%w: %s cannot be empty", ErrModeration, field)
	}
	if len([]rune(value)) > limit {
		return fmt.Errorf("%w: %s exceeds %d characters", ErrModeration, field, limit)
	}
	return nil
}

func validateBlocklist(value, field string) error {
	lower := strings.ToLower(value)
	for _, phrase := range blocklist {
		if phrase != "" && strings.Contains(lower, phrase) {
			return fmt.Errorf("%w: %s contains blocked phrase", ErrModeration, field)
		}
	}
	return nil
}

func validateEmoji(value, field string) error {
	count := 0
	for _, r := range value {
		if isEmoji(r) {
			count++
			if count > maxEmojiCount {
				return fmt.Errorf("%w: %s exceeds emoji limit", ErrModeration, field)
			}
		}
	}
	return nil
}

func validateURLs(value, field string) error {
	if matches := urlRegex.FindAllString(value, -1); len(matches) > maxURLCount {
		return fmt.Errorf("%w: %s contains too many URLs", ErrModeration, field)
	}
	return nil
}

func isEmoji(r rune) bool {
	// Approximation: treat symbols and pictographs as emoji.
	return unicode.Is(unicode.S, r) || (r >= 0x1F300 && r <= 0x1FAFF)
}
