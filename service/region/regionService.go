package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/cloud-barista/mc-data-manager/models"
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
var regionEndpoint string = "http://localhost:1323/tumblebug/provider/%s/region"

// var regionEndpoint string = "http://mc-infra-manager:1323/tumblebug/provider/%s/region"

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
	endpoint := fmt.Sprintf(regionEndpoint, cspType)

	// API 호출
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil
	}

	req.Header.Set("Content-Type", "application/json")
	username := "default"
	password := "default"
	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var regions models.Regions
	if err := json.Unmarshal(body, &regions); err != nil {
		return nil
	}

	// regionName-zone 평탄화
	var result []string
	for _, region := range regions.Regions {
		result = append(result, region.RegionName)
	}
	sort.Strings(result)

	// 캐시 갱신
	cache[cspType] = &RegionCache{
		Data:        result,
		LastFetched: time.Now(),
	}

	return result
}
