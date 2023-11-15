package controllers

import (
	"mime/multipart"
	"strconv"

	"github.com/cloud-barista/cm-data-mold/pkg/dummy/semistructed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/structed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/unstructed"
	"github.com/sirupsen/logrus"
)

type GenDataParams struct {
	Region    string `json:"region" form:"region"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Bucket    string `json:"bucket" form:"bucket"`
	Endpoint  string `json:"endpoint"`
	DummyPath string `json:"path"`

	CheckSQL        string `json:"checkSQL" form:"checkSQL"`
	CheckCSV        string `json:"checkCSV" form:"checkCSV"`
	CheckTXT        string `json:"checkTXT" form:"checkTXT"`
	CheckPNG        string `json:"checkPNG" form:"checkPNG"`
	CheckGIF        string `json:"checkGIF" form:"checkGIF"`
	CheckZIP        string `json:"checkZIP" form:"checkZIP"`
	CheckJSON       string `json:"checkJSON" form:"checkJSON"`
	CheckXML        string `json:"checkXML" form:"checkXML"`
	CheckServerJSON string
	CheckServerSQL  string

	SizeSQL        string `json:"sizeSQL" form:"sizeSQL"`
	SizeCSV        string `json:"sizeCSV" form:"sizeCSV"`
	SizeTXT        string `json:"sizeTXT" form:"sizeTXT"`
	SizePNG        string `json:"sizePNG" form:"sizePNG"`
	SizeGIF        string `json:"sizeGIF" form:"sizeGIF"`
	SizeZIP        string `json:"sizeZIP" form:"sizeZIP"`
	SizeJSON       string `json:"sizeJSON" form:"sizeJSON"`
	SizeXML        string `json:"sizeXML" form:"sizeXML"`
	SizeServerJSON string
	SizeServerSQL  string

	DBProvider   string `json:"provider"`
	DBHost       string `json:"host"`
	DBPort       string `json:"port"`
	DBUser       string `json:"username"`
	DBPassword   string `json:"password"`
	DatabaseName string `json:"databaseName"`

	GCSCredential *multipart.FileHeader `form:"gcpCredential"`
	ProjectID     string                `form:"projectid"`
}

func genData(params GenDataParams, logger *logrus.Logger) error {
	if params.CheckSQL == "on" {
		logger.Info("Start creating sql dummy")
		sql, _ := strconv.Atoi(params.SizeSQL)
		if err := structed.GenerateRandomSQL(params.DummyPath, sql); err != nil {
			logger.Info("Failed to create sql dummy")
			return err
		}
		logger.Info("Successfully generated sql dummy")
	}

	if params.CheckCSV == "on" {
		logger.Info("Start creating csv dummy")
		csv, _ := strconv.Atoi(params.SizeCSV)
		if err := structed.GenerateRandomCSV(params.DummyPath, csv); err != nil {
			logger.Info("Failed to create csv dummy")
			return err
		}
		logger.Info("Successfully generated csv dummy")
	}

	if params.CheckTXT == "on" {
		logger.Info("Start creating txt dummy")
		txt, _ := strconv.Atoi(params.SizeTXT)
		if err := unstructed.GenerateRandomTXT(params.DummyPath, txt); err != nil {
			logger.Info("Failed to create txt dummy")
			return err
		}
		logger.Info("Successfully generated txt dummy")
	}

	if params.CheckPNG == "on" {
		logger.Info("Start creating png dummy")
		png, _ := strconv.Atoi(params.SizePNG)
		if err := unstructed.GenerateRandomPNGImage(params.DummyPath, png); err != nil {
			logger.Info("Failed to create png dummy")
			return err
		}
		logger.Info("Successfully generated png dummy")
	}

	if params.CheckGIF == "on" {
		logger.Info("Start creating gif dummy")
		gif, _ := strconv.Atoi(params.SizeGIF)
		if err := unstructed.GenerateRandomGIF(params.DummyPath, gif); err != nil {
			logger.Info("Failed to create gif dummy")
			return err
		}
		logger.Info("Successfully generated gif dummy")
	}

	if params.CheckZIP == "on" {
		logger.Info("Start creating a pile of zip files that compressed txt")
		zip, _ := strconv.Atoi(params.SizeZIP)
		if err := unstructed.GenerateRandomZIP(params.DummyPath, zip); err != nil {
			logger.Info("Failed to create zip file dummy compressed txt")
			return err
		}
		logger.Info("Successfully created zip file dummy compressed txt")
	}

	if params.CheckJSON == "on" {
		logger.Info("Start creating json dummy")
		json, _ := strconv.Atoi(params.SizeJSON)
		if err := semistructed.GenerateRandomJSON(params.DummyPath, json); err != nil {
			logger.Info("Failed to create json dummy")
			return err
		}
		logger.Info("Successfully generated json dummy")
	}

	if params.CheckXML == "on" {
		logger.Info("Start creating xml dummy")
		xml, _ := strconv.Atoi(params.SizeXML)
		if err := semistructed.GenerateRandomXML(params.DummyPath, xml); err != nil {
			logger.Info("Failed to create xml dummy")
			return err
		}
		logger.Info("Successfully generated xml dummy")
	}

	if params.CheckServerJSON == "on" {
		logger.Info("Start creating json dummy")
		json, _ := strconv.Atoi(params.SizeServerJSON)
		if err := semistructed.GenerateRandomJSONWithServer(params.DummyPath, json); err != nil {
			logger.Info("Failed to create json dummy")
			return err
		}
		logger.Info("Successfully generated json dummy")
	}

	if params.CheckServerSQL == "on" {
		logger.Info("Start creating sql dummy")
		sql, _ := strconv.Atoi(params.SizeServerSQL)
		if err := structed.GenerateRandomSQLWithServer(params.DummyPath, sql); err != nil {
			logger.Info("Failed to create sql dummy")
			return err
		}
		logger.Info("Successfully generated sql dummy")
	}

	return nil
}
