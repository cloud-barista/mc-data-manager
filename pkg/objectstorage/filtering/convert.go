package filtering

import (
	"regexp"
	"time"

	"github.com/cloud-barista/mc-data-manager/models"
)

func FromParams(p *models.ObjectFilterParams) (*ObjectFilter, error) {
	if p == nil {
		return nil, nil
	}

	var re *regexp.Regexp
	var err error
	if p.Regex != "" {
		re, err = regexp.Compile(p.Regex)
		if err != nil {
			return nil, err
		}
	}

	var after, before *time.Time
	if p.ModifiedAfter != nil && *p.ModifiedAfter != "" {
		t, err := time.Parse(time.RFC3339, *p.ModifiedAfter)
		if err != nil {
			return nil, err
		}
		after = &t
	}
	if p.ModifiedBefore != nil && *p.ModifiedBefore != "" {
		t, err := time.Parse(time.RFC3339, *p.ModifiedBefore)
		if err != nil {
			return nil, err
		}
		before = &t
	}

	return &ObjectFilter{
		Path:              p.Path,
		Contains:          p.Contains,
		Suffixes:          p.Suffixes,
		Exact:             p.Exact,
		Regex:             re,
		MinSize:           p.MinSize,
		MaxSize:           p.MaxSize,
		ModifiedAfter:     after,
		ModifiedBefore:    before,
		SizeFilteringUnit: p.SizeFilteringUnit,
	}, nil
}
