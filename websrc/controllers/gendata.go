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
	"errors"
	"reflect"
	"strconv"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/dummy/semistructured"
	"github.com/cloud-barista/mc-data-manager/pkg/dummy/structured"
	"github.com/cloud-barista/mc-data-manager/pkg/dummy/unstructured"
	"github.com/rs/zerolog"
	"github.com/spf13/cast"
)

func genData(params models.GenFileParams, logger *zerolog.Logger) error {

	if !hasAnyTrue(params.FileFormatParams) {
		err := errors.New("no file format selected")
		logger.Info().Msgf("%+v", params)
		logger.Error().Err(err).Msg("At least one file format must be selected")
		return err
	}
	if cast.ToBool(params.CheckSQL) {
		logger.Info().Msg("Start creating SQL dummy")
		sql, _ := strconv.Atoi(params.SizeSQL)
		if err := structured.GenerateRandomSQL(params.DummyPath, sql); err != nil {
			logger.Error().Err(err).Msg("Failed to create SQL dummy")
			return err
		}
		logger.Info().Msg("Successfully generated SQL dummy")
	}

	if cast.ToBool(params.CheckCSV) {
		logger.Info().Msg("Start creating CSV dummy")
		csv, _ := strconv.Atoi(params.SizeCSV)
		if err := structured.GenerateRandomCSV(params.DummyPath, csv); err != nil {
			logger.Error().Err(err).Msg("Failed to create CSV dummy")
			return err
		}
		logger.Info().Msg("Successfully generated CSV dummy")
	}

	if cast.ToBool(params.CheckTXT) {
		logger.Info().Msg("Start creating TXT dummy")
		txt, _ := strconv.Atoi(params.SizeTXT)
		if err := unstructured.GenerateRandomTXT(params.DummyPath, txt); err != nil {
			logger.Error().Err(err).Msg("Failed to create TXT dummy")
			return err
		}
		logger.Info().Msg("Successfully generated TXT dummy")
	}

	if cast.ToBool(params.CheckPNG) {
		logger.Info().Msg("Start creating PNG dummy")
		png, _ := strconv.Atoi(params.SizePNG)
		if err := unstructured.GenerateRandomPNGImage(params.DummyPath, png); err != nil {
			logger.Error().Err(err).Msg("Failed to create PNG dummy")
			return err
		}
		logger.Info().Msg("Successfully generated PNG dummy")
	}

	if cast.ToBool(params.CheckGIF) {
		logger.Info().Msg("Start creating GIF dummy")
		gif, _ := strconv.Atoi(params.SizeGIF)
		if err := unstructured.GenerateRandomGIF(params.DummyPath, gif); err != nil {
			logger.Error().Err(err).Msg("Failed to create GIF dummy")
			return err
		}
		logger.Info().Msg("Successfully generated GIF dummy")
	}

	if cast.ToBool(params.CheckZIP) {
		logger.Info().Msg("Start creating a pile of ZIP files that compress TXT")
		zip, _ := strconv.Atoi(params.SizeZIP)
		if err := unstructured.GenerateRandomZIP(params.DummyPath, zip); err != nil {
			logger.Error().Err(err).Msg("Failed to create ZIP file dummy compressed TXT")
			return err
		}
		logger.Info().Msg("Successfully created ZIP file dummy compressed TXT")
	}

	if cast.ToBool(params.CheckJSON) {
		logger.Info().Msg("Start creating JSON dummy")
		json, _ := strconv.Atoi(params.SizeJSON)
		if err := semistructured.GenerateRandomJSON(params.DummyPath, json); err != nil {
			logger.Error().Err(err).Msg("Failed to create JSON dummy")
			return err
		}
		logger.Info().Msg("Successfully generated JSON dummy")
	}

	if cast.ToBool(params.CheckXML) {
		logger.Info().Msg("Start creating XML dummy")
		xml, _ := strconv.Atoi(params.SizeXML)
		if err := semistructured.GenerateRandomXML(params.DummyPath, xml); err != nil {
			logger.Error().Err(err).Msg("Failed to create XML dummy")
			return err
		}
		logger.Info().Msg("Successfully generated XML dummy")
	}

	if cast.ToBool(params.CheckServerJSON) {
		logger.Info().Msg("Start creating JSON dummy")
		json, _ := strconv.Atoi(params.SizeServerJSON)
		if err := semistructured.GenerateRandomJSONWithServer(params.DummyPath, json); err != nil {
			logger.Error().Err(err).Msg("Failed to create JSON dummy")
			return err
		}
		logger.Info().Msg("Successfully generated JSON dummy")
	}

	if cast.ToBool(params.CheckServerSQL) {
		logger.Info().Msg("Start creating SQL dummy")
		sql, _ := strconv.Atoi(params.SizeServerSQL)
		if err := structured.GenerateRandomSQLWithServer(params.DummyPath, sql); err != nil {
			logger.Error().Err(err).Msg("Failed to create SQL dummy")
			return err
		}
		logger.Info().Msg("Successfully generated SQL dummy")
	}

	return nil
}
func hasAnyTrue(fileFormatParams models.FileFormatParams) bool {
	v := reflect.ValueOf(fileFormatParams)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Bool && field.Bool() {
			return true
		}
	}
	return false
}
