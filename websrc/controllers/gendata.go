package controllers

import (
	"mime/multipart"
	"strconv"

	"github.com/cloud-barista/cm-data-mold/pkg/dummy/semistructed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/structed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/unstructed"
)

type GenDataParams struct {
	Region    string `json:"region" form:"region"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Bucket    string `json:"bucket" form:"bucket"`
	Endpoint  string `json:"endpoint"`
	DummyPath string `json:"path"`

	CheckSQL  string `json:"checkSQL" form:"checkSQL"`
	CheckCSV  string `json:"checkCSV" form:"checkCSV"`
	CheckTXT  string `json:"checkTXT" form:"checkTXT"`
	CheckPNG  string `json:"checkPNG" form:"checkPNG"`
	CheckGIF  string `json:"checkGIF" form:"checkGIF"`
	CheckZIP  string `json:"checkZIP" form:"checkZIP"`
	CheckJSON string `json:"checkJSON" form:"checkJSON"`
	CheckXML  string `json:"checkXML" form:"checkXML"`

	SizeSQL  string `json:"sizeSQL" form:"sizeSQL"`
	SizeCSV  string `json:"sizeCSV" form:"sizeCSV"`
	SizeTXT  string `json:"sizeTXT" form:"sizeTXT"`
	SizePNG  string `json:"sizePNG" form:"sizePNG"`
	SizeGIF  string `json:"sizeGIF" form:"sizeGIF"`
	SizeZIP  string `json:"sizeZIP" form:"sizeZIP"`
	SizeJSON string `json:"sizeJSON" form:"sizeJSON"`
	SizeXML  string `json:"sizeXML" form:"sizeXML"`

	DBProvider   string `json:"provider"`
	DBHost       string `json:"host"`
	DBPort       string `json:"port"`
	DBUser       string `json:"username"`
	DBPassword   string `json:"password"`
	DatabaseName string `json:"databaseName"`

	GCSCredential *multipart.FileHeader `form:"gcsCredential"`
	ProjectID     string                `form:"projectid"`
}

func genData(params GenDataParams) error {
	if params.CheckSQL == "on" {
		sql, _ := strconv.Atoi(params.SizeSQL)
		if err := structed.GenerateRandomSQL(params.DummyPath, sql); err != nil {
			return err
		}
	}

	if params.CheckCSV == "on" {
		csv, _ := strconv.Atoi(params.SizeCSV)
		if err := structed.GenerateRandomCSV(params.DummyPath, csv); err != nil {
			return err
		}
	}

	if params.CheckTXT == "on" {
		txt, _ := strconv.Atoi(params.SizeTXT)
		if err := unstructed.GenerateRandomTXT(params.DummyPath, txt); err != nil {
			return err
		}
	}

	if params.CheckPNG == "on" {
		png, _ := strconv.Atoi(params.SizePNG)
		if err := unstructed.GenerateRandomPNGImage(params.DummyPath, png); err != nil {
			return err
		}
	}

	if params.CheckGIF == "on" {
		gif, _ := strconv.Atoi(params.SizeGIF)
		if err := unstructed.GenerateRandomGIF(params.DummyPath, gif); err != nil {
			return err
		}
	}

	if params.CheckZIP == "on" {
		zip, _ := strconv.Atoi(params.SizeZIP)
		if err := unstructed.GenerateRandomZIP(params.DummyPath, zip); err != nil {
			return err
		}
	}

	if params.CheckJSON == "on" {
		json, _ := strconv.Atoi(params.SizeJSON)
		if err := semistructed.GenerateRandomJSON(params.DummyPath, json); err != nil {
			return err
		}
	}

	if params.CheckXML == "on" {
		xml, _ := strconv.Atoi(params.SizeXML)
		if err := semistructed.GenerateRandomXML(params.DummyPath, xml); err != nil {
			return err
		}
	}

	return nil
}
