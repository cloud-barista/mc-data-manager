package filtering

import (
	"regexp"
	"time"
)

type ObjectFilter struct {
	Path              string
	Contains          []string
	Suffixes          []string
	Exact             []string
	Regex             *regexp.Regexp
	MinSize           *float64
	MaxSize           *float64
	ModifiedAfter     *time.Time
	ModifiedBefore    *time.Time
	SizeFilteringUnit string
}
