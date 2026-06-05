package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

var (
	runtimeNsId   string
	runtimeNsIdMu sync.RWMutex
)

func SetNsId(nsId string) {
	runtimeNsIdMu.Lock()
	defer runtimeNsIdMu.Unlock()
	runtimeNsId = nsId
}

func GetNsId() string {
	runtimeNsIdMu.RLock()
	if runtimeNsId != "" {
		defer runtimeNsIdMu.RUnlock()
		return runtimeNsId
	}
	runtimeNsIdMu.RUnlock()

	nsId := os.Getenv("TUMBLEBUG_NS_ID")
	if nsId == "" {
		nsId = "default"
	}
	return nsId
}

func RequestTumblebug(path string, method string, connName string, jsonBody []byte) ([]byte, error) {
	baseUrl := os.Getenv("TUMBLEBUG_URL")
	url := fmt.Sprintf("%s%s", baseUrl, path)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	username := os.Getenv("TUMBLEBUG_USERNAME")
	if username == "" {
		username = "default"
	}
	password := os.Getenv("TUMBLEBUG_PASSWORD")
	if password == "" {
		password = "default"
	}
	req.SetBasicAuth(username, password)
	req.Header.Add("Accept", "application/json")
	if connName != "" {
		req.Header.Set("credential", connName)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	// HTTP 코드 확인
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: unexpected status %d, response: %s",
			resp.StatusCode, string(body))
	}

	return body, nil
}
