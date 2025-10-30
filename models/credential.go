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
package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type CredentialCreateRequest struct {
	CspType        string          `json:"cspType"`
	Name           string          `json:"name,omitempty"`
	CredentialJson json.RawMessage `json:"credentialJson,omitempty" swaggertype:"object"`
	S3AccessKey    string          `json:"s3AccessKey"`
	S3SecretKey    string          `json:"s3SecretKey"`
}

type CredentialListResponse struct {
	CredentialId uint64 `json:"credentialId"`
	CspType      string `json:"cspType"`
	Name         string `json:"name,omitempty"`
}

type Credential struct {
	CredentialId   uint64    `gorm:"column:credentialId;primaryKey;autoIncrement" json:"credentialId"`
	CspType        string    `gorm:"column:cspType;size:50;not null" json:"cspType"`
	Name           string    `gorm:"column:name;size:150" json:"name,omitempty"`
	CredentialJson string    `gorm:"column:credentialJson;type:longtext" json:"credentialJson,omitempty"`
	CreatedAt      time.Time `gorm:"column:createAt;autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updateAt;autoUpdateTime" json:"updatedAt,omitempty"`
}

// TableName 명시 (Go struct -> DB 테이블명 매핑)
func (Credential) TableName() string {
	return "tbCredential"
}

func (cr *CredentialCreateRequest) GetCredential() (string, error) {
	if len(cr.CredentialJson) == 0 {
		return "", fmt.Errorf("credentialJson is empty")
	}

	switch cr.CspType {
	case "aws":
		var aws AWSCredentials
		if err := json.Unmarshal(cr.CredentialJson, &aws); err != nil {
			return "", fmt.Errorf("invalid aws credential json: %w", err)
		}

		b, _ := json.Marshal(aws)
		return string(b), nil
	case "ncp":
		var ncp NCPCredentials
		if err := json.Unmarshal(cr.CredentialJson, &ncp); err != nil {
			return "", fmt.Errorf("invalid ncp credential json: %w", err)
		}

		b, _ := json.Marshal(ncp)
		return string(b), nil
	case "gcp":
		var gcp GCPCredentials
		if err := json.Unmarshal(cr.CredentialJson, &gcp); err != nil {
			return "", fmt.Errorf("invalid gcp credential json: %w", err)
		}

		b, _ := json.Marshal(gcp)
		return string(b), nil

	default:
		return "", fmt.Errorf("unsupported cspType: %q", cr.CspType)
	}
}
