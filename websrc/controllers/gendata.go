package controllers

import (
	"strconv"

	"github.com/cloud-barista/cm-data-mold/pkg/dummy/semistructed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/structed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/unstructed"
)

type GenDataParams struct {
	Region    string `json:"region"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Bucket    string `json:"bucket"`
	Endpoint  string `json:"endpoint"`
	DummyPath string `json:"path"`

	CheckSQL  string `json:"checkSQL"`
	CheckCSV  string `json:"checkCSV"`
	CheckTXT  string `json:"checkTXT"`
	CheckPNG  string `json:"checkPNG"`
	CheckGIF  string `json:"checkGIF"`
	CheckZIP  string `json:"checkZIP"`
	CheckJSON string `json:"checkJSON"`
	CheckXML  string `json:"checkXML"`

	SizeSQL  string `json:"sizeSQL"`
	SizeCSV  string `json:"sizeCSV"`
	SizeTXT  string `json:"sizeTXT"`
	SizePNG  string `json:"sizePNG"`
	SizeGIF  string `json:"sizeGIF"`
	SizeZIP  string `json:"sizeZIP"`
	SizeJSON string `json:"sizeJSON"`
	SizeXML  string `json:"sizeXML"`

	DBProvider   string `json:"provider"`
	DBHost       string `json:"host"`
	DBPort       string `json:"port"`
	DBUser       string `json:"username"`
	DBPassword   string `json:"password"`
	DatabaseName string `json:"databaseName"`
}

type CloudParams struct {
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
