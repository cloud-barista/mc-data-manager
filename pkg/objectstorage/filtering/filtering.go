package filtering

import (
	"math"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Candidate struct {
	Key          string
	Size         int64
	LastModified time.Time
}

func MatchCandidate(flt *ObjectFilter, c Candidate) bool {
	if flt == nil {
		return true
	}

	// Exact
	if len(flt.Exact) > 0 {
		ok := false
		base := path.Base(c.Key)
		for _, e := range flt.Exact {
			if c.Key == e || base == e {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}

	// Suffix
	if len(flt.Suffixes) > 0 {
		ok := false
		for _, s := range flt.Suffixes {
			if strings.HasSuffix(c.Key, s) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}

	// Contains
	if len(flt.Contains) > 0 {
		ok := false
		for _, v := range flt.Contains {
			if strings.Contains(strings.ToLower(c.Key), strings.ToLower(v)) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}

	rbytes := math.Round(roundedUnit(c.Size, flt.SizeFilteringUnit)*10) / 10

	if flt.MinSize != nil && rbytes < *flt.MinSize {
		return false
	}
	if flt.MaxSize != nil && rbytes > *flt.MaxSize {
		return false
	}

	if flt.ModifiedAfter != nil {
		after := flt.ModifiedAfter.UTC()
		if c.LastModified.UTC().Before(after) {
			return false
		}
	}
	
	if flt.ModifiedBefore != nil {
		before := flt.ModifiedBefore.UTC()
		if c.LastModified.UTC().After(before) {
			return false
		}
	}

	return true
}

func roundedUnit(sizeBytes int64, unit string) float64 {
	log.Debug().Str("sizeUnit", unit).Msg("[data filtering size unit]")
	switch strings.ToUpper(unit) {
	case "GB":
		return float64(sizeBytes) / (1024 * 1024 * 1024)
	case "MB":
		return float64(sizeBytes) / (1024 * 1024)
	case "KB":
		return float64(sizeBytes) / 1024
	default:
		return float64(sizeBytes)
	}
}
