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
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/service/osc"
	"github.com/rs/zerolog/log"
)

func ImportOSFunc(params *models.CommandTask) error {
	var OSC *osc.OSController
	var err error
	log.Info().Msgf("User Information")
	OSC, err = GetOS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("OSController error importing into objectstorage : %v", err)
		return err
	}

	log.Info().Msgf("Launch OSController MPut")
	if err := OSC.MPut(params.Directory); err != nil {
		log.Error().Msgf("MPut error importing into objectstorage")
		log.Info().Msgf("params : %+v", params.TargetPoint)

		return err
	}
	log.Info().Msgf("successfully imported : %s", params.Directory)
	return nil
}

func ExportOSFunc(params *models.CommandTask) error {
	var OSC *osc.OSController
	var err error
	log.Info().Msgf("User Information")
	OSC, err = GetOS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("OSController error importing into objectstorage : %v", err)
		return err
	}

	log.Info().Msgf("Launch OSController MGet")
	if err := OSC.MGet(params.Directory); err != nil {
		log.Error().Msgf("MGet error exporting into objectstorage : %v", err)
		return err
	}
	log.Info().Msgf("successfully exported : %s", params.Directory)
	return nil
}

func MigrationOSFunc(params *models.CommandTask) error {
	var src *osc.OSController
	var srcErr error
	var dst *osc.OSController
	var dstErr error
	log.Info().Msgf("Source Information")
	src, srcErr = GetOS(&params.SourcePoint)
	if srcErr != nil {
		log.Error().Msgf("OSController error migration into objectstorage : %v", srcErr)
		return srcErr
	}
	log.Info().Msgf("Target Information")
	dst, dstErr = GetOS(&params.TargetPoint)
	if dstErr != nil {
		log.Error().Msgf("OSController error migration into objectstorage : %v", dstErr)
		return dstErr
	}

	log.Info().Msgf("Launch OSController Copy")
	if err := src.Copy(dst); err != nil {
		log.Error().Msgf("Copy error copying into objectstorage : %v", err)
		return err
	}
	log.Info().Msgf("successfully migrationed")
	return nil
}

func DeleteOSFunc(params *models.CommandTask) error {
	var OSC *osc.OSController
	var err error
	log.Info().Msgf("User Information")
	OSC, err = GetOS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("OSController error importing into objectstorage : %v", err)
		return err
	}

	log.Info().Msgf("Launch OSController Delete")
	if err := OSC.DeleteBucket(); err != nil {
		log.Error().Msgf("Delete error deleting into objectstorage : %v", err)
		return err
	}
	log.Info().Msgf("successfully deleted")

	return nil
}
