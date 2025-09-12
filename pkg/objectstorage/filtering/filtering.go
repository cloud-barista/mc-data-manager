package filtering

import (
	"math"
	"path"
	"strings"
	"time"

    "github.com/rs/zerolog/log"
)

const MiB = int64(1024 * 1024)

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

	mb := float64(c.Size) / (1024 * 1024)
	roundedMB := math.Round(mb*10) / 10
	byteSize := int64(roundedMB * 1024 * 1024)

	log.Debug().
		Int64("original_bytes", c.Size).
		Float64("original_mb", mb).
		Float64("rounded_mb", roundedMB).
		Int64("rounded_bytes", byteSize).
		Msg("[Filter] Size comparison info")

	rbytes := roundedDisplayBytes(c.Size)

	if flt.MinSize != nil && rbytes < *flt.MinSize {
		return false
	}
	if flt.MaxSize != nil && rbytes > *flt.MaxSize {
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


func roundedDisplayBytes(sizeBytes int64) int64 {
	tenthsMiB := (sizeBytes*10 + MiB/2) / MiB 
	return (tenthsMiB * MiB) / 10
}