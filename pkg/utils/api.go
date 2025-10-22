package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

func RequestTumblebug(path string, method string, connName string, jsonBody []byte) ([]byte, error) {
	baseUrl := os.Getenv("TUMBLEBUG_URL")
	url := fmt.Sprintf("%s%s", baseUrl, path)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	username := "default"
	password := "default"
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
