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

import "time"

type Credential struct {
	CredentialId   uint64    `gorm:"column:credentialId;primaryKey;autoIncrement" json:"credentialId"`
	CspType        string    `gorm:"column:cspType;size:50;not null" json:"cspType"`
	Name           string    `gorm:"column:name;size:150" json:"name,omitempty"`
	CredentialJson string    `gorm:"column:credentialJson;type:longtext"json:"credentialJson,omitempty"`
	CreatedAt      time.Time `gorm:"column:createAt;autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updateAt;autoUpdateTime" json:"updatedAt,omitempty"`
}

// TableName 명시 (Go struct -> DB 테이블명 매핑)
func (Credential) TableName() string {
	return "tbCredential"
}
