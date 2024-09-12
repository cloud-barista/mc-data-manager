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
package execfunc

import (
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/dummy/semistructured"
	"github.com/cloud-barista/mc-data-manager/pkg/dummy/structured"
	"github.com/cloud-barista/mc-data-manager/pkg/dummy/unstructured"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
)

func DummyCreate(params models.CommandTask) error {
	log.Info().Msgf("check directory DummyPaths")
	if cast.ToInt(params.SizeSQL) != 0 {
		log.Info().Msgf("start sql generation")
		if err := structured.GenerateRandomSQL(params.DummyPath, cast.ToInt(params.SizeSQL)); err != nil {
			log.Error().Msgf("failed to generate sql")
			return err
		}
		log.Info().Msgf("successfully generated sql : %s", params.DummyPath)
	}
	log.Info().Msgf("start Serversql generation Boolean? :", (cast.ToInt(params.SizeServerSQL) != 0))
	if cast.ToInt(params.SizeServerSQL) != 0 {
		log.Info().Msgf("start Serversql generation")
		if err := structured.GenerateRandomSQLWithServer(params.DummyPath, cast.ToInt(params.SizeServerSQL)); err != nil {
			log.Error().Msgf("failed to generate sql")
			return err
		}
		log.Info().Msgf("successfully generated sql : %s", params.DummyPath)
	}

	if cast.ToInt(params.SizeCSV) != 0 {
		log.Info().Msgf("start csv generation")
		if err := structured.GenerateRandomCSV(params.DummyPath, cast.ToInt(params.SizeCSV)); err != nil {
			log.Error().Msgf("failed to generate csv")
			return err
		}
		log.Info().Msgf("successfully generated csv : %s", params.DummyPath)
	}

	if cast.ToInt(params.SizeJSON) != 0 {
		log.Info().Msgf("start json generation")
		if err := semistructured.GenerateRandomJSON(params.DummyPath, cast.ToInt(params.SizeJSON)); err != nil {
			log.Error().Msgf("failed to generate json")
			return err
		}
		log.Info().Msgf("successfully generated json : %s", params.DummyPath)
	}

	if cast.ToInt(params.SizeXML) != 0 {
		log.Info().Msgf("start xml generation")
		if err := semistructured.GenerateRandomXML(params.DummyPath, cast.ToInt(params.SizeXML)); err != nil {
			log.Error().Msgf("failed to generate xml")
			return err
		}
		log.Info().Msgf("successfully generated xml : %s", params.DummyPath)
	}

	if cast.ToInt(params.SizeTXT) != 0 {
		log.Info().Msgf("start txt generation")
		if err := unstructured.GenerateRandomTXT(params.DummyPath, cast.ToInt(params.SizeTXT)); err != nil {
			log.Error().Msgf("failed to generate txt")
			return err
		}
		log.Info().Msgf("successfully generated txt : %s", params.DummyPath)
	}

	if cast.ToInt(params.SizePNG) != 0 {
		log.Info().Msgf("start png generation")
		if err := unstructured.GenerateRandomPNGImage(params.DummyPath, cast.ToInt(params.SizePNG)); err != nil {
			log.Error().Msgf("failed to generate png")
			return err
		}
		log.Info().Msgf("successfully generated png : %s", params.DummyPath)
	}

	if cast.ToInt(params.SizeGIF) != 0 {
		log.Info().Msgf("start gif generation")
		if err := unstructured.GenerateRandomGIF(params.DummyPath, cast.ToInt(params.SizeGIF)); err != nil {
			log.Error().Msgf("failed to generate gif")
			return err
		}
		log.Info().Msgf("successfully generated gif : %s", params.DummyPath)
	}

	if cast.ToInt(params.SizeZIP) != 0 {
		log.Info().Msgf("start zip generation")
		if err := unstructured.GenerateRandomZIP(params.DummyPath, cast.ToInt(params.SizeZIP)); err != nil {
			log.Error().Msgf("failed to generate zip")
			return err
		}
		log.Info().Msgf("successfully generated zip : %s", params.DummyPath)
	}
	return nil
}
