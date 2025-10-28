package service

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
)

// 🔹 공통 캐시 구조
type RegionCache struct {
	Data        []string
	LastFetched time.Time
}

var mu sync.Mutex
var cacheTTL time.Duration = 600 * time.Minute

// 🔹 CSP별 endpoint 맵
var cache map[string]*RegionCache = make(map[string]*RegionCache)

// 🔹 실제 호출 함수 (CSP 공통)
func GetRegions(cspType string) []string {
	mu.Lock()
	defer mu.Unlock()

	// 캐시 확인
	if cache, ok := cache[cspType]; ok {
		if time.Since(cache.LastFetched) < cacheTTL && len(cache.Data) > 0 {
			return cache.Data
		}
	}

	// CSP별 endpoint 확인
	path := "/tumblebug/connConfig?filterRegionRepresentative=true"
	method := http.MethodGet

	// API 호출
	body, err := utils.RequestTumblebug(path, method, "", nil)
	if err != nil {
		return nil
	}

	var conns models.ConnectionConfigList
	if err := json.Unmarshal(body, &conns); err != nil {
		return nil
	}

	var result []string
	for _, connConfig := range conns.ConnectionConfig {
		if connConfig.ProviderName != cspType {
			continue
		}
		result = append(result, connConfig.RegionDetail.RegionName)
	}
	sort.Strings(result)

	// 캐시 갱신
	cache[cspType] = &RegionCache{
		Data:        result,
		LastFetched: time.Now(),
	}

	return result
}
