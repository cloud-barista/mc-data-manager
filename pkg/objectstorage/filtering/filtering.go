package filtering

import (
	"path"
	"strings"
	"time"
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
			if strings.Contains(c.Key, v) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}

	// Regex
	if flt.Regex != nil && !flt.Regex.MatchString(c.Key) {
		return false
	}

	// Size
	if flt.MinSize != nil && c.Size < *flt.MinSize {
		return false
	}
	if flt.MaxSize != nil && c.Size > *flt.MaxSize {
		return false
	}

	// Modified range
	if flt.ModifiedAfter != nil && !c.LastModified.After(*flt.ModifiedAfter) {
		return false
	}
	if flt.ModifiedBefore != nil && !c.LastModified.Before(*flt.ModifiedBefore) {
		return false
	}

	return true
}
