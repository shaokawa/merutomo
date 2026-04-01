package auth

import (
	"regexp"
	"strings"
)

const (
	minUsernameLength      = 3
	maxUsernameLength      = 30
	defaultEmailVisibility = "approval_required"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

func validateRegistrationProfile(displayName, username, emailVisibility string) error {
	if strings.TrimSpace(displayName) == "" {
		return ErrInvalidProfile
	}

	normalizedUsername := normalizeUsername(username)
	if normalizedUsername == "" {
		return ErrInvalidProfile
	}

	if len(normalizedUsername) < minUsernameLength || len(normalizedUsername) > maxUsernameLength {
		return ErrInvalidProfile
	}

	if !usernamePattern.MatchString(normalizedUsername) {
		return ErrInvalidProfile
	}

	if !isValidEmailVisibility(emailVisibility) {
		return ErrInvalidProfile
	}

	return nil
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func normalizeEmailVisibility(emailVisibility string) string {
	normalizedVisibility := strings.TrimSpace(strings.ToLower(emailVisibility))
	if normalizedVisibility == "" {
		return defaultEmailVisibility
	}

	return normalizedVisibility
}

func isValidEmailVisibility(emailVisibility string) bool {
	switch normalizeEmailVisibility(emailVisibility) {
	case "public", "approval_required", "private":
		return true
	default:
		return false
	}
}
