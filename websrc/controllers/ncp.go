/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

func GetNCPRegions() []string {
	return []string{"kr-standard", "us-standard", "sg-standard", "jp-standard", "de-standard"}
}

// var (
// 	cachedRegions     []string
// 	lastFetched       time.Time
// 	mu                sync.Mutex
// 	cacheTTL          = 10 * time.Minute                                      // TTL 10분
// 	ncpRegionEndpoint = "http://localhost:1323/tumblebug/provider/ncp/region" // 예시 URL
// 	// ncpRegionEndpoint = "http://mc-infra-manager:1323/tumblebug/provider/ncp/region" // 예시 URL
// )

// func GetNCPRegions() []string {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	// 캐시 유효하면 그대로 반환
// 	if time.Since(lastFetched) < cacheTTL && len(cachedRegions) > 0 {
// 		return cachedRegions
// 	}

// 	var result []string
// 	// 아니면 API 호출
// 	req, err := http.NewRequest("GET", ncpRegionEndpoint, nil)
// 	if err != nil {
// 		return result
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	username := "default"
// 	password := "default"
// 	req.SetBasicAuth(username, password)

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return result
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return result
// 	}

// 	var regions models.Regions
// 	if err := json.Unmarshal(body, &regions); err != nil {
// 		return result
// 	}

// 	for _, region := range regions.Regions {
// 		for _, zone := range region.Zones {
// 			result = append(result, fmt.Sprintf("%s-%s", region.RegionName, zone))
// 		}
// 	}
// 	// 캐시 저장
// 	cachedRegions = result
// 	lastFetched = time.Now()

// 	return result
// }
