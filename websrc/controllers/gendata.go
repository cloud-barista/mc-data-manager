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
	"mime/multipart"
	"strconv"

	"github.com/cloud-barista/mc-data-manager/pkg/dummy/semistructured"
	"github.com/cloud-barista/mc-data-manager/pkg/dummy/structured"
	"github.com/cloud-barista/mc-data-manager/pkg/dummy/unstructured"
	"github.com/sirupsen/logrus"
)

type GenDataParams struct {
	Region    string `json:"region" form:"region"`
	AccessKey string `json:"accessKey" form:"accessKey"`
	SecretKey string `json:"secretKey" form:"secretKey"`
	Bucket    string `json:"bucket" form:"bucket"`
	Endpoint  string `json:"endpoint" form:"endpoint"`
	DummyPath string `json:"path" form:"path"`

	CheckSQL        string `json:"checkSQL" form:"checkSQL"`
	CheckCSV        string `json:"checkCSV" form:"checkCSV"`
	CheckTXT        string `json:"checkTXT" form:"checkTXT"`
	CheckPNG        string `json:"checkPNG" form:"checkPNG"`
	CheckGIF        string `json:"checkGIF" form:"checkGIF"`
	CheckZIP        string `json:"checkZIP" form:"checkZIP"`
	CheckJSON       string `json:"checkJSON" form:"checkJSON"`
	CheckXML        string `json:"checkXML" form:"checkXML"`
	CheckServerJSON string `json:"checkServerJSON" form:"checkServerJSON"`
	CheckServerSQL  string `json:"checkServerSQL" form:"checkServerSQL"`

	SizeSQL        string `json:"sizeSQL" form:"sizeSQL"`
	SizeCSV        string `json:"sizeCSV" form:"sizeCSV"`
	SizeTXT        string `json:"sizeTXT" form:"sizeTXT"`
	SizePNG        string `json:"sizePNG" form:"sizePNG"`
	SizeGIF        string `json:"sizeGIF" form:"sizeGIF"`
	SizeZIP        string `json:"sizeZIP" form:"sizeZIP"`
	SizeJSON       string `json:"sizeJSON" form:"sizeJSON"`
	SizeXML        string `json:"sizeXML" form:"sizeXML"`
	SizeServerJSON string `json:"sizeServerJSON" form:"sizeServerJSON"`
	SizeServerSQL  string `json:"sizeServerSQL" form:"sizeServerSQL"`

	DBProvider   string `json:"provider" form:"provider"`
	DBHost       string `json:"host" form:"host"`
	DBPort       string `json:"port" form:"port"`
	DBUser       string `json:"username" form:"username"`
	DBPassword   string `json:"password" form:"password"`
	DatabaseName string `json:"databaseName" form:"databaseName"`

	GCPCredential *multipart.FileHeader `form:"gcpCredential" swaggerignore:"true"`
	ProjectID     string                `json:"projectId" form:"projectid"`
}

func genData(params GenDataParams, logger *logrus.Logger) error {
	if params.CheckSQL == "on" {
		logger.Info("Start creating sql dummy")
		sql, _ := strconv.Atoi(params.SizeSQL)
		if err := structured.GenerateRandomSQL(params.DummyPath, sql); err != nil {
			logger.Info("Failed to create sql dummy")
			return err
		}
		logger.Info("Successfully generated sql dummy")
	}

	if params.CheckCSV == "on" {
		logger.Info("Start creating csv dummy")
		csv, _ := strconv.Atoi(params.SizeCSV)
		if err := structured.GenerateRandomCSV(params.DummyPath, csv); err != nil {
			logger.Info("Failed to create csv dummy")
			return err
		}
		logger.Info("Successfully generated csv dummy")
	}

	if params.CheckTXT == "on" {
		logger.Info("Start creating txt dummy")
		txt, _ := strconv.Atoi(params.SizeTXT)
		if err := unstructured.GenerateRandomTXT(params.DummyPath, txt); err != nil {
			logger.Info("Failed to create txt dummy")
			return err
		}
		logger.Info("Successfully generated txt dummy")
	}

	if params.CheckPNG == "on" {
		logger.Info("Start creating png dummy")
		png, _ := strconv.Atoi(params.SizePNG)
		if err := unstructured.GenerateRandomPNGImage(params.DummyPath, png); err != nil {
			logger.Info("Failed to create png dummy")
			return err
		}
		logger.Info("Successfully generated png dummy")
	}

	if params.CheckGIF == "on" {
		logger.Info("Start creating gif dummy")
		gif, _ := strconv.Atoi(params.SizeGIF)
		if err := unstructured.GenerateRandomGIF(params.DummyPath, gif); err != nil {
			logger.Info("Failed to create gif dummy")
			return err
		}
		logger.Info("Successfully generated gif dummy")
	}

	if params.CheckZIP == "on" {
		logger.Info("Start creating a pile of zip files that compressed txt")
		zip, _ := strconv.Atoi(params.SizeZIP)
		if err := unstructured.GenerateRandomZIP(params.DummyPath, zip); err != nil {
			logger.Info("Failed to create zip file dummy compressed txt")
			return err
		}
		logger.Info("Successfully created zip file dummy compressed txt")
	}

	if params.CheckJSON == "on" {
		logger.Info("Start creating json dummy")
		json, _ := strconv.Atoi(params.SizeJSON)
		if err := semistructured.GenerateRandomJSON(params.DummyPath, json); err != nil {
			logger.Info("Failed to create json dummy")
			return err
		}
		logger.Info("Successfully generated json dummy")
	}

	if params.CheckXML == "on" {
		logger.Info("Start creating xml dummy")
		xml, _ := strconv.Atoi(params.SizeXML)
		if err := semistructured.GenerateRandomXML(params.DummyPath, xml); err != nil {
			logger.Info("Failed to create xml dummy")
			return err
		}
		logger.Info("Successfully generated xml dummy")
	}

	if params.CheckServerJSON == "on" {
		logger.Info("Start creating json dummy")
		json, _ := strconv.Atoi(params.SizeServerJSON)
		if err := semistructured.GenerateRandomJSONWithServer(params.DummyPath, json); err != nil {
			logger.Info("Failed to create json dummy")
			return err
		}
		logger.Info("Successfully generated json dummy")
	}

	if params.CheckServerSQL == "on" {
		logger.Info("Start creating sql dummy")
		sql, _ := strconv.Atoi(params.SizeServerSQL)
		if err := structured.GenerateRandomSQLWithServer(params.DummyPath, sql); err != nil {
			logger.Info("Failed to create sql dummy")
			return err
		}
		logger.Info("Successfully generated sql dummy")
	}

	return nil
}
