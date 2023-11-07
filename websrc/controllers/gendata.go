package controllers

import (
	"strconv"

	"github.com/cloud-barista/cm-data-mold/pkg/dummy/semistructed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/structed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/unstructed"
)

func genData(dummyPath, checkSQL, sizeSQL, checkCSV, sizeCSV, checkTXT, sizeTXT, checkPNG, sizePNG, checkGIF, sizeGIF, checkZIP, sizeZIP, checkJSON, sizeJSON, checkXML, sizeXML string) error {
	if checkSQL == "sql" {
		sql, _ := strconv.Atoi(sizeSQL)
		if err := structed.GenerateRandomSQL(dummyPath, sql); err != nil {
			return err
		}
	}

	if checkCSV == "csv" {
		csv, _ := strconv.Atoi(sizeCSV)
		if err := structed.GenerateRandomCSV(dummyPath, csv); err != nil {
			return err
		}
	}

	if checkTXT == "txt" {
		txt, _ := strconv.Atoi(sizeTXT)
		if err := unstructed.GenerateRandomTXT(dummyPath, txt); err != nil {
			return err
		}
	}

	if checkPNG == "png" {
		png, _ := strconv.Atoi(sizePNG)
		if err := unstructed.GenerateRandomPNGImage(dummyPath, png); err != nil {
			return err
		}
	}

	if checkGIF == "gif" {
		gif, _ := strconv.Atoi(sizeGIF)
		if err := unstructed.GenerateRandomGIF(dummyPath, gif); err != nil {
			return err
		}
	}

	if checkZIP == "zip" {
		zip, _ := strconv.Atoi(sizeZIP)
		if err := unstructed.GenerateRandomZIP(dummyPath, zip); err != nil {
			return err
		}
	}

	if checkJSON == "json" {
		json, _ := strconv.Atoi(sizeJSON)
		if err := semistructed.GenerateRandomJSON(dummyPath, json); err != nil {
			return err
		}
	}

	if checkXML == "xml" {
		xml, _ := strconv.Atoi(sizeXML)
		if err := semistructed.GenerateRandomXML(dummyPath, xml); err != nil {
			return err
		}
	}

	return nil
}
