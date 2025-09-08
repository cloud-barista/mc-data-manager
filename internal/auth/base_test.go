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
package auth

import (
	"os"
	"testing"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	service "github.com/cloud-barista/mc-data-manager/service/credential"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	dsn := "mcmp:mcmp@tcp(127.0.0.1:3301)/mcmp?parseTime=true"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	os.Setenv("ENCODING_SECRET_KEY", "12345678901234567890123456789012")

	// 전역 DB 세팅
	config.DB = db

	// CredentialService를 FileProfileManager에 연결
	csvc := service.NewCredentialService(db)
	csvc.AesConverter = utils.NewAESConverter()

	config.DefaultProfileManager = &config.FileProfileManager{
		CredentialService: csvc,
	}

	os.Exit(m.Run())
}

func TestGetOS(t *testing.T) {
	if config.DefaultProfileManager == nil {
		t.Fatal("DefaultProfileManager is nil")
	}

	pc := models.ProviderConfig{
		BaseParams: models.BaseParams{
			ProviderParams:   models.ProviderParams{Provider: "aws"},
			RegionParams:     models.RegionParams{Region: "ap-northeast-2"},
			CredentialParams: models.CredentialParams{CredentialId: 1},
		},
	}

	credMgr := config.DefaultProfileManager
	creds, err := credMgr.LoadCredentialsById(uint64(pc.CredentialId), pc.Provider)
	if err != nil {
		t.Fatalf("LoadCredentialsById error: %v", err)
	}

	t.Logf("creds type=%T value=%+v", creds, creds)
}
