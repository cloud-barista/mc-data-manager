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
	"github.com/cloud-barista/cm-data-mold/internal/auth"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/semistructured"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/structured"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/unstructured"
	"github.com/sirupsen/logrus"
)

func DummyCreate(datamoldParams auth.DatamoldParams) error {
	logrus.Info("check directory paths")
	if datamoldParams.SqlSize != 0 {
		logrus.Info("start sql generation")
		if err := structured.GenerateRandomSQL(datamoldParams.DstPath, datamoldParams.SqlSize); err != nil {
			logrus.Error("failed to generate sql")
			return err
		}
		logrus.Infof("successfully generated sql : %s", datamoldParams.DstPath)
	}

	if datamoldParams.CsvSize != 0 {
		logrus.Info("start csv generation")
		if err := structured.GenerateRandomCSV(datamoldParams.DstPath, datamoldParams.CsvSize); err != nil {
			logrus.Error("failed to generate csv")
			return err
		}
		logrus.Infof("successfully generated csv : %s", datamoldParams.DstPath)
	}

	if datamoldParams.JsonSize != 0 {
		logrus.Info("start json generation")
		if err := semistructured.GenerateRandomJSON(datamoldParams.DstPath, datamoldParams.JsonSize); err != nil {
			logrus.Error("failed to generate json")
			return err
		}
		logrus.Infof("successfully generated json : %s", datamoldParams.DstPath)
	}

	if datamoldParams.XmlSize != 0 {
		logrus.Info("start xml generation")
		if err := semistructured.GenerateRandomXML(datamoldParams.DstPath, datamoldParams.XmlSize); err != nil {
			logrus.Error("failed to generate xml")
			return err
		}
		logrus.Infof("successfully generated xml : %s", datamoldParams.DstPath)
	}

	if datamoldParams.TxtSize != 0 {
		logrus.Info("start txt generation")
		if err := unstructured.GenerateRandomTXT(datamoldParams.DstPath, datamoldParams.TxtSize); err != nil {
			logrus.Error("failed to generate txt")
			return err
		}
		logrus.Infof("successfully generated txt : %s", datamoldParams.DstPath)
	}

	if datamoldParams.PngSize != 0 {
		logrus.Info("start png generation")
		if err := unstructured.GenerateRandomPNGImage(datamoldParams.DstPath, datamoldParams.PngSize); err != nil {
			logrus.Error("failed to generate png")
			return err
		}
		logrus.Infof("successfully generated png : %s", datamoldParams.DstPath)
	}

	if datamoldParams.GifSize != 0 {
		logrus.Info("start gif generation")
		if err := unstructured.GenerateRandomGIF(datamoldParams.DstPath, datamoldParams.GifSize); err != nil {
			logrus.Error("failed to generate gif")
			return err
		}
		logrus.Infof("successfully generated gif : %s", datamoldParams.DstPath)
	}

	if datamoldParams.ZipSize != 0 {
		logrus.Info("start zip generation")
		if err := unstructured.GenerateRandomZIP(datamoldParams.DstPath, datamoldParams.ZipSize); err != nil {
			logrus.Error("failed to generate zip")
			return err
		}
		logrus.Infof("successfully generated zip : %s", datamoldParams.DstPath)
	}
	return nil
}
