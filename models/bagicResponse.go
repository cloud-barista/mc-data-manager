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
	"github.com/cloud-barista/mc-data-manager/pkg/rdbms/mysql/diagnostics"
	"github.com/cloud-barista/mc-data-manager/pkg/sysbench"
)

type BasicPageResponse struct {
	Content string  `json:"Content"`
	Error   *string `json:"Error"`
	OS      string  `json:"OS"`
	TmpPath string  `json:"TmpPath"`

	Regions    []string `json:"Regions"`
	AWSRegions []string `json:"AWSRegions"`
	GCPRegions []string `json:"GCPRegions"`
	NCPRegions []string `json:"NCPRegions"`
}

type BasicResponse struct {
	Result string  `json:"Result"`
	Error  *string `json:"Error"`
}

type DiagnoseResponse struct {
	Result      string                  `json:"Result"`
	Diagnostics diagnostics.TimedResult `json:"Diagnostics,omitempty"`
	Error       *string                 `json:"Error"`
}

type SysbenchResponse struct {
	Result         string                  `json:"Result"`
	SysbenchResult sysbench.SysbenchParsed `json:"SysbenchResult,omitempty"`
	Error          *string                 `json:"Error"`
}
