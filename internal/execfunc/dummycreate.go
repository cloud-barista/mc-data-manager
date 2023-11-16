package execfunc

import (
	"github.com/cloud-barista/cm-data-mold/internal/auth"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/semistructed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/structed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/unstructed"
	"github.com/sirupsen/logrus"
)

func DummyCreate(datamoldParams auth.DatamoldParams) error {
	logrus.Info("check directory paths")
	if datamoldParams.SqlSize != 0 {
		logrus.Info("start sql generation")
		if err := structed.GenerateRandomSQL(datamoldParams.DstPath, datamoldParams.SqlSize); err != nil {
			logrus.Error("failed to generate sql")
			return err
		}
		logrus.Infof("successfully generated sql : %s", datamoldParams.DstPath)
	}

	if datamoldParams.CsvSize != 0 {
		logrus.Info("start csv generation")
		if err := structed.GenerateRandomCSV(datamoldParams.DstPath, datamoldParams.CsvSize); err != nil {
			logrus.Error("failed to generate csv")
			return err
		}
		logrus.Infof("successfully generated csv : %s", datamoldParams.DstPath)
	}

	if datamoldParams.JsonSize != 0 {
		logrus.Info("start json generation")
		if err := semistructed.GenerateRandomJSON(datamoldParams.DstPath, datamoldParams.JsonSize); err != nil {
			logrus.Error("failed to generate json")
			return err
		}
		logrus.Infof("successfully generated json : %s", datamoldParams.DstPath)
	}

	if datamoldParams.XmlSize != 0 {
		logrus.Info("start xml generation")
		if err := semistructed.GenerateRandomXML(datamoldParams.DstPath, datamoldParams.XmlSize); err != nil {
			logrus.Error("failed to generate xml")
			return err
		}
		logrus.Infof("successfully generated xml : %s", datamoldParams.DstPath)
	}

	if datamoldParams.TxtSize != 0 {
		logrus.Info("start txt generation")
		if err := unstructed.GenerateRandomTXT(datamoldParams.DstPath, datamoldParams.TxtSize); err != nil {
			logrus.Error("failed to generate txt")
			return err
		}
		logrus.Infof("successfully generated txt : %s", datamoldParams.DstPath)
	}

	if datamoldParams.PngSize != 0 {
		logrus.Info("start png generation")
		if err := unstructed.GenerateRandomPNGImage(datamoldParams.DstPath, datamoldParams.PngSize); err != nil {
			logrus.Error("failed to generate png")
			return err
		}
		logrus.Infof("successfully generated png : %s", datamoldParams.DstPath)
	}

	if datamoldParams.GifSize != 0 {
		logrus.Info("start gif generation")
		if err := unstructed.GenerateRandomGIF(datamoldParams.DstPath, datamoldParams.GifSize); err != nil {
			logrus.Error("failed to generate gif")
			return err
		}
		logrus.Infof("successfully generated gif : %s", datamoldParams.DstPath)
	}

	if datamoldParams.ZipSize != 0 {
		logrus.Info("start zip generation")
		if err := unstructed.GenerateRandomZIP(datamoldParams.DstPath, datamoldParams.ZipSize); err != nil {
			logrus.Error("failed to generate zip")
			return err
		}
		logrus.Infof("successfully generated zip : %s", datamoldParams.DstPath)
	}
	return nil
}
