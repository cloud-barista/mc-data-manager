package filtering

import (
	"regexp"
	"time"
)

type ObjectFilter struct {
	Prefix         string
	Contains       []string
	Suffixes       []string
	Exact          []string
	Regex          *regexp.Regexp
	MinSize        *int64
	MaxSize        *int64
	ModifiedAfter  *time.Time
	ModifiedBefore *time.Time
}
