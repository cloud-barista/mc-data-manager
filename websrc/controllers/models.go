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

import (
	"github.com/cloud-barista/mc-data-manager/models"
)

type (
	// Gen
	GenDataParams      = models.GenDataParams
	GenMySQLParams     = models.GenMySQLParams
	GenFirestoreParams = models.FirestoreParams
	// Mig
	MigrationMySQLForm   = models.MigrationMySQLForm
	MigrationMySQLParams = models.MigrationMySQLParams
	MySQLParams          = models.MySQLParams
	MigrationForm        = models.MigrationForm
	LinuxMigrationParams = models.LinuxMigrationParams
	NCPMigrationParams   = models.NCPMigrationParams
	AWSMigrationParams   = models.AWSMigrationParams
	GCPMigrationParams   = models.GCPMigrationParams
	MongoMigrationParams = models.MongoMigrationParams
)
