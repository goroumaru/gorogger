package gorogger

import "strings"

const (
	NOT_USED OutputLevel = "" // Don't output
	DBG      OutputLevel = "debug"
	INFO     OutputLevel = "info"
	WARN     OutputLevel = "warn"
	ERR      OutputLevel = "error"
)

type OutputLevel string

func GetLevel(lvl string) OutputLevel {
	s := strings.ToLower(
		strings.TrimSpace(lvl),
	)
	switch {
	case contains(s, []string{"debug", "dbg"}):
		return DBG
	case contains(s, []string{"infomation", "info"}):
		return INFO
	case contains(s, []string{"warning", "warn"}):
		return WARN
	case contains(s, []string{"error", "err"}):
		return ERR
	default:
		return NOT_USED
	}
}

func contains(target string, refs []string) bool {
	for _, ref := range refs {
		if strings.Contains(target, ref) {
			return true
		}
	}
	return false
}
