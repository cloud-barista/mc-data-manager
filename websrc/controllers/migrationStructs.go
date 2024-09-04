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

// MigrationForm represents the form data required for migration processes.
// @Description MigrationForm contains all the necessary fields for migrating data between different services.

func GetMigrationParamsFormFormData(form MigrationMySQLForm) MigrationMySQLParams {
	source := MySQLParams{
		Host:         form.SHost,
		Port:         form.SPort,
		User:         form.SUsername,
		Password:     form.SPassword,
		DatabaseName: form.SDatabaseName,
	}
	target := MySQLParams{
		Host:         form.DProvider,
		Port:         form.DPort,
		User:         form.DUsername,
		Password:     form.DPassword,
		DatabaseName: form.DDatabaseName,
	}
	return MigrationMySQLParams{SourcePoint: source, TargetPoint: target}
}
