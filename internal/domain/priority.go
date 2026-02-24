package domain

import "strings"

const DefaultPriority = "P2"

func NormalizePriority(priority string) string {
	normalized := strings.ToUpper(strings.TrimSpace(priority))
	if normalized == "" {
		return DefaultPriority
	}
	return normalized
}

func IsValidPriority(priority string) bool {
	switch NormalizePriority(priority) {
	case "P1", "P2", "P3", "P4":
		return true
	default:
		return false
	}
}

func PriorityRank(priority string) int {
	switch NormalizePriority(priority) {
	case "P1":
		return 0
	case "P2":
		return 1
	case "P3":
		return 2
	case "P4":
		return 3
	default:
		return 4
	}
}
